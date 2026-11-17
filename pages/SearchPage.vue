<template>
<AppLayout :pageTitle="'search | cptx'" :description="'Search every croquis, rima and recuerdo, and the library.'" :canonicalURL="'https://cptx.ee/search/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <h1 class="addons-title">search</h1>
    <form class="nl-form sr-form" action="/search/" method="get" role="search">
      <label for="sr-q">words to look for in every croquis, rima and recuerdo, and the library</label>
      <div class="nl-row">
        <input id="sr-q" type="search" name="q" :value="q" placeholder="café, Tallinn, lluvia…" autocomplete="off">
        <button type="submit">search</button>
      </div>
    </form>

    <p v-if="hasQuery" class="hm-stats">{{ total }} posts<span v-if="hasBooks"> · books in the library below</span> for “{{ q }}”<span v-if="more"> · showing the best 50</span></p>
    <p v-if="none" class="addons-intro">Nothing found. Try fewer words, or the start of a word (“llov” finds “llover” and “llovía”).</p>

    <div v-if="hasPosts">
      <article v-for="p in posts" class="sr-hit">
        <p class="sr-meta">{{ p.Day }} · {{ p.Section }}</p>
        <h2 class="post-entry-title sr-title"><a :href="p.Path">{{ p.Title }}</a></h2>
        <p class="sr-snippet" v-html="p.Snippet"></p>
      </article>
    </div>

    <div v-if="hasBooks">
      <h2 class="addons-sub">in the library</h2>
      <div class="pl-list rd-pages">
        <a v-for="b in books" class="pl-row" :href="b.Path"><span>{{ b.Title }}</span><span class="pl-count">{{ b.Author }}</span></a>
      </div>
    </div>
  </div>
</AppLayout>
</template>
