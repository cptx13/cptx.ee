<template>
<AppLayout :pageTitle="pageTitle" :description="'Some of the films Merzeede and I own, by poster: sort them by genre, title, director, decade or colour, and find out where to watch each one.'" :canonicalURL="canonical" :lang="'en'" :ogType="'website'">
  <div class="list-container addons library films">
    <p class="addons-crumb"><a href="/archives/">archives</a> / films<span v-if="hasFocus"> / {{ focus }}</span></p>
    <h1 class="addons-title">films</h1>
    <p class="addons-intro">Some of the films Merzeede and I own, on DVD and Blu-ray. Open one to read about it or find where to watch it.</p>
    <p class="hm-stats">{{ stats }}</p>

    <div class="hm-filters">
      <p><span class="hm-filter-label">sort by</span><a v-for="o in groupings" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a><a v-if="hasFocus" class="hm-opt" :href="allHref" rel="nofollow">all shelves</a></p>
    </div>

    <p v-if="empty" class="addons-intro">The shelf is still empty.</p>

    <section v-for="s in shelves" class="lib-section">
      <h2 class="lib-shelf-label"><a :href="s.Href" rel="nofollow">{{ s.Label }}</a> <span class="lib-shelf-count">{{ s.Count }}</span></h2>
      <div class="fm-shelf">
        <a v-for="f in s.Items" class="fm-link" :href="f.Href" :title="f.Label">
          <img v-if="f.HasCover" class="fm-poster" :src="f.Thumb" :alt="f.Label" width="120" height="180" loading="lazy">
          <span v-else class="fm-poster fm-gen" :style="f.GenStyle"><span class="fm-gen-dir">{{ f.Director }}</span><span class="fm-gen-title">{{ f.Title }}</span><span class="fm-gen-year">{{ f.Year }}</span></span>
        </a>
      </div>
    </section>

    <p class="addon-share"><a class="share-link" :href="shareURL" :data-share-url="shareURL">share this view</a></p>
  </div>
</AppLayout>
</template>
