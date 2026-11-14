<template>
<AppLayout :pageTitle="pageTitle" :description="description" :canonicalURL="canonical" :lang="'en'" :ogType="'book'">
  <div class="list-container addons library">
    <p class="addons-crumb"><a href="/archives/">archives</a> / <a href="/archives/library/">library</a> / <a :href="shelfHref">{{ shelfLabel }}</a></p>
    <div class="bk">
      <div class="bk-cover">
        <img v-if="book.HasCover" class="lib-cover bk-img" :src="large" :alt="book.Label" :width="coverW" :height="coverH">
        <span v-else class="lib-gen bk-gen" :style="book.GenStyle"><span class="lib-gen-title">{{ title }}</span><span class="lib-gen-author">{{ author }}</span></span>
      </div>
      <div class="bk-info">
        <h1 class="bk-title" :lang="lang">{{ title }}</h1>
        <p v-if="subtitle" class="bk-subtitle" :lang="lang">{{ subtitle }}</p>
        <p class="bk-author">{{ author }}<span v-if="year"> · {{ year }}</span></p>
        <p v-if="series" class="bk-series">{{ series }}</p>
        <dl class="bk-facts">
          <template v-for="f in facts"><dt>{{ f.Label }}</dt><dd>{{ f.Value }}</dd></template>
        </dl>
        <p v-if="own" class="bk-buy"><a :href="buy" rel="noopener">I wrote this one · buy it →</a></p>
        <p v-if="hasSubjects" class="bk-subjects"><span v-for="s in subjects" class="bk-subject">{{ s }}</span></p>
        <p v-if="hasLinks" class="bk-links">more about it: <a v-for="l in links" :href="l.Href" rel="noopener">{{ l.Label }}</a></p>
      </div>
    </div>

    <div v-if="hasMentions" class="bk-mentions">
      <h2 class="lib-shelf-label">in the blog</h2>
      <article v-for="p in mentions" class="post-entry">
        <span class="post-entry-date">{{ p.Day }}</span>
        <h2 class="post-entry-title"><a :href="p.Path">{{ p.Title }}</a></h2>
      </article>
    </div>

    <div v-if="hasNeighbours" class="bk-neighbours">
      <h2 class="lib-shelf-label">next to it on the shelf</h2>
      <div class="lib-shelf lib-shelf-covers">
        <a v-for="b in neighbours" class="lib-cover-link" :href="b.Href" :title="b.Label">
          <img v-if="b.HasCover" class="lib-cover" :src="b.Thumb" :alt="b.Label" height="150" :width="b.CoverW" loading="lazy">
          <span v-else class="lib-gen" :style="b.GenStyle"><span class="lib-gen-title">{{ b.Title }}</span><span class="lib-gen-author">{{ b.Author }}</span></span>
        </a>
      </div>
    </div>
  </div>
</AppLayout>
</template>
