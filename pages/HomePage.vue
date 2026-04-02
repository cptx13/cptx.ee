<template>
<AppLayout :pageTitle="'cptx'" :description="'a collection of mathoms'" :canonicalURL="'https://cptx.ee/'">
  <div class="intro">
    <p>A collection of mathoms. Try the shuffle button!</p>
    <SocialIcons></SocialIcons>
  </div>
  <div class="list-container">
    <PostEntry v-for="post in posts" :dateFormatted="post.DateFormatted" :path="post.Path" :title="post.Title"></PostEntry>
  </div>
</AppLayout>
</template>
