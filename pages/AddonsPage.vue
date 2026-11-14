<template>
<AppLayout :pageTitle="'addons | cptx'" :description="'Ways to wander the archive: a heatmap of every post and what was written on this day.'" :canonicalURL="'https://cptx.ee/addons/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <h1 class="addons-title">addons</h1>
    <p class="addons-intro">My playground: where I try new features that make the blog more immersive. Other ways into the {{ count }} croquis and recuerdos.</p>
    <a class="addon-card" href="/addons/map/">
      <span class="addon-card-title">archive map →</span>
      <span class="addon-card-desc">Every day of writing as a heatmap. Filter by year, section, tag; colour by posts or words. Every view has its own shareable link.</span>
    </a>
    <a class="addon-card" href="/addons/today/">
      <span class="addon-card-title">on this day →</span>
      <span class="addon-card-desc">What was written on today's date in earlier years.</span>
    </a>
    <a class="addon-card" href="/addons/places/">
      <span class="addon-card-title">city lights →</span>
      <span class="addon-card-desc">A pixel map of the world: every city mentioned in a post lights up. Tap a light to read where it appears.</span>
    </a>
    <a class="addon-card" href="/addons/words/">
      <span class="addon-card-title">words →</span>
      <span class="addon-card-desc">The words used most, without articles or pronouns, as a heatmap. Tap a word to see it year by year.</span>
    </a>
  </div>
</AppLayout>
</template>
