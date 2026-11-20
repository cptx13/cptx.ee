// Feeds: the cheatsheet RSS reader (shown on /addons/reads/?view=latest) — sources, fetching, cache and refresh limits (B-11).

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

// FeedItem represents a single article from an RSS feed.
type FeedItem struct {
	Title         string    `json:"title"`
	Link          string    `json:"link"`
	Source        string    `json:"source"`
	Date          time.Time `json:"date"`
	DateFormatted string    `json:"date_formatted"`
}

// FeedCache represents the on-disk cache of fetched feeds.
type FeedCache struct {
	LastRefreshed time.Time  `json:"last_refreshed"`
	Items         []FeedItem `json:"items"`
	Errors        []string   `json:"errors"`
}

const feedCacheFile = "feed-cache.json"

// Feed refresh limits (B-04): the server refreshes on its own every
// feedAutoInterval; the public refresh link is a no-op if a refresh is running
// or one finished less than feedManualGap ago, so crawlers can't hammer feeds.
const (
	feedAutoInterval = 6 * time.Hour
	feedManualGap    = 10 * time.Minute
)

var feedRefresh struct {
	sync.Mutex
	running bool
	last    time.Time
}

// refreshFeeds fetches all feeds unless a refresh is running or the last one
// finished less than minGap ago. It reports whether it fetched.
func refreshFeeds(minGap time.Duration, why string) bool {
	feedRefresh.Lock()
	if feedRefresh.running || time.Since(feedRefresh.last) < minGap {
		feedRefresh.Unlock()
		return false
	}
	feedRefresh.running = true
	feedRefresh.Unlock()
	defer func() {
		feedRefresh.Lock()
		feedRefresh.running = false
		feedRefresh.last = time.Now()
		feedRefresh.Unlock()
	}()

	log.Printf("Refreshing feeds (%s)...", why)
	cache := fetchAllFeeds()
	if err := saveFeedCache(cache); err != nil {
		log.Printf("error saving feed cache: %v", err)
	}
	log.Printf("Feed refresh complete: %d items, %d errors", len(cache.Items), len(cache.Errors))
	return true
}

// startFeedAutoRefresh refreshes the feed cache in the background whenever it
// is older than feedAutoInterval. FEEDS_AUTO_REFRESH=0 turns it off (tests).
func startFeedAutoRefresh() {
	if os.Getenv("FEEDS_AUTO_REFRESH") == "0" {
		return
	}
	if cache, err := loadFeedCache(); err == nil {
		feedRefresh.Lock()
		feedRefresh.last = cache.LastRefreshed
		feedRefresh.Unlock()
	}
	go func() {
		for {
			refreshFeeds(feedAutoInterval, "scheduled")
			time.Sleep(10 * time.Minute)
		}
	}()
}

// feedSources maps a display name to the RSS/Atom feed URL.
var feedSources = []struct {
	Name    string
	FeedURL string
}{
	{"No One's Happy", "https://nooneshappy.com/rss.xml"},
	{"Mrus", "https://xn--gckvb8fzb.com/index.xml"},
	{"Seth For Privacy", "https://sethforprivacy.com/index.xml"},
	{"Bartosz Ciechanowski", "https://ciechanow.ski/atom.xml"},
	{"Matthew Ball", "https://www.matthewball.co/matthewball?format=rss"},
	{"Matthew Green", "https://blog.cryptographyengineering.com/feed/"},
	{"Uncharted Territories", "https://unchartedterritories.tomaspueyo.com/feed"},
	{"Fitness Revolucionario", "https://www.fitnessrevolucionario.com/feed/"},
	{"Adam Thiede", "https://adamthiede.com/rss.xml"},
	{"Bunnie Studios", "https://www.bunniestudios.com/blog/?feed=rss2"},
	{"Privacy Guides", "https://www.privacyguides.org/articles/feed_rss_created.xml"},
	{"Import AI", "https://jack-clark.net/feed/"},
}

func loadFeedCache() (*FeedCache, error) {
	data, err := os.ReadFile(feedCacheFile)
	if err != nil {
		return nil, err
	}
	var cache FeedCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func saveFeedCache(cache *FeedCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(feedCacheFile, data, 0o644)
}

func fetchAllFeeds() *FeedCache {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type result struct {
		items []FeedItem
		err   string
	}

	results := make([]result, len(feedSources))
	var wg sync.WaitGroup

	for i, src := range feedSources {
		wg.Add(1)
		go func(idx int, name, feedURL string) {
			defer wg.Done()
			feedCtx, feedCancel := context.WithTimeout(ctx, 10*time.Second)
			defer feedCancel()

			parser := gofeed.NewParser()
			feed, err := parser.ParseURLWithContext(feedURL, feedCtx)
			if err != nil {
				results[idx] = result{err: fmt.Sprintf("%s: %v", name, err)}
				return
			}

			var items []FeedItem
			for _, item := range feed.Items {
				var pubDate time.Time
				if item.PublishedParsed != nil {
					pubDate = *item.PublishedParsed
				} else if item.UpdatedParsed != nil {
					pubDate = *item.UpdatedParsed
				}
				items = append(items, FeedItem{
					Title:         item.Title,
					Link:          item.Link,
					Source:        name,
					Date:          pubDate,
					DateFormatted: pubDate.Format("2006.01.02"),
				})
			}
			results[idx] = result{items: items}
		}(i, src.Name, src.FeedURL)
	}

	wg.Wait()

	var allItems []FeedItem
	var allErrors []string
	for _, r := range results {
		if r.err != "" {
			allErrors = append(allErrors, r.err)
		} else {
			allItems = append(allItems, r.items...)
		}
	}

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Date.After(allItems[j].Date)
	})

	return &FeedCache{
		LastRefreshed: time.Now(),
		Items:         allItems,
		Errors:        allErrors,
	}
}
