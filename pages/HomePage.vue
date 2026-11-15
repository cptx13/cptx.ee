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
    <h2 class="home-sub">latest <a id="random-post-button" href="#" title="shuffle" aria-label="Open a random post" class="toolbar-icon shuffle-button"><svg fill="currentColor" viewBox="0 0 24 24" width="16" height="16"><path d="M16 3h5v5h-2V6.41l-3.29 3.3-1.42-1.42L17.59 5H16V3Zm3 14h2v4h-5v-2h1.59l-3.3-3.29 1.42-1.42L19 17.59V17ZM3 7h3.59l8 8H21v2h-7.41l-8-8H3V7Zm0 10v-2h2.59l2-2 1.42 1.42L6.41 17H3Z"/></svg></a></h2>
    <PostEntry v-for="post in latest" :dateFormatted="post.DateShort" :path="post.Path" :title="post.Title"></PostEntry>
    <p class="home-all"><a href="/all/">all {{ postCount }} posts, with search →</a></p>

    <h2 class="home-sub">sections</h2>
    <div class="home-sections">
      <a href="/categories/croquis/">croquis<span>{{ croquisCount }} short pieces and poems</span></a>
      <a href="/categories/recuerdos/">recuerdos<span>{{ recuerdosCount }} daily memories, 2021–2024</span></a>
      <a href="/photos/">photos<span>{{ photoPlaces }} places</span></a>
      <a href="/addons/">addons<span>maps, words, readers</span></a>
      <a href="/pockets/">pockets<span>books, links, reads</span></a>
      <a href="/archives/">archives<span>my books, our library</span></a>
    </div>
  </div>
</AppLayout>
</template>
