<template>
<AppLayout :pageTitle="'written by me | archives | cptx'" :description="'The books I wrote: poems and recuerdos, with links to buy them.'" :canonicalURL="'https://cptx.ee/archives/written-by-me/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons archives">
    <p class="addons-crumb"><a href="/archives/">archives</a> / written by me</p>
    <h1 class="addons-title">written by me</h1>
    <p class="addons-intro">The books I wrote: poems and recuerdos.</p>
    <section class="ar-mine">
      <div class="ar-mine-row">
        <div v-for="b in mine" class="ar-mine-book">
          <a :href="b.Href" class="lib-cover-link" :title="b.Label">
            <img v-if="b.HasCover" class="lib-cover" :src="b.Thumb" :alt="b.Label" height="150" :width="b.CoverW" loading="lazy">
            <span v-else class="lib-gen" :style="b.GenStyle" :lang="b.Lang"><span class="lib-gen-author">{{ b.Author }}</span><span class="lib-gen-mid"><span class="lib-gen-title">{{ b.Main }}</span><span v-if="b.Sub" class="lib-gen-sub">{{ b.Sub }}</span></span><span class="lib-gen-year">{{ b.Year }}</span></span>
          </a>
          <div class="bk-cap" :style="b.CapStyle" :lang="b.Lang">
            <a :href="b.Href" class="bk-cap-title">{{ b.Main }}</a>
            <span v-if="b.Sub" class="bk-cap-sub">{{ b.Sub }}</span>
            <span class="bk-cap-year">{{ b.Year }}</span>
            <a class="bk-cap-buy" :href="b.Buy" rel="noopener">buy →</a>
          </div>
        </div>
      </div>
    </section>
  </div>
</AppLayout>
</template>
