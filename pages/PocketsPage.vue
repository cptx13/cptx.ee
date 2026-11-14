<template>
<AppLayout :pageTitle="'pockets | cptx'" :description="'A collection of stuff I would love to carry in my pockets: books, links, projects and articles.'" :canonicalURL="'https://cptx.ee/pockets/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <h1 class="addons-title">pockets</h1>
    <p class="addons-intro">A collection of stuff I would love to carry in my pockets.</p>
    <a v-for="pocket in pockets" class="addon-card ic-card" :href="pocket.Href">
      <LineIcon :name="pocket.Icon"></LineIcon>
      <span class="ic-text">
        <span class="addon-card-title">{{ pocket.Title }} →</span>
        <span class="addon-card-desc">{{ pocket.Desc }}<span v-if="pocket.Meta"> · {{ pocket.Meta }}</span></span>
      </span>
    </a>
  </div>
</AppLayout>
</template>
