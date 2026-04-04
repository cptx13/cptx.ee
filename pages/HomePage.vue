<template>
<AppLayout :pageTitle="'cptx'" :description="'a collection of mathoms'" :canonicalURL="'https://cptx.ee/'">
  <div class="intro">
    <p>A collection of mathoms. Try the shuffle button!</p>
    <SocialIcons></SocialIcons>
  </div>
  <div class="list-container">
    <h2 style="font-size: 1.1em; margin-bottom: 1em; font-family: inherit;">posts ({{ postCount }})</h2>
    <PostEntry v-for="post in posts" :dateFormatted="post.DateYear" :path="post.Path" :title="post.Title"></PostEntry>
  </div>
</AppLayout>
</template>
