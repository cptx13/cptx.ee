<template>
<div v-if="prevPost || nextPost" class="single-pagination">
  <hr>
  <div class="flexnowrap">
    <div class="single-pagination-prev">
      <div v-if="prevPost" class="single-pagination-container-prev">
        <div class="single-pagination-text">&larr;</div>
        <div class="single-pagination-text">
          <a :href="prevPost.Path">{{ prevPost.Title }}</a>
        </div>
      </div>
    </div>
    <div class="single-pagination-next">
      <div v-if="nextPost" class="single-pagination-container-next">
        <div class="single-pagination-text">
          <a :href="nextPost.Path">{{ nextPost.Title }}</a>
        </div>
        <div class="single-pagination-text">&rarr;</div>
      </div>
    </div>
  </div>
  <hr>
</div>
</template>
