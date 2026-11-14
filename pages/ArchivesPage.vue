<template>
<AppLayout :pageTitle="'archives | cptx'" :description="'Things I keep: the books I wrote, my library, and later records, films and other collections.'" :canonicalURL="'https://cptx.ee/archives/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons archives">
    <h1 class="addons-title">archives</h1>
    <p class="addons-intro">Things I keep. The books I wrote, the books on my shelves, and later the records, films and other collections.</p>

    <section v-if="hasMine" class="ar-mine">
      <h2 class="lib-shelf-label">written by me</h2>
      <div class="ar-mine-row">
        <div v-for="b in mine" class="ar-mine-book">
          <a :href="b.Href" class="lib-cover-link" :title="b.Label">
            <img v-if="b.HasCover" class="lib-cover" :src="b.Thumb" :alt="b.Label" height="150" :width="b.CoverW" loading="lazy">
            <span v-else class="lib-gen" :style="b.GenStyle"><span class="lib-gen-title">{{ b.Title }}</span><span class="lib-gen-author">{{ b.Author }}</span></span>
          </a>
          <p class="ar-mine-meta">{{ b.Title }} <span class="ar-mine-year">{{ b.Year }}</span><br><a :href="b.Buy" rel="noopener">buy →</a></p>
        </div>
      </div>
    </section>

    <a class="addon-card ar-card" href="/archives/library/">
      <span class="addon-card-title">library →</span>
      <span class="addon-card-desc">{{ count }} books on my shelves, as covers or spines. Shelve them by genre, colour, author, decade, language or size.</span>
      <span class="ar-strip" aria-hidden="true"><span v-for="b in strip" class="lib-spine" :style="b.MiniStyle"></span></span>
    </a>
    <div class="addon-card ar-card ar-soon">
      <span class="addon-card-title">vinyl</span>
      <span class="addon-card-desc">Soon: records, with their covers and pressings.</span>
    </div>
    <div class="addon-card ar-card ar-soon">
      <span class="addon-card-title">films &amp; other collections</span>
      <span class="addon-card-desc">Later.</span>
    </div>
  </div>
</AppLayout>
</template>
