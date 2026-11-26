<template>
<AppLayout :pageTitle="pageTitle" :description="description" :canonicalURL="canonical" :lang="'en'" :ogType="'video.movie'">
  <div class="list-container addons library films">
    <p class="addons-crumb"><a href="/archives/">archives</a> / <a href="/archives/films/">films</a> / <a :href="shelfHref">{{ shelfLabel }}</a></p>
    <div class="bk">
      <div class="bk-cover fm-big">
        <img v-if="film.HasCover" class="fm-poster" :src="large" :alt="film.Label" width="240" height="360">
        <span v-else class="fm-poster fm-gen" :style="film.GenStyle"><span class="fm-gen-dir">{{ director }}</span><span class="fm-gen-title">{{ title }}</span><span class="fm-gen-year">{{ year }}</span></span>
        <span v-if="posterPhoto" class="fm-photo-note">photo of our copy</span>
      </div>
      <div class="bk-info">
        <h1 class="bk-title">{{ title }}</h1>
        <p class="bk-author"><span v-if="hasDirector">by <strong>{{ director }}</strong></span><span v-if="year" class="bk-year">{{ year }}</span></p>
        <dl class="bk-facts">
          <template v-for="f in facts"><dt>{{ f.Label }}</dt><dd>{{ f.Value }}</dd></template>
        </dl>
        <p class="vn-listen"><span class="vn-listen-label">watch</span><a v-for="l in watch" class="vn-pill" :href="l.Href" rel="noopener">{{ l.Label }}<span v-if="l.Search" class="vn-search"> · search</span></a></p>
        <p class="bk-links">more about it: <a v-for="l in about" :href="l.Href" rel="noopener">{{ l.Label }}</a></p>
      </div>
    </div>

    <div v-if="hasNeighbours" class="bk-neighbours">
      <h2 class="lib-shelf-label">next to it on the shelf</h2>
      <div class="fm-shelf">
        <a v-for="f in neighbours" class="fm-link" :href="f.Href" :title="f.Label">
          <img v-if="f.HasCover" class="fm-poster" :src="f.Thumb" :alt="f.Label" width="120" height="180" loading="lazy">
          <span v-else class="fm-poster fm-gen" :style="f.GenStyle"><span class="fm-gen-dir">{{ f.Director }}</span><span class="fm-gen-title">{{ f.Title }}</span><span class="fm-gen-year">{{ f.Year }}</span></span>
        </a>
      </div>
    </div>
  </div>
</AppLayout>
</template>
