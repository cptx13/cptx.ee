<template>
<AppLayout :pageTitle="folder.Title + ' | photos | cptx'" :description="folder.Title + ' - photo gallery'" :canonicalURL="'https://cptx.ee' + folder.Path">
  <div class="list-container" style="margin-top: 2.5em;">
    <p style="margin-bottom: 1.5em;"><a href="/photos/">← back to photos</a></p>
    <h1 style="font-size: 1.6em; margin-bottom: 1em;">{{ folder.Title }}</h1>
    <p style="color: var(--content-secondary); margin-bottom: 2em;">{{ folder.Count }} photos</p>

    <div class="photo-grid">
      <div v-for="photo in folder.Photos" class="photo-item">
        <a :href="photo.URL" target="_blank">
          <img :src="photo.URL" :alt="photo.ID" loading="lazy">
          <span class="photo-id">{{ photo.ID }}</span>
        </a>
      </div>
    </div>
  </div>

  <div class="back-to-top">
    <a href="#top">back to top</a>
  </div>
</AppLayout>
</template>
