<template>
<AppLayout :pageTitle="'about | cptx'" :description="'About cptx.ee: the manifesto, how to get in touch, and the legal docs.'" :canonicalURL="'https://cptx.ee/about/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <h1 class="addons-title">about</h1>
    <p class="addons-intro">Why this site exists, how to reach me, and what it does with your data (almost nothing).</p>
    <a class="addon-card ic-card" href="/manifesto/">
      <LineIcon :name="'paper'"></LineIcon>
      <span class="ic-text">
      <span class="addon-card-title">manifesto →</span>
      <span class="addon-card-desc">Why this site exists: a collection of mathoms, the principles behind it, and the sections.</span>
    </span>
    </a>
    <a class="addon-card ic-card" href="/contact/">
      <LineIcon :name="'mail'"></LineIcon>
      <span class="ic-text">
      <span class="addon-card-title">contact →</span>
      <span class="addon-card-desc">Email, and where else to find me.</span>
    </span>
    </a>
    <a class="addon-card ic-card" href="/legaldocs/">
      <LineIcon :name="'shield'"></LineIcon>
      <span class="ic-text">
      <span class="addon-card-title">legal docs →</span>
      <span class="addon-card-desc">Privacy, cookies and how the site works. No trackers, no ads, no cookies.</span>
    </span>
    </a>
  </div>
</AppLayout>
</template>
