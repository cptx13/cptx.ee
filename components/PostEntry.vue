<template>
<article style="display: flex; align-items: baseline; gap: 0.5em; margin-bottom: 0.5em; font-family: inherit; font-size: 1em;">
  <span style="color: #888; font-size: 0.95em; min-width: 8ch; text-align: right; font-family: inherit;">{{ dateFormatted }}</span>
  <h2 style="margin: 0; font-size: 1.1em; display: inline; font-family: inherit;"><a :href="path">{{ title }}</a></h2>
</article>
</template>
