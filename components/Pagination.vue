<template>
<nav v-if="prevPost || nextPost || post.ListPath" class="post-nav" aria-label="More to read">
  <div class="post-nav-pair">
    <a v-if="prevPost" class="post-nav-link post-nav-prev" :href="prevPost.Path" rel="prev">
      <span class="post-nav-label">&larr; older</span>
      <span class="post-nav-title">{{ prevPost.Title }}</span>
      <span class="post-nav-date">{{ prevPost.DateFormatted }}</span>
    </a>
    <a v-if="nextPost" class="post-nav-link post-nav-next" :href="nextPost.Path" rel="next">
      <span class="post-nav-label">newer &rarr;</span>
      <span class="post-nav-title">{{ nextPost.Title }}</span>
      <span class="post-nav-date">{{ nextPost.DateFormatted }}</span>
    </a>
  </div>
  <p v-if="post.ListPath" class="post-nav-all"><a :href="post.ListPath">{{ post.ListLabel }}</a></p>
</nav>
</template>
