<template>
<AppLayout :pageTitle="'kiosk | cptx newsletter'" :description="'Every issue of the cptx.ee newsletter, as covers on a kiosk.'" :canonicalURL="'https://cptx.ee/newsletter/kiosk/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <p class="addons-crumb"><a href="/newsletter/">newsletter</a> / kiosk</p>
    <h1 class="addons-title">kiosk</h1>
    <p class="addons-intro"><span v-if="hasIssues">{{ issueCountLabel }}, newest first. Pick one up.</span><span v-else>Every issue of the newsletter ends up here, newest first.</span></p>
    <div class="kiosk">
      <KioskCover v-for="issue in issues" :issue="issue"></KioskCover>
      <div v-if="!hasIssues" class="kiosk-issue kiosk-ghost">
        <span class="kiosk-mast">cptx</span>
        <span class="kiosk-no">No. 1</span>
        <span class="kiosk-head">coming soon</span>
        <span class="kiosk-date">the first issue is on its way</span>
      </div>
    </div>
    <p class="addons-intro"><a href="/newsletter/">← subscribe</a></p>
  </div>
</AppLayout>
</template>
