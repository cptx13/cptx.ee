<template>
<AppLayout :pageTitle="folder.Title + ' | photos | cptx'" :description="folder.Location + ', ' + folder.Year + ': ' + folder.Count + ' photos'" :canonicalURL="'https://cptx.ee' + folder.Path" :lang="'en'" :ogType="'website'">
  <div class="list-container addons ph">
    <p class="addons-crumb"><a href="/photos/">photos</a> /</p>
    <h1 class="addons-title">{{ folder.Location }}</h1>
    <p class="addons-intro">{{ folder.Year }} · {{ folder.Count }} photos</p>

    <div class="ph-masonry ph-masonry-lg">
      <a v-for="photo in folder.Photos" class="ph-item" :href="photo.URL" :data-id="photo.ID">
        <img v-if="photo.HasSize" :src="photo.ThumbURL" :alt="folder.Title + ', photo ' + photo.Index" :width="photo.W" :height="photo.H" loading="lazy">
        <img v-else :src="photo.ThumbURL" :alt="folder.Title + ', photo ' + photo.Index" loading="lazy">
      </a>
    </div>

    <p class="ph-top"><a href="#top">back to top ↑</a></p>
  </div>
  <script src="/js/photos.js" defer></script>
</AppLayout>
</template>
