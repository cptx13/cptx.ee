<template>
<AppLayout :pageTitle="post.Title + ' | cptx'" :description="post.Summary" :canonicalURL="'https://cptx.ee' + post.Path">
  <article>
    <header class="single-intro-container">
      <h1 class="single-title">{{ post.Title }}</h1>
      <div v-if="post.Author || post.DateFormatted" class="single-subsummary">
        <div>
          <p class="single-date">
            <span v-if="post.Author">{{ post.Author }} / </span>
            <time v-if="post.DateFormatted">{{ post.DateFormatted }}</time>
            <span v-if="post.ShowTags && post.Tags && post.Tags.length > 0">
              &nbsp;/&nbsp;
              <span v-for="(tag, i) in post.Tags"><span v-if="i > 0">, </span>#{{ tag }}</span>
            </span>
          </p>
        </div>
      </div>
    </header>
    <div class="single-content" v-html="post.Content"></div>
  </article>
  <Pagination :prevPost="prevPost" :nextPost="nextPost"></Pagination>
  <div class="back-to-top">
    <a href="#top">back to top</a>
  </div>
</AppLayout>
</template>
