<template>
<AppLayout :pageTitle="'reads | cptx'" :description="'Articles from other people I have enjoyed, and the latest from the blogs I follow.'" :canonicalURL="'https://cptx.ee/addons/reads/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/addons/">addons</a> / reads</p>
    <h1 class="addons-title">reads</h1>
    <p class="addons-intro">Articles from other people I have enjoyed, and the latest from the {{ sourceCount }} blogs I follow in the <a href="/pockets/cheatsheet/">cheatsheet</a>.</p>

    <div class="hm-filters">
      <p><span class="hm-filter-label">show</span><a href="/addons/reads/" :class="{ 'hm-opt': true, 'hm-active': isKept }">kept ★ ({{ keptCount }})</a><a href="/addons/reads/?view=latest" :class="{ 'hm-opt': true, 'hm-active': isLatest }" rel="nofollow">latest from the blogs</a></p>
    </div>

    <div v-if="isKept">
      <div v-if="hasKept">
        <section v-for="m in months" class="kp-month">
          <h2 class="kp-month-label">{{ m.Label }}</h2>
          <div v-for="r in m.Items" class="kp-item">
            <a :href="r.URL" class="kp-title" target="_blank" rel="noopener">{{ r.Title }}</a>
            <span class="kp-meta"><span v-if="r.HasAuthor">{{ r.Author }} · </span>{{ r.Host }}</span>
            <span v-if="r.HasNote" class="kp-note">{{ r.Note }}</span>
          </div>
        </section>
      </div>
      <p v-else class="addons-intro">Nothing kept yet.</p>
    </div>

    <div v-if="isLatest">
      <p class="hm-stats">★ = also kept<span v-if="hasRefreshed"> · refreshed {{ lastRefreshed }}</span> · <a href="/addons/reads/refresh" rel="nofollow">↻ refresh</a></p>
      <div v-if="hasLatest">
        <div v-for="it in latest" class="kp-item">
          <a :href="it.Link" class="kp-title" target="_blank" rel="noopener"><span v-if="it.Kept">★ </span>{{ it.Title }}</a>
          <span class="kp-meta">{{ it.Source }} · {{ it.Date }}</span>
        </div>
      </div>
      <p v-else class="addons-intro">No articles fetched yet. Try <a href="/addons/reads/refresh" rel="nofollow">↻ refresh</a>.</p>
      <details v-if="hasErrors" class="kp-errors">
        <summary>{{ errorCount }} blog(s) could not be fetched</summary>
        <ul><li v-for="e in errors">{{ e }}</li></ul>
      </details>
    </div>
  </div>
</AppLayout>
</template>
