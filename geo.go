// Country lookup for the readers addon (A-03): DB-IP "IP to Country Lite"
// (CC BY 4.0, https://db-ip.com), a CSV of "first,last,country" ranges kept in
// <data>/geo/dbip-country-lite.csv.gz and refreshed monthly by the server.
// Only the country code is ever kept; the IP address is dropped after the lookup.

package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type geoRange4 struct {
	lo, hi uint32
	cc     uint16
}

type geoRange6 struct {
	lo, hi [16]byte
	cc     uint16
}

type geoDB struct {
	v4    []geoRange4
	v6    []geoRange6
	codes []string // index → ISO code
}

// lookup returns the ISO country code for addr, or "" when unknown.
func (g *geoDB) lookup(addr netip.Addr) string {
	if g == nil || !addr.IsValid() {
		return ""
	}
	addr = addr.Unmap()
	if addr.Is4() {
		b := addr.As4()
		x := binary.BigEndian.Uint32(b[:])
		i := sort.Search(len(g.v4), func(i int) bool { return g.v4[i].hi >= x })
		if i < len(g.v4) && g.v4[i].lo <= x {
			return g.codes[g.v4[i].cc]
		}
		return ""
	}
	x := addr.As16()
	i := sort.Search(len(g.v6), func(i int) bool { return string(g.v6[i].hi[:]) >= string(x[:]) })
	if i < len(g.v6) && string(g.v6[i].lo[:]) <= string(x[:]) {
		return g.codes[g.v6[i].cc]
	}
	return ""
}

// parseGeoCSV reads "first,last,cc" lines (gzip or plain). Adjacent ranges with
// the same country are merged to keep memory small.
func parseGeoCSV(r io.Reader) (*geoDB, error) {
	g := &geoDB{}
	idx := map[string]uint16{}
	code := func(cc string) uint16 {
		if i, ok := idx[cc]; ok {
			return i
		}
		idx[cc] = uint16(len(g.codes))
		g.codes = append(g.codes, cc)
		return idx[cc]
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		f := strings.Split(strings.TrimSpace(sc.Text()), ",")
		if len(f) < 3 {
			continue
		}
		lo, err1 := netip.ParseAddr(strings.Trim(f[0], `"`))
		hi, err2 := netip.ParseAddr(strings.Trim(f[1], `"`))
		cc := strings.ToUpper(strings.Trim(f[2], `" `))
		if err1 != nil || err2 != nil || len(cc) != 2 {
			continue
		}
		c := code(cc)
		if lo.Is4() && hi.Is4() {
			a, b := lo.As4(), hi.As4()
			r := geoRange4{binary.BigEndian.Uint32(a[:]), binary.BigEndian.Uint32(b[:]), c}
			if n := len(g.v4); n > 0 && g.v4[n-1].cc == c && g.v4[n-1].hi+1 == r.lo {
				g.v4[n-1].hi = r.hi
				continue
			}
			g.v4 = append(g.v4, r)
		} else if lo.Is6() && hi.Is6() {
			g.v6 = append(g.v6, geoRange6{lo.As16(), hi.As16(), c})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.Slice(g.v4, func(i, j int) bool { return g.v4[i].lo < g.v4[j].lo })
	sort.Slice(g.v6, func(i, j int) bool { return string(g.v6[i].lo[:]) < string(g.v6[j].lo[:]) })
	if len(g.v4) == 0 && len(g.v6) == 0 {
		return nil, fmt.Errorf("no ranges")
	}
	return g, nil
}

func loadGeoFile(path string) (*geoDB, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		r = gz
	}
	return parseGeoCSV(r)
}

// geoKeeper holds the current database and refreshes it monthly.
type geoKeeper struct {
	mu   sync.RWMutex
	db   *geoDB
	path string
}

func (k *geoKeeper) lookup(addr netip.Addr) string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.db.lookup(addr)
}

func (k *geoKeeper) load() {
	db, err := loadGeoFile(k.path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("geo: %v", err)
		}
		return
	}
	k.mu.Lock()
	k.db = db
	k.mu.Unlock()
	log.Printf("geo: %d IPv4 + %d IPv6 ranges", len(db.v4), len(db.v6))
}

// refreshLoop downloads this month's file when the copy on disk is missing or
// older than 32 days (DB-IP publishes on the 1st), then reloads it.
func (k *geoKeeper) refreshLoop() {
	for {
		if st, err := os.Stat(k.path); err != nil || time.Since(st.ModTime()) > 32*24*time.Hour {
			if err := k.download(); err != nil {
				log.Printf("geo update: %v", err)
			} else {
				k.load()
			}
		}
		time.Sleep(24 * time.Hour)
	}
}

func (k *geoKeeper) download() error {
	now := time.Now().UTC()
	var lastErr error
	for _, t := range []time.Time{now, now.AddDate(0, -1, 0)} {
		url := "https://download.db-ip.com/free/dbip-country-lite-" + t.Format("2006-01") + ".csv.gz"
		client := &http.Client{Timeout: 2 * time.Minute}
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("%s: %s", url, resp.Status)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(k.path), 0o755); err != nil {
			resp.Body.Close()
			return err
		}
		tmp := strings.TrimSuffix(k.path, ".gz") + ".part.gz" // keeps .gz so the check below unzips it
		out, err := os.Create(tmp)
		if err != nil {
			resp.Body.Close()
			return err
		}
		_, err = io.Copy(out, io.LimitReader(resp.Body, 200<<20))
		resp.Body.Close()
		out.Close()
		if err != nil {
			return err
		}
		if _, err := loadGeoFile(tmp); err != nil { // only swap in a file that parses
			os.Remove(tmp)
			return fmt.Errorf("downloaded file: %w", err)
		}
		return os.Rename(tmp, k.path)
	}
	return lastErr
}
