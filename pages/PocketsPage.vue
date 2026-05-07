<template>
<AppLayout :pageTitle="'pockets | cptx'" :description="'pockets - cptx'" :canonicalURL="'https://cptx.ee/pockets/'">
  <div class="pockets-container" style="margin-top: 2.5em; padding-bottom: 3em;">
    <p style="color: #8c8c8c; font-size: 0.9em; margin-bottom: 1.5em;">A collection of stuff I would love to carry in my pockets.</p>
    <div class="pocket-grid">
      <div v-for="(pocket, idx) in pockets" class="pocket-card" :class="{ active: idx == activeIndex }" :data-index="idx">
        <div class="pc-icon">{{ pocket.Icon }}</div>
        <div class="pc-title">{{ pocket.Title }}</div>
        <div class="pc-desc">{{ pocket.Desc }}</div>
      </div>
    </div>
    <div v-for="(pocket, idx) in pockets" class="pocket-expanded" :class="{ active: idx == activeIndex }" :data-panel="idx">
      <div v-html="pocket.ContentHTML"></div>
    </div>
  </div>
  <script>
  (function() {
    document.addEventListener('click', function(e) {
      var card = e.target.closest('.pocket-card');
      if (!card) return;
      var idx = card.getAttribute('data-index');
      document.querySelectorAll('.pocket-card').forEach(function(c) {
        c.classList.toggle('active', c.getAttribute('data-index') === idx);
      });
      document.querySelectorAll('.pocket-expanded').forEach(function(p) {
        p.classList.toggle('active', p.getAttribute('data-panel') === idx);
      });
    });
  })();
  </script>
</AppLayout>
</template>
