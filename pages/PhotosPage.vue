<template>
<AppLayout :pageTitle="'photos | cptx'" :description="'Photos from places I have been, by year.'" :canonicalURL="'https://cptx.ee/photos/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons ph">
    <h1 class="addons-title">photos</h1>
    <p class="addons-intro">{{ total }} photos · {{ places }} places · {{ years }}</p>

    <p class="ph-views" role="group" aria-label="View">
      <a href="#places" class="hm-opt ph-view hm-active" data-view="places">places</a>
      <a href="#all" class="hm-opt ph-view" data-view="all">all photos</a>
    </p>

    <div id="places" class="ph-cards">
      <a v-for="folder in folders" :href="folder.Path" class="ph-card">
        <img v-if="folder.Cover" class="ph-card-img" :src="folder.Cover" alt="" loading="lazy">
        <span class="ph-card-text">
          <span class="ph-card-title">{{ folder.Location }}</span>
          <span class="ph-card-meta">{{ folder.Year }} · {{ folder.Count }} photos</span>
        </span>
      </a>
    </div>

    <div id="all" class="ph-all">
      <section v-for="folder in folders" class="ph-group">
        <h2 class="lib-shelf-label"><a :href="folder.Path">{{ folder.Location }}</a> <span class="lib-shelf-count">{{ folder.Year }} · {{ folder.Count }}</span></h2>
        <div class="ph-masonry">
          <a v-for="photo in folder.Photos" class="ph-item" :href="photo.URL" :data-id="photo.ID" :style="'--r:' + photo.Ratio">
            <img v-if="photo.HasSize" :src="photo.ThumbURL" :alt="folder.Title + ', photo ' + photo.Index" :width="photo.W" :height="photo.H" loading="lazy">
            <img v-else :src="photo.ThumbURL" :alt="folder.Title + ', photo ' + photo.Index" loading="lazy">
          </a>
        </div>
      </section>
    </div>
  </div>
  <script>
  (function() {
    var root = document.querySelector('.ph');
    if (!root) return;
    var views = root.querySelectorAll('.ph-view');
    function show(v) {
      if (v !== 'all') v = 'places';
      document.getElementById('places').hidden = v !== 'places';
      document.getElementById('all').hidden = v !== 'all';
      views.forEach(function(a) { a.classList.toggle('hm-active', a.getAttribute('data-view') === v); });
    }
    views.forEach(function(a) {
      a.addEventListener('click', function(e) {
        e.preventDefault();
        var v = a.getAttribute('data-view');
        history.replaceState(null, '', v === 'all' ? '#all' : location.pathname);
        show(v);
      });
    });
    show(location.hash.slice(1));
  })();
  </script>
  <script src="/js/photos.js" defer></script>
</AppLayout>
</template>
