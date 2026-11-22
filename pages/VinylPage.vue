<template>
<AppLayout :pageTitle="pageTitle" :description="'Some of the records Merzeede and I own, by cover: sort them by genre, artist, decade or colour, and listen on Spotify or YouTube Music.'" :canonicalURL="canonical" :lang="'en'" :ogType="'website'">
  <div class="list-container addons library vinyl">
    <p class="addons-crumb"><a href="/archives/">archives</a> / vinyl<span v-if="hasFocus"> / {{ focus }}</span></p>
    <h1 class="addons-title">vinyl</h1>
    <p class="addons-intro">Some of the records Merzeede and I own. Not all of them have been scanned yet, so the crates keep growing. Open one to listen to it or read about it.</p>
    <p class="hm-stats">{{ stats }}</p>

    <div class="hm-filters">
      <p><span class="hm-filter-label">sort by</span><a v-for="o in groupings" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a><a v-if="hasFocus" class="hm-opt" :href="allHref" rel="nofollow">all crates</a></p>
    </div>

    <p v-if="empty" class="addons-intro">The crates are still empty.</p>

    <section v-for="s in shelves" class="lib-section">
      <h2 class="lib-shelf-label"><a :href="s.Href" rel="nofollow">{{ s.Label }}</a> <span class="lib-shelf-count">{{ s.Count }}</span></h2>
      <div class="vn-crate">
        <a v-for="r in s.Items" class="vn-link" :href="r.Href" :title="r.Label">
          <span class="vn-disc" :style="r.DiscStyle" aria-hidden="true"></span>
          <img v-if="r.HasCover" class="vn-sleeve" :src="r.Thumb" :alt="r.Label" width="150" height="150" loading="lazy">
          <span v-else class="vn-sleeve vn-gen" :style="r.GenStyle"><span class="vn-gen-artist">{{ r.Artist }}</span><span class="vn-gen-title">{{ r.Title }}</span><span class="vn-gen-year">{{ r.Year }}</span></span>
        </a>
      </div>
    </section>

    <p class="addon-share"><a class="share-link" :href="shareURL" :data-share-url="shareURL">share this view</a></p>
  </div>
</AppLayout>
</template>
