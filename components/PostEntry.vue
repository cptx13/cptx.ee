<template>
<article style="display: flex; align-items: baseline; gap: 0.5em; margin-bottom: 0.5em; font-family: inherit; font-size: 1em; min-width: 0;">
  <span style="color: #888; font-size: 0.95em; white-space: nowrap; font-family: inherit; flex-shrink: 0;">{{ dateFormatted }}</span>
  <h2 style="margin: 0; font-size: 1.1em; font-family: inherit; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;"><a :href="path">{{ title }}</a></h2>
</article>
</template>
