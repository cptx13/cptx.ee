<template>
<AppLayout :pageTitle="'words | cptx'" :description="'The most used words across the archive, without articles or pronouns.'" :canonicalURL="'https://cptx.ee/addons/words/'" :lang="'es'" :ogType="'website'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/addons/">addons</a> / words</p>
    <h1 class="addons-title">words</h1>
    <p class="hm-stats">top 120 of {{ distinct }} different words · {{ yearLabel }} · articles, pronouns, prepositions and other small words left out</p>

    <div class="hm-filters">
      <p><span class="hm-filter-label">year</span><a v-for="o in years" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
      <p><span class="hm-filter-label">section</span><a v-for="o in sections" :href="o.Href" :class="{ 'hm-opt': true, 'hm-active': o.Active }" rel="nofollow">{{ o.Label }}</a></p>
    </div>

    <div class="wd-grid">
      <a v-for="t in tiles" :class="t.Class" :href="t.Href" :title="t.Label" rel="nofollow"><span class="wd-word">{{ t.Word }}</span><span class="wd-count">{{ t.Count }}</span></a>
    </div>

    <div v-if="hasSelected" id="word" class="hm-day">
      <h2>“{{ selected }}” · {{ selCount }}</h2>
      <div class="wd-bars">
        <p v-for="b in bars" class="wd-bar"><span class="wd-bar-year">{{ b.Year }}</span><span class="wd-bar-track"><span class="wd-bar-fill" :style="b.Width"></span></span><span class="wd-bar-n">{{ b.Count }}</span></p>
      </div>
      <article v-for="p in selPosts" class="post-entry">
        <span class="post-entry-date">{{ p.Day }}</span>
        <h2 class="post-entry-title"><a :href="p.Path">{{ p.Title }}</a></h2>
      </article>
    </div>

    <p class="addon-share"><a class="share-link" :href="shareURL" :data-share-url="shareURL">share this view</a></p>
  </div>
</AppLayout>
</template>
