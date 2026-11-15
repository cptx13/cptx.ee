<template>
<a :class="'kiosk-issue ' + issue.Palette" :href="issue.Path">
  <span class="kiosk-art-wrap" v-html="issue.Art"></span>
  <span class="kiosk-mast">cptx</span>
  <span class="kiosk-no">{{ issue.Label }} · {{ issue.Week }}</span>
  <span class="kiosk-head">{{ issue.Title }}</span>
  <span class="kiosk-date">{{ issue.Date }}<span v-if="issue.Reads"> · {{ issue.Reads }}</span></span>
</a>
</template>
