<template>
<AppLayout :pageTitle="'on this day: ' + dateLabel + ' | cptx'" :description="'What was written on ' + dateLabel + ' in earlier years.'" :canonicalURL="'https://cptx.ee/addons/today/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/addons/">addons</a> / on this day</p>
    <h1 class="addons-title">{{ dateLabel }}</h1>
    <p class="today-nav"><a :href="prevHref" rel="nofollow">← previous day</a> · <a href="/addons/today/">today</a> · <a :href="nextHref" rel="nofollow">next day →</a></p>
    <div v-if="hasPosts">
      <article v-for="p in posts" class="post-entry">
        <span class="post-entry-date">{{ p.Year }}</span>
        <h2 class="post-entry-title"><a :href="p.Path">{{ p.Title }}</a></h2>
      </article>
    </div>
    <p v-else class="addons-intro">Nothing was written on {{ dateLabel }} in earlier years.</p>
  </div>
</AppLayout>
</template>
