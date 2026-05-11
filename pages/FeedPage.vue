<template>
<AppLayout :pageTitle="'cheatsheet feed | cptx'" :description="'RSS feed from cheatsheet blogs'" :canonicalURL="'https://cptx.ee/pockets/cheatsheet/feed/'">
  <div class="list-container" style="margin-top: 2.5em; padding-bottom: 3em;">
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 1em;">
      <h2 style="font-size: 1.1em; margin: 0; font-family: inherit;">cheatsheet feed</h2>
      <a href="/pockets/cheatsheet/feed/refresh" class="feed-refresh-btn" title="Refresh feeds">↻ refresh</a>
    </div>
    <p style="color: var(--content-secondary); font-size: 0.8em; margin-bottom: 1.5em;">
      Latest articles from <a href="/pockets/" style="text-decoration: underline;">cheatsheet</a> blogs.
      <span v-if="lastRefreshed"> Last refreshed: {{ lastRefreshed }}</span>
    </p>

    <div v-if="noCache" style="color: var(--content-secondary); margin-top: 2em; text-align: center;">
      <p style="font-size: 0.95em;">No cached feed data yet.</p>
      <p style="font-size: 0.85em; margin-top: 0.5em;">Click <a href="/pockets/cheatsheet/feed/refresh" style="text-decoration: underline;">↻ refresh</a> to fetch articles.</p>
    </div>

    <div v-if="hasItems" class="feed-list">
      <div v-for="item in items" class="feed-item">
        <div class="feed-item-meta">
          <span class="feed-item-source">{{ item.Source }}</span>
          <span class="feed-item-date">{{ item.DateFormatted }}</span>
        </div>
        <a :href="item.Link" target="_blank" rel="noopener" class="feed-item-title">{{ item.Title }}</a>
      </div>
    </div>

    <div v-if="hasErrors" style="margin-top: 2em;">
      <details style="color: var(--content-secondary); font-size: 0.75em;">
        <summary style="cursor: pointer;">{{ errorCount }} feed(s) could not be fetched</summary>
        <ul style="margin-top: 0.5em; list-style: disc; margin-left: 1.5em;">
          <li v-for="e in errors">{{ e }}</li>
        </ul>
      </details>
    </div>
  </div>

  <style>
    .feed-refresh-btn {
      font-size: 0.8em;
      color: var(--content-secondary);
      border: 1px solid var(--code-border);
      padding: 0.25em 0.7em;
      border-radius: 4px;
      text-decoration: none;
      transition: color 0.2s, border-color 0.2s;
    }
    .feed-refresh-btn:hover {
      color: var(--content-primary);
      border-color: var(--content-primary);
    }
    .feed-list {
      display: flex;
      flex-direction: column;
      gap: 0;
    }
    .feed-item {
      padding: 0.7em 0;
      border-bottom: 1px solid var(--code-border);
    }
    .feed-item:first-child {
      border-top: 1px solid var(--code-border);
    }
    .feed-item-meta {
      display: flex;
      justify-content: space-between;
      font-size: 0.7em;
      color: var(--content-secondary);
      margin-bottom: 0.2em;
    }
    .feed-item-source {
      font-weight: 700;
    }
    .feed-item-date {
      white-space: nowrap;
    }
    .feed-item-title {
      font-size: 0.9em;
      text-decoration: none;
      color: var(--content-primary);
    }
    .feed-item-title:hover {
      text-decoration: underline;
    }
  </style>
</AppLayout>
</template>
