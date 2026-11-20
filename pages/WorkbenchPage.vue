<template>
<AppLayout :pageTitle="'workbench | cptx'" :description="'My projects: an iPod, a de-kindled Kindle, a retro handheld, Pop!_OS, a home server, this site, and how each one is put together.'" :canonicalURL="'https://cptx.ee/pockets/workbench/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/pockets/">pockets</a> / workbench</p>
    <h1 class="addons-title">workbench</h1>
    <p class="addons-intro">What is on my bench, and how each one is put together.</p>
    <div class="wb-grid">
      <a v-for="p in projects" class="wb-card" :href="p.Href">
        <span class="wb-top"><LineIcon :name="p.Icon"></LineIcon><span :class="'wb-status wb-' + p.Status">{{ p.StatusLabel }}</span></span>
        <span class="wb-title">{{ p.Title }} →</span>
        <span class="wb-summary">{{ p.Summary }}</span>
      </a>
    </div>
    <h2 class="addons-sub">also on the bench</h2>
    <p class="addons-intro">Every project has the same page: the hardware, the software, a step-by-step guide and notes. My books live in <a href="/archives/written-by-me/">archives / written by me</a>.</p>
  </div>
</AppLayout>
</template>
