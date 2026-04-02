<template>
<div v-if="tags" class="single-tags">
  <span v-for="tag in tags">#{{ tag }}</span>
</div>
</template>
