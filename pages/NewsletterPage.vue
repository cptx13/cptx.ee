<template>
<AppLayout :pageTitle="'newsletter | cptx'" :description="'The cptx.ee newsletter: a weekly issue (what changed, what was written, what is being written, other things worth passing on) and the occasional note about the site.'" :canonicalURL="'https://cptx.ee/newsletter/'" :lang="'en'" :ogType="'website'">
  <div class="list-container addons">
    <h1 class="addons-title">newsletter</h1>
    <p class="addons-intro">The newsletter brings a weekly <strong>issue</strong> by email, and now and then a short note about the site. Every issue is also kept here, on the <a href="/newsletter/kiosk/">kiosk</a>.</p>

    <p class="nl-lead">in every issue</p>
    <div class="nl-sections">
      <p><span class="nl-k">what changed</span> on the site</p>
      <p><span class="nl-k">what was written</span> this week</p>
      <p><span class="nl-k">what is being written</span> and might come next</p>
      <p><span class="nl-k">other stuff</span> weekend reads, music, things worth passing on</p>
    </div>

    <form class="nl-form" action="#" method="post" onsubmit="return false">
      <label for="nl-email">your email</label>
      <div class="nl-row">
        <input id="nl-email" type="email" name="email" placeholder="you@example.com" disabled>
        <button type="submit" disabled>subscribe</button>
      </div>
      <p class="nl-note">Sign-up opens soon. Only your email address is kept; you confirm by email first, and every email has a one-click unsubscribe. <a href="/legaldocs/privacy/#newsletter">How the newsletter works →</a></p>
    </form>

    <h2 class="addons-sub">latest issues <a class="nl-more" href="/newsletter/kiosk/">the kiosk →</a></h2>
    <div class="kiosk">
      <KioskCover v-for="issue in issues" :issue="issue"></KioskCover>
      <a v-if="!hasIssues" class="kiosk-issue kiosk-ghost" href="/newsletter/kiosk/">
        <span class="kiosk-mast">cptx</span>
        <span class="kiosk-no">issue 01</span>
        <span class="kiosk-head">coming soon</span>
        <span class="kiosk-date">the first issue is on its way</span>
      </a>
    </div>
  </div>
</AppLayout>
</template>
