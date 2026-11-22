package main

import (
	"html/template"
	"strings"
	"testing"
)

func TestLinkBedside(t *testing.T) {
	kc := &libBook{Key: "kc", Title: "Kitchen Confidential", Author: "Anthony Bourdain", HasCover: true, ThumbW: 40, ThumbH: 60}
	eye := &libBook{Key: "eye", Title: "Eye of the Bedlam Bride", Author: "Matt Dinniman"}
	wm := &libBook{Key: "wm", Title: "Watchmen: International Edition", Author: "Alan Moore & Dave Gibbons"}
	other := &libBook{Key: "o", Title: "Butter", Author: "Someone Else"}
	lib := &library{Books: []*libBook{kc, eye, wm, other}}
	p := &Post{Content: template.HTML(`<h4 id="x">2026.09</h4>
<ul>
<li><em>Kitchen Confidential</em> (2000) by Anthony Bourdain</li>
<li><em>Butter</em> (2017) by Asako Yuzuki</li>
</ul>
<h4>2026.08</h4>
<ul>
<li><em>Watchmen: The Deluxe Edition</em> (2013) by Alan Moore</li>
<li><em>The Eye of the Bedlam Bride</em> (2023) by Matt Dinniman</li>
</ul>`)}
	linkBedside(p, lib)
	got := string(p.Content)
	for _, want := range []string{
		`<li class="bs-book"><a class="bs-cover" href="/archives/library/kc/"`,
		`<img src="/archives/library/covers/kc.jpg" alt="" width="40" height="60"`,
		`<a href="/archives/library/kc/"><em>Kitchen Confidential</em></a> (2000) by Anthony Bourdain`,
		`<li><em>Butter</em> (2017) by Asako Yuzuki</li>`,                               // different author: left alone
		`<li><a href="/archives/library/wm/"><em>Watchmen: The Deluxe Edition</em></a>`, // no cover: plain line, linked
		`<a href="/archives/library/eye/"><em>The Eye of the Bedlam Bride</em></a>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Join(kc.Bedside, ",") != "2026.09" || strings.Join(eye.Bedside, ",") != "2026.08" || len(other.Bedside) != 0 {
		t.Errorf("months: kc %v eye %v other %v", kc.Bedside, eye.Bedside, other.Bedside)
	}
}
