<template>
<a class="kiosk-issue" :href="issue.Path">
  <span class="kiosk-mast">cptx</span>
  <span class="kiosk-no">No. {{ issue.No }} · {{ issue.Week }}</span>
  <span class="kiosk-head">{{ issue.Title }}</span>
  <span class="kiosk-date">{{ issue.Date }}</span>
</a>
</template>
