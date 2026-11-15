<template>
<AppLayout :pageTitle="'legal docs | cptx'" :description="'Privacy, cookies, and how cptx.ee works: what the site collects (almost nothing) and why.'" :canonicalURL="'https://cptx.ee/legaldocs/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <h1 class="addons-title">legal docs</h1>
    <AboutNav :current="'legal'"></AboutNav>
    <p class="addons-intro">What this site collects, what it stores in your browser, and how it works. Short version: no trackers, no ads, no cookies.</p>
    <a v-for="doc in docs" class="addon-card ic-card" :href="doc.Href">
      <LineIcon :name="doc.Icon"></LineIcon>
      <span class="ic-text">
        <span class="addon-card-title">{{ doc.Title }} →</span>
        <span class="addon-card-desc">{{ doc.Desc }}<span v-if="doc.Updated"> · updated {{ doc.Updated }}</span></span>
      </span>
    </a>
  </div>
</AppLayout>
</template>
