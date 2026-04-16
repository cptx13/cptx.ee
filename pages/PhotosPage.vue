<template>
<AppLayout :pageTitle="'photos | cptx'" :description="'photo galleries'" :canonicalURL="'https://cptx.ee/photos/'">
  <div class="list-container" style="margin-top: 2.5em;">
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5em;">
      <h2 style="font-size: 1.1em; margin: 0; font-family: inherit;">photos</h2>
      <div class="photos-toggle-group">
        <button id="toggle-folders" class="photos-toggle active" onclick="switchView('folders')">folders</button>
        <button id="toggle-all" class="photos-toggle" onclick="switchView('all')">all</button>
      </div>
    </div>

    <div id="view-folders" class="folder-grid">
      <a v-for="folder in folders" :href="folder.Path" class="folder-card">
        <span class="folder-card-title">{{ folder.Title }}</span>
        <span class="folder-card-count">{{ folder.Count }} photos</span>
      </a>
    </div>

    <div id="view-all" class="photo-grid" style="display: none;">
      <div v-for="photo in allPhotos" class="photo-item">
        <a :href="photo.URL" target="_blank">
          <img :src="photo.ThumbURL" :alt="photo.ID" loading="lazy">
          <span class="photo-id">{{ photo.ID }}</span>
        </a>
      </div>
    </div>
  </div>

  <script>
  function switchView(view) {
    var foldersView = document.getElementById('view-folders');
    var allView = document.getElementById('view-all');
    var btnFolders = document.getElementById('toggle-folders');
    var btnAll = document.getElementById('toggle-all');
    if (view === 'folders') {
      foldersView.style.display = '';
      allView.style.display = 'none';
      btnFolders.classList.add('active');
      btnAll.classList.remove('active');
    } else {
      foldersView.style.display = 'none';
      allView.style.display = '';
      btnFolders.classList.remove('active');
      btnAll.classList.add('active');
    }
  }
  </script>
</AppLayout>
</template>
