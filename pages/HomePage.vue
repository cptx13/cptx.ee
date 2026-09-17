<template>
<AppLayout :pageTitle="'cptx'" :description="'a collection of mathoms'" :canonicalURL="'https://cptx.ee/'">
  <div class="intro">
    <p>A collection of mathoms.</p>
  </div>
  <div class="list-container">
    <div class="list-toolbar">
      <span class="list-toolbar-title">posts ({{ postCount }})</span>
      <a id="random-post-button" href="#" title="shuffle" aria-label="Shuffle post" class="toolbar-icon shuffle-button"><svg fill="currentColor" viewBox="0 0 24 24" width="16" height="16"><path d="M16 3h5v5h-2V6.41l-3.29 3.3-1.42-1.42L17.59 5H16V3Zm3 14h2v4h-5v-2h1.59l-3.3-3.29 1.42-1.42L19 17.59V17ZM3 7h3.59l8 8H21v2h-7.41l-8-8H3V7Zm0 10v-2h2.59l2-2 1.42 1.42L6.41 17H3Z"/></svg></a>
      <button id="search-toggle" title="search" aria-label="Search posts" class="toolbar-icon"><svg fill="currentColor" viewBox="0 0 24 24" width="16" height="16"><path d="M15.5 14h-.79l-.28-.27A6.47 6.47 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg></button>
    </div>
    <div class="search-bar" id="search-bar" style="display:none;"><input id="post-search" type="text" placeholder="search..." autocomplete="off"></div>
    <div id="post-list">
      <PostEntry v-for="post in posts" :dateFormatted="post.DateShort" :path="post.Path" :title="post.Title"></PostEntry>
    </div>
    <p id="no-results" style="display:none; color: var(--content-secondary); font-size: 0.9em;">No posts found.</p>
  </div>
</AppLayout>
</template>
