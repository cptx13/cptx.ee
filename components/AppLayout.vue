<template>
<html :lang="lang" dir="ltr">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light dark">
  <script>(function () { var t; try { t = localStorage.getItem("theme"); } catch (e) {} if (!t) t = window.matchMedia && matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"; if (t === "dark") document.documentElement.classList.add("dark"); })();</script>
  <link rel="icon" type="image/ico" href="/favicon.ico">
  <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
  <link rel="icon" type="image/png" sizes="192x192" href="/android-chrome-192x192.png">
  <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
  <meta name="description" :content="description">
  <title>{{ pageTitle }}</title>
  <link rel="canonical" :href="canonicalURL">
  <meta property="og:site_name" content="cptx">
  <meta property="og:type" :content="ogType">
  <meta property="og:title" :content="pageTitle">
  <meta property="og:description" :content="description">
  <meta property="og:url" :content="canonicalURL">
  <meta property="og:image" content="https://cptx.ee/android-chrome-192x192.png">
  <meta name="twitter:card" content="summary">
  <link rel="alternate" type="application/rss+xml" title="cptx — croquis" href="/index.xml">
  <link rel="sitemap" type="application/xml" href="/sitemap.xml">
  <link rel="preload" href="/fonts/jetbrains-mono/JetBrainsMono-Regular.woff2" as="font" type="font/woff2" crossorigin>
  <link rel="stylesheet" href="/css/style.css" media="all">
</head>
<body>
  <a class="skip-link" href="#main">Skip to content</a>
  <div class="content">
    <header>
      <SiteHeader></SiteHeader>
    </header>
    <main class="main" id="main">
      <slot />
    </main>
  </div>
  <SiteFooter></SiteFooter>
  <script src="/js/theme-switch.js"></script>
  <script defer src="/js/post-search.js"></script>
  <script>
  (function () {
    var button = document.getElementById("random-post-button");
    if (!button) return;
    var pool = null;
    button.addEventListener("click", function (event) {
      event.preventDefault();
      if (pool) {
        window.location.href = pool[Math.floor(Math.random() * pool.length)];
        return;
      }
      fetch("/shuffle-pool.json").then(function(r){return r.json()}).then(function(data){
        pool = data;
        if (pool.length) window.location.href = pool[Math.floor(Math.random() * pool.length)];
      });
    });
  })();
  </script>
</body>
</html>
</template>
