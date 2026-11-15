<template>
<AppLayout :pageTitle="'addons | cptx'" :description="'Ways to wander the archive: a heatmap of every post and what was written on this day.'" :canonicalURL="'https://cptx.ee/addons/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <h1 class="addons-title">addons</h1>
    <p class="addons-intro">My playground: where I try new features that make the blog more immersive. Other ways into the {{ count }} croquis and recuerdos.</p>
    <a class="addon-card ic-card" href="/addons/map/">
      <LineIcon :name="'grid'"></LineIcon>
      <span class="ic-text">
      <span class="addon-card-title">archive map →</span>
      <span class="addon-card-desc">Every day of writing as a heatmap. Filter by year, section, tag; colour by posts or words. Every view has its own shareable link.</span>
    </span>
    </a>
    <a class="addon-card ic-card" href="/addons/today/">
      <LineIcon :name="'calendar'"></LineIcon>
      <span class="ic-text">
      <span class="addon-card-title">on this day →</span>
      <span class="addon-card-desc">What was written on today's date in earlier years.</span>
    </span>
    </a>
    <a class="addon-card ic-card" href="/addons/places/">
      <LineIcon :name="'globe'"></LineIcon>
      <span class="ic-text">
      <span class="addon-card-title">city lights →</span>
      <span class="addon-card-desc">A pixel map of the world: every city mentioned in a post lights up. Tap a light to read where it appears.</span>
    </span>
    </a>
    <a class="addon-card ic-card" href="/addons/words/">
      <LineIcon :name="'words'"></LineIcon>
      <span class="ic-text">
      <span class="addon-card-title">words →</span>
      <span class="addon-card-desc">The words used most, without articles or pronouns, as a heatmap. Tap a word to see it year by year.</span>
    </span>
    </a>
    <a class="addon-card ic-card" href="/addons/readers/">
      <LineIcon :name="'eye'"></LineIcon>
      <span class="ic-text">
      <span class="addon-card-title">readers →</span>
      <span class="addon-card-desc">Who comes by, counted without cookies: visits month by month, the most read posts, and a map of where readers are.</span>
    </span>
    </a>
  </div>
</AppLayout>
</template>
