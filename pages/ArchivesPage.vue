<template>
<AppLayout :pageTitle="'archives | cptx'" :description="'Things we keep: the books I wrote, and the books, records and films Merzeede and I own. CDs are next.'" :canonicalURL="'https://cptx.ee/archives/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons archives">
    <h1 class="addons-title">archives</h1>
    <p class="addons-intro">Things we keep: the books I wrote, and the books, records and films Merzeede and I own. CDs are next.</p>

    <a v-if="hasMine" class="addon-card ar-card ic-card" href="/archives/written-by-me/">
      <LineIcon :name="'pen'"></LineIcon>
      <span class="ic-text">
        <span class="addon-card-title">written by me →</span>
        <span class="addon-card-desc">The {{ mineCount }} books I wrote: poems and recuerdos, with links to buy them.</span>
        <span class="ar-covers" aria-hidden="true">
          <span v-for="b in mine" class="ar-cover">
            <img v-if="b.HasCover" :src="b.Thumb" alt="" height="64" loading="lazy">
            <span v-else class="ar-cover-gen" :style="b.GenStyle"></span>
          </span>
        </span>
      </span>
    </a>
    <a class="addon-card ar-card ic-card" href="/archives/library/">
      <LineIcon :name="'books'"></LineIcon>
      <span class="ic-text">
        <span class="addon-card-title">library →</span>
        <span class="addon-card-desc">Some of the books Merzeede and I own: {{ count }} so far, as not all of them have been scanned. As covers or spines; shelve them by genre, colour, author, decade, language or size.</span>
        <span class="ar-strip" aria-hidden="true"><span v-for="b in strip" class="lib-spine" :style="b.MiniStyle"></span></span>
      </span>
    </a>
    <a v-if="hasVinyl" class="addon-card ar-card ic-card" href="/archives/vinyl/">
      <LineIcon :name="'vinyl'"></LineIcon>
      <span class="ic-text">
        <span class="addon-card-title">vinyl →</span>
        <span class="addon-card-desc">Some of the records Merzeede and I own: {{ vinylCount }} so far. Sort them by genre, artist, decade or colour, and listen on Spotify or YouTube Music.</span>
        <span class="ar-covers" aria-hidden="true">
          <span v-for="r in vinylStrip" class="ar-cover">
            <img v-if="r.HasCover" :src="r.Thumb" alt="" width="64" height="64" loading="lazy">
            <span v-else class="ar-cover-gen vn-mini" :style="r.GenStyle"></span>
          </span>
        </span>
      </span>
    </a>
    <div v-else class="addon-card ar-card ar-soon ic-card">
      <LineIcon :name="'vinyl'"></LineIcon>
      <span class="ic-text">
        <span class="addon-card-title">vinyl</span>
        <span class="addon-card-desc">Soon: records, with their covers and pressings.</span>
      </span>
    </div>
    <a v-if="hasFilms" class="addon-card ar-card ic-card" href="/archives/films/">
      <LineIcon :name="'film'"></LineIcon>
      <span class="ic-text">
        <span class="addon-card-title">films →</span>
        <span class="addon-card-desc">Some of the films Merzeede and I own: {{ filmsCount }} so far. Sort them by genre, title, director, decade or colour, and find where to watch them.</span>
        <span class="ar-covers" aria-hidden="true">
          <span v-for="f in filmsStrip" class="ar-cover">
            <img v-if="f.HasCover" class="fm-thumb" :src="f.Thumb" alt="" width="43" height="64" loading="lazy">
            <span v-else class="ar-cover-gen fm-mini" :style="f.GenStyle"></span>
          </span>
        </span>
      </span>
    </a>
    <div class="addon-card ar-card ar-soon ic-card">
      <LineIcon :name="'cd'"></LineIcon>
      <span class="ic-text">
        <span class="addon-card-title">CDs</span>
        <span class="addon-card-desc">Next.</span>
      </span>
    </div>
  </div>
</AppLayout>
</template>
