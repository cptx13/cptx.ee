<template>
<AppLayout :pageTitle="Title + ' | workbench | cptx'" :description="Summary" :canonicalURL="'https://cptx.ee' + Href" :lang="'en'" :ogType="'article'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/pockets/">pockets</a> / <a href="/pockets/workbench/">workbench</a> / {{ Title }}</p>
    <h1 class="addons-title">{{ Title }}</h1>
    <p class="hm-stats"><span :class="'wb-status wb-' + Status">{{ StatusLabel }}</span> {{ Summary }}</p>
    <p v-if="hasLink" class="hm-stats"><a :href="link">{{ linkLabel }} ↗</a></p>
    <h2 class="addons-sub">hardware</h2>
    <p class="wb-kit"><span v-for="k in hardware" class="wb-part">{{ k }}</span></p>
    <h2 class="addons-sub">software</h2>
    <p class="wb-kit"><span v-for="k in software" class="wb-part">{{ k }}</span></p>
    <div class="wb-body" v-html="body"></div>
    <p class="hm-stats wb-back"><a href="/pockets/workbench/">← all projects</a></p>
  </div>
</AppLayout>
</template>
