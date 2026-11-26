<template>
<svg class="line-icon" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
  <g v-if="name == 'bedside'"><path d="M9 3h6l2 5H7z"/><path d="M12 8v4"/><rect x="5" y="12" width="14" height="9" rx="1"/><path d="M5 16.5h14"/><path d="M11 14.2h2"/><path d="M11 18.7h2"/></g>
  <g v-if="name == 'link'"><path d="M10 14a4 4 0 0 0 5.66 0l3-3a4 4 0 0 0-5.66-5.66l-1 1"/><path d="M14 10a4 4 0 0 0-5.66 0l-3 3a4 4 0 0 0 5.66 5.66l1-1"/></g>
  <g v-if="name == 'wrench'"><path d="M14.5 5.5a4.2 4.2 0 0 0-5 5.5L3.5 17a2.1 2.1 0 0 0 3 3l6-6a4.2 4.2 0 0 0 5.5-5l-2.6 2.6-2.4-.6-.6-2.4z"/></g>
  <g v-if="name == 'paper'"><path d="M17 20H5a2 2 0 0 1-2-2V4h14v14a2 2 0 0 0 4 0V8h-4"/><path d="M6 8h8"/><path d="M6 12h3"/><path d="M6 16h3"/><rect x="11.5" y="11" width="2.5" height="5"/></g>
  <g v-if="name == 'grid'"><rect x="3" y="3" width="4.5" height="4.5" rx="1"/><rect x="9.75" y="3" width="4.5" height="4.5" rx="1"/><rect x="16.5" y="3" width="4.5" height="4.5" rx="1"/><rect x="3" y="9.75" width="4.5" height="4.5" rx="1"/><rect x="9.75" y="9.75" width="4.5" height="4.5" rx="1"/><rect x="16.5" y="9.75" width="4.5" height="4.5" rx="1"/><rect x="3" y="16.5" width="4.5" height="4.5" rx="1"/><rect x="9.75" y="16.5" width="4.5" height="4.5" rx="1"/></g>
  <g v-if="name == 'calendar'"><rect x="3" y="5" width="18" height="16" rx="2"/><path d="M3 10h18"/><path d="M8 3v4"/><path d="M16 3v4"/><circle cx="12" cy="15.5" r="1.6"/></g>
  <g v-if="name == 'globe'"><circle cx="12" cy="12" r="9"/><path d="M3 12h18"/><path d="M12 3c2.5 2.6 3.8 5.6 3.8 9s-1.3 6.4-3.8 9c-2.5-2.6-3.8-5.6-3.8-9s1.3-6.4 3.8-9z"/></g>
  <g v-if="name == 'words'"><path d="M3 19 7.5 5 12 19"/><path d="M4.7 14h5.6"/><path d="M15 10h6"/><path d="M15 14.5h6"/><path d="M15 19h4"/></g>
  <g v-if="name == 'pen'"><path d="M12 3l5.5 7.5L12 21l-5.5-10.5z"/><path d="M12 12v4.5"/><circle cx="12" cy="10.5" r="1.3"/></g>
  <g v-if="name == 'books'"><rect x="3.5" y="4" width="4" height="16" rx=".6"/><rect x="8.5" y="6.5" width="4" height="13.5" rx=".6"/><path d="M14.6 7.2l3.6-1 3.4 13.3-3.6 1z"/><path d="M5.5 8v2"/><path d="M10.5 10v2"/></g>
  <g v-if="name == 'vinyl'"><circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="3"/><circle cx="12" cy="12" r=".6"/><path d="M12 6a6 6 0 0 1 6 6"/></g>
  <g v-if="name == 'film'"><rect x="3" y="4" width="18" height="16" rx="1.5"/><path d="M7.5 4v16"/><path d="M16.5 4v16"/><path d="M3 8h4.5"/><path d="M3 12h4.5"/><path d="M3 16h4.5"/><path d="M16.5 8H21"/><path d="M16.5 12H21"/><path d="M16.5 16H21"/></g>
  <g v-if="name == 'shield'"><path d="M12 3l7.5 3v5.5c0 4.6-3.2 8.3-7.5 9.5-4.3-1.2-7.5-4.9-7.5-9.5V6z"/><path d="M8.8 12l2.2 2.2 4.2-4.4"/></g>
  <g v-if="name == 'cookie'"><path d="M20.8 12.6A8.8 8.8 0 1 1 11.4 3.2a3 3 0 0 0 3.7 3.7 3 3 0 0 0 2 3.9 3 3 0 0 0 3.7 1.8z"/><circle cx="8.5" cy="10" r=".9"/><circle cx="11.5" cy="15.5" r=".9"/><circle cx="15.8" cy="14.6" r=".9"/></g>
  <g v-if="name == 'gear'"><circle cx="12" cy="12" r="3"/><path d="M12 2.8v2.4M12 18.8v2.4M2.8 12h2.4M18.8 12h2.4M5.5 5.5l1.7 1.7M16.8 16.8l1.7 1.7M5.5 18.5l1.7-1.7M16.8 7.2l1.7-1.7"/><circle cx="12" cy="12" r="6.4"/></g>
  <g v-if="name == 'mail'"><rect x="3" y="5.5" width="18" height="13" rx="1.5"/><path d="M3.5 6.5 12 13l8.5-6.5"/></g>
  <g v-if="name == 'rss'"><circle cx="5.5" cy="18.5" r="1.5"/><path d="M4 11a9 9 0 0 1 9 9"/><path d="M4 4a16 16 0 0 1 16 16"/></g>
  <g v-if="name == 'eye'"><path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12z"/><circle cx="12" cy="12" r="3"/></g>
  <g v-if="name == 'ipod'"><rect x="6" y="2.5" width="12" height="19" rx="2"/><rect x="8.5" y="5" width="7" height="5.5" rx=".6"/><circle cx="12" cy="16" r="3"/><circle cx="12" cy="16" r=".9"/></g>
  <g v-if="name == 'ereader'"><rect x="5" y="2.5" width="14" height="19" rx="1.8"/><path d="M8 6.5h8"/><path d="M8 9.5h8"/><path d="M8 12.5h8"/><path d="M8 15.5h5"/><path d="M10.5 19h3"/></g>
  <g v-if="name == 'notebook'"><rect x="5" y="3" width="14" height="18" rx="1.5"/><path d="M9 3v18"/><path d="M3.5 7h3"/><path d="M3.5 12h3"/><path d="M3.5 17h3"/><path d="M12 8h4"/></g>
  <g v-if="name == 'server'"><rect x="4" y="4" width="16" height="7" rx="1.2"/><rect x="4" y="13" width="16" height="7" rx="1.2"/><circle cx="7.5" cy="7.5" r=".8"/><circle cx="7.5" cy="16.5" r=".8"/><path d="M11 7.5h6"/><path d="M11 16.5h6"/></g>
  <g v-if="name == 'phone'"><rect x="6.5" y="2.5" width="11" height="19" rx="2"/><path d="M10.5 5h3"/><circle cx="12" cy="18.3" r=".8"/></g>
  <g v-if="name == 'gamepad'"><rect x="6" y="2.5" width="12" height="19" rx="2"/><rect x="8" y="4.5" width="8" height="6" rx=".6"/><path d="M8.5 15h3"/><path d="M10 13.5v3"/><circle cx="15" cy="14" r=".8"/><circle cx="14" cy="16.5" r=".8"/></g>
  <g v-if="name == 'laptop'"><rect x="5" y="5" width="14" height="10" rx="1"/><path d="M3 18.5h18l-1.5-3.5h-15z"/></g>
  <g v-if="name == 'cd'"><circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="2.5"/><path d="M7.2 9.2a5.5 5.5 0 0 1 2-2"/><path d="M16.8 14.8a5.5 5.5 0 0 1-2 2"/></g>
</svg>
</template>
