<template>
<AppLayout :pageTitle="'cptx'" :description="'a collection of mathoms'" :canonicalURL="'https://cptx.ee/'">
  <div class="intro">
    <p>A collection of mathoms. Try the shuffle button!</p>
    <SocialIcons></SocialIcons>
  </div>
  <div class="list-container">
    <h2 style="font-size: 1.1em; margin-bottom: 1em; font-family: inherit;">posts ({{ postCount }})</h2>
    <div class="search-bar"><input id="post-search" type="text" placeholder="search posts..." autocomplete="off"></div>
    <div id="post-list">
      <PostEntry v-for="post in posts" :dateFormatted="post.DateShort" :path="post.Path" :title="post.Title"></PostEntry>
    </div>
    <p id="no-results" style="display:none; color: var(--content-secondary); font-size: 0.9em;">No posts found.</p>
  </div>
</AppLayout>
</template>
