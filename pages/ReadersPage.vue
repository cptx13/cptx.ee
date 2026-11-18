<template>
<AppLayout :pageTitle="'readers | cptx'" :description="'Who comes by, counted without cookies: visits by month, the most read posts and a map of where readers are.'" :canonicalURL="'https://cptx.ee/addons/readers/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/addons/">addons</a> / readers</p>
    <h1 class="addons-title">readers</h1>
    <p class="hm-stats">{{ visits }} visits · {{ visitors }} visitors · {{ reads }} reads · {{ countryCount }} countries<span v-if="hasSince"> · since {{ since }}</span></p>
    <p class="addons-intro">Counted on this server, without cookies or trackers, and only for people: a page counts once someone scrolls, taps or stays on it, so scrapers that just copy pages are left out. A visit is a page opened by a person (once a day per page); a read is someone reaching the end of a post after enough time to have read it. Visitors are counted per day, so someone who comes back tomorrow counts again. <a href="/legaldocs/privacy/#counting-readers">How it works →</a></p>

    <h2 class="addons-sub">by month</h2>
    <div class="wd-bars rd-months">
      <p v-for="m in months" class="wd-bar"><span class="wd-bar-year rd-month">{{ m.Label }}</span><span class="wd-bar-track"><span class="wd-bar-fill" :style="m.Width"></span></span><span class="wd-bar-n rd-n">{{ m.Visits }}</span></p>
    </div>
    <p class="hm-legend">page visits per month · {{ monthsNote }}</p>

    <h2 class="addons-sub">where readers are</h2>
    <div class="pl-wrap" v-html="svg"></div>
    <div v-if="hasCountries" class="pl-list">
      <p v-for="c in countries" class="pl-row"><span>{{ c.Name }}</span><span class="pl-count">{{ c.N }}</span></p>
    </div>
    <p v-if="hasMoreCountries" class="hm-legend">top 10 · {{ moreCountries }} more on the map</p>
    <p v-else class="addons-intro">No readers counted yet.</p>

    <h2 class="addons-sub">most read</h2>
    <div v-if="hasRead">
      <article v-for="p in mostRead" class="post-entry">
        <span class="post-entry-date">{{ p.Day }}</span>
        <h2 class="post-entry-title"><a :href="p.Path">{{ p.Title }}</a><span class="pl-count rd-reads">{{ p.Reads }}</span></h2>
      </article>
    </div>
    <p v-else class="addons-intro">Nothing read to the end yet.</p>

    <h2 class="addons-sub">most visited pages</h2>
    <div v-if="hasVisited" class="pl-list rd-pages">
      <a v-for="p in mostVisited" class="pl-row" :href="p.Path"><span>{{ p.Label }}</span><span class="pl-count">{{ p.N }}</span></a>
    </div>
    <p v-else class="addons-intro">No visits counted yet.</p>

    <p class="addon-credit">Countries: <a href="https://db-ip.com">IP Geolocation by DB-IP</a> (CC BY 4.0). Each country's light sits at the middle of its cities in GeoNames. Bots, link previews and scrapers are not counted.</p>
  </div>
</AppLayout>
</template>
