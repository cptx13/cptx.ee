<template>
<AppLayout :pageTitle="'cptx'" :description="'a collection of mathoms'" :canonicalURL="'https://cptx.ee/'" :lang="'en'" :ogType="'website'">
  <div class="intro">
    <p>A collection of mathoms.</p>
  </div>
  <div v-if="hasOnThisDay" class="home-today">
    <p class="home-today-title">on this day <a href="/addons/today/">→</a></p>
    <article v-for="p in onThisDay" class="post-entry">
      <span class="post-entry-date">{{ p.Year }}</span>
      <h2 class="post-entry-title"><a :href="p.Path">{{ p.Title }}</a></h2>
    </article>
  </div>
  <div class="list-container">
    <div class="list-toolbar">
      <span class="list-toolbar-title">posts ({{ postCount }})</span>
      <button id="random-post-button" type="button" class="home-random" title="open a random post" hidden><svg aria-hidden="true" fill="currentColor" viewBox="0 0 24 24" width="15" height="15"><path d="M16 3h5v5h-2V6.41l-3.29 3.3-1.42-1.42L17.59 5H16V3Zm3 14h2v4h-5v-2h1.59l-3.3-3.29 1.42-1.42L19 17.59V17ZM3 7h3.59l8 8H21v2h-7.41l-8-8H3V7Zm0 10v-2h2.59l2-2 1.42 1.42L6.41 17H3Z"/></svg><span>random</span></button>
    </div>
    <form class="home-search" action="/search/" method="get" role="search">
      <svg class="home-search-icon" aria-hidden="true" fill="currentColor" viewBox="0 0 24 24" width="16" height="16"><path d="M15.5 14h-.79l-.28-.27A6.47 6.47 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg>
      <input id="post-search" type="search" name="q" placeholder="search posts…" aria-label="Search posts" autocomplete="off" enterkeyhint="search">
      <kbd class="home-search-key" aria-hidden="true">/</kbd>
    </form>
    <p id="search-full" class="home-search-full" hidden><a id="search-full-link" href="/search/">↵ search the full text of every post and the library</a></p>
    <div id="post-list">
      <PostEntry v-for="post in posts" :dateFormatted="post.DateShort" :path="post.Path" :title="post.Title"></PostEntry>
    </div>
    <p id="no-results" class="home-no-results" hidden>No titles match — press enter to search the full text.</p>
  </div>
</AppLayout>
</template>
