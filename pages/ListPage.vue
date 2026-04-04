<template>
<AppLayout :pageTitle="title + ' | cptx'" :description="description" :canonicalURL="canonicalURL">
  <div class="list-container" style="margin-top: 2.5em;">
    <div v-if="contentHTML" style="margin-bottom: 2em;" v-html="contentHTML"></div>
    <div v-if="posts.length === 0 && !contentHTML">
      <p>Work in progress, to be uploaded soon</p>
    </div>
    <PostEntry v-for="post in posts" :dateFormatted="post.DateShort" :path="post.Path" :title="post.Title"></PostEntry>
  </div>
</AppLayout>
</template>
