<template>
<AppLayout :pageTitle="post.Title + ' | cptx'" :description="post.Description" :canonicalURL="'https://cptx.ee' + post.Path" :lang="post.Lang" :ogType="'article'">
  <article>
    <header class="single-intro-container">
      <p v-if="post.Path == '/manifesto/'" class="addons-crumb"><a href="/about/">about</a> / manifesto</p>
      <p v-if="post.Section == 'legaldocs'" class="addons-crumb"><a href="/about/">about</a> / <a href="/legaldocs/">legal docs</a></p>
      <h1 class="single-title">{{ post.Title }}</h1>
      <div v-if="post.Section == 'pockets'" class="single-subsummary"><p class="single-date pk-sub"><a href="/pockets/">← all pockets</a> · updated {{ post.DateFormatted }}</p></div>
      <div v-if="post.Section == 'legaldocs'" class="single-subsummary"><p class="single-date pk-sub"><a href="/legaldocs/">← legal docs</a> · updated {{ post.DateFormatted }}</p></div>
      <div v-if="post.Section != 'pockets' && post.Section != 'legaldocs' && (post.Author || post.DateFormatted)" class="single-subsummary">
        <div>
          <p class="single-date">
            <span v-if="post.Author">{{ post.Author }} / </span>
            <time v-if="post.DateFormatted">{{ post.DateFormatted }}</time>
            <span v-if="post.ReadsLabel" class="reads-label"> · {{ post.ReadsLabel }}</span>
            <span v-if="post.ShowTags && post.Tags && post.Tags.length > 0">
              &nbsp;/&nbsp;
              <span v-for="(tag, i) in post.Tags"><span v-if="i > 0">, </span>#{{ tag }}</span>
            </span>
          </p>
        </div>
      </div>
    </header>
    <div :class="post.ContentClass" v-html="post.Content"></div>
    <div v-if="post.Readable" id="read-marker" :data-path="post.Path" :data-words="post.Words" aria-hidden="true"></div>
    <div v-if="post.Readable" class="post-end">
      <p><a class="share-link" :href="'https://cptx.ee' + post.Path" :data-share-url="'https://cptx.ee' + post.Path">share this post</a> · <a href="/index.xml">rss</a> · <a href="/newsletter/">newsletter</a></p>
    </div>
  </article>
  <Pagination :prevPost="prevPost" :nextPost="nextPost"></Pagination>
  <div class="back-to-top">
    <a href="#top">back to top</a>
  </div>
  <script src="/js/read.js" defer></script>
</AppLayout>
</template>
