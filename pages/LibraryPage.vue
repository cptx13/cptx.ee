<template>
<AppLayout :pageTitle="pageTitle" :description="'Some of the books Merzeede and I own, by cover or spine: shelve them by genre, colour, author, decade, language or size.'" :canonicalURL="canonical" :lang="'en'" :ogType="'website'">
  <div class="list-container addons library">
    <p class="addons-crumb"><a href="/archives/">archives</a> / library<span v-if="hasFocus"> / {{ focus }}</span></p>
    <h1 class="addons-title">library</h1>
    <p class="addons-intro">Some of the books Merzeede and I own. Not all of them have been scanned yet, so the shelves keep growing.</p>
    <p class="hm-stats">{{ stats }}</p>

    <div class="hm-filters">
      <p><span class="hm-filter-label">shelve by</span><a v-for="o in groupings" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
      <p><span class="hm-filter-label">show</span><a v-for="o in views" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a><a v-if="hasFocus" class="hm-opt" :href="allHref" rel="nofollow">all shelves</a></p>
    </div>

    <p v-if="empty" class="addons-intro">The shelves are still empty.</p>

    <section v-for="s in shelves" class="lib-section">
      <h2 class="lib-shelf-label"><a :href="s.Href" rel="nofollow">{{ s.Label }}</a> <span class="lib-shelf-count">{{ s.Count }}</span></h2>
      <div v-if="spines" class="lib-shelf lib-shelf-spines">
        <a v-for="b in s.Items" class="lib-slot" :href="b.Href" :title="b.Label"><span class="lib-spine" :style="b.SpineStyle" :lang="b.Lang"><span class="lib-spine-title">{{ b.Main }}</span><span v-if="b.ShowAuthor" class="lib-spine-author">{{ b.Surname }}</span></span></a>
      </div>
      <div v-else class="lib-shelf lib-shelf-covers">
        <a v-for="b in s.Items" class="lib-cover-link" :href="b.Href" :title="b.Label">
          <img v-if="b.HasCover" class="lib-cover" :src="b.Thumb" :alt="b.Label" height="150" :width="b.CoverW" loading="lazy">
          <span v-else class="lib-gen" :style="b.GenStyle" :lang="b.Lang"><span class="lib-gen-author">{{ b.Author }}</span><span class="lib-gen-mid"><span class="lib-gen-title">{{ b.Main }}</span><span v-if="b.Sub" class="lib-gen-sub">{{ b.Sub }}</span></span><span class="lib-gen-year">{{ b.Year }}</span></span>
        </a>
      </div>
    </section>

    <p class="addon-share"><a class="share-link" :href="shareURL" :data-share-url="shareURL">share this view</a></p>
  </div>
</AppLayout>
</template>
