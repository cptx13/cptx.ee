<template>
<AppLayout :pageTitle="'addons | cptx'" :description="'Ways to wander the archive: a heatmap of every post and what was written on this day.'" :canonicalURL="'https://cptx.ee/addons/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <h1 class="addons-title">addons</h1>
    <p class="addons-intro">Other ways into the {{ count }} croquis and recuerdos.</p>
    <a class="addon-card" href="/addons/map/">
      <span class="addon-card-title">archive map →</span>
      <span class="addon-card-desc">Every day of writing as a heatmap. Filter by year, section, tag; colour by posts or words. Every view has its own shareable link.</span>
    </a>
    <a class="addon-card" href="/addons/today/">
      <span class="addon-card-title">on this day →</span>
      <span class="addon-card-desc">What was written on today's date in earlier years.</span>
    </a>
  </div>
</AppLayout>
</template>
