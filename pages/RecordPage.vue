<template>
<AppLayout :pageTitle="pageTitle" :description="description" :canonicalURL="canonical" :lang="'en'" :ogType="'music.album'">
  <div class="list-container addons library vinyl">
    <p class="addons-crumb"><a href="/archives/">archives</a> / <a href="/archives/vinyl/">vinyl</a> / <a :href="shelfHref">{{ shelfLabel }}</a></p>
    <div class="bk">
      <div class="bk-cover vn-big">
        <span class="vn-disc" :style="record.DiscStyle" aria-hidden="true"></span>
        <img v-if="record.HasCover" class="vn-sleeve" :src="large" :alt="record.Label" width="300" height="300">
        <span v-else class="vn-sleeve vn-gen" :style="record.GenStyle"><span class="vn-gen-artist">{{ artist }}</span><span class="vn-gen-title">{{ title }}</span><span class="vn-gen-year">{{ year }}</span></span>
      </div>
      <div class="bk-info">
        <h1 class="bk-title">{{ title }}</h1>
        <p class="bk-author"><span>by <strong>{{ artist }}</strong></span><span v-if="year" class="bk-year">{{ year }}</span></p>
        <dl class="bk-facts">
          <template v-for="f in facts"><dt>{{ f.Label }}</dt><dd>{{ f.Value }}</dd></template>
        </dl>
        <p class="vn-listen"><span class="vn-listen-label">listen</span><a v-for="l in listen" class="vn-pill" :href="l.Href" rel="noopener">{{ l.Label }}<span v-if="l.Search" class="vn-search"> · search</span></a></p>
        <p v-if="hasAbout" class="bk-links">more about it: <a v-for="l in about" :href="l.Href" rel="noopener">{{ l.Label }}</a></p>
      </div>
    </div>

    <div v-if="hasNeighbours" class="bk-neighbours">
      <h2 class="lib-shelf-label">next to it in the crate</h2>
      <div class="vn-crate">
        <a v-for="r in neighbours" class="vn-link" :href="r.Href" :title="r.Label">
          <span class="vn-disc" :style="r.DiscStyle" aria-hidden="true"></span>
          <img v-if="r.HasCover" class="vn-sleeve" :src="r.Thumb" :alt="r.Label" width="150" height="150" loading="lazy">
          <span v-else class="vn-sleeve vn-gen" :style="r.GenStyle"><span class="vn-gen-artist">{{ r.Artist }}</span><span class="vn-gen-title">{{ r.Title }}</span><span class="vn-gen-year">{{ r.Year }}</span></span>
        </a>
      </div>
    </div>
  </div>
</AppLayout>
</template>
