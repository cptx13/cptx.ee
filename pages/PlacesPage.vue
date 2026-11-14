<template>
<AppLayout :pageTitle="'city lights | cptx'" :description="'A pixel map of the world: every city mentioned in a post lights up.'" :canonicalURL="'https://cptx.ee/addons/places/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/addons/">addons</a> / city lights</p>
    <h1 class="addons-title">city lights</h1>
    <p class="hm-stats">{{ placeCount }} places · {{ mentions }} mentions · {{ yearLabel }}</p>

    <div class="hm-filters">
      <p><span class="hm-filter-label">map</span><a v-for="o in regions" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
      <p><span class="hm-filter-label">year</span><a v-for="o in years" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
      <p><span class="hm-filter-label">section</span><a v-for="o in sections" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
    </div>

    <div class="pl-wrap" v-html="svg"></div>
    <p class="hm-legend">brighter and bigger = mentioned more often · tap a light</p>

    <div v-if="hasSelected" id="place" class="hm-day">
      <h2>{{ selected }} · {{ selPostCount }} posts</h2>
      <article v-for="p in selPosts" class="post-entry">
        <span class="post-entry-date">{{ p.Day }}</span>
        <h2 class="post-entry-title"><a :href="p.Path">{{ p.Title }}</a></h2>
      </article>
      <p v-if="hasMore" class="addons-intro">…and {{ selMore }} more (newest shown).</p>
    </div>

    <h2 class="addons-sub">most mentioned</h2>
    <div class="pl-list">
      <a v-for="r in rows" :href="r.Href" :class="{ 'pl-row': true, 'hm-active': r.Selected }" rel="nofollow"><span>{{ r.Name }}</span><span class="pl-count">{{ r.Mentions }}</span></a>
    </div>

    <p class="addon-share"><a class="share-link" :href="shareURL" :data-share-url="shareURL">share this view</a></p>
    <p class="addon-credit">Places from GeoNames (CC BY 4.0) via all-the-cities; land from Natural Earth. Found automatically in the text, so a few may be wrong.</p>
  </div>
</AppLayout>
</template>
