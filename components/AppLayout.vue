<template>
<html lang="en-us" dir="ltr">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width">
  <link rel="icon" type="image/ico" href="/favicon.ico">
  <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
  <link rel="icon" type="image/png" sizes="192x192" href="/android-chrome-192x192.png">
  <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
  <meta name="description" :content="description">
  <title>{{ pageTitle }}</title>
  <link rel="canonical" :href="canonicalURL">
  <link rel="stylesheet" href="/css/style.css" media="all">
</head>
<body class="auto">
  <div class="content">
    <header>
      <SiteHeader></SiteHeader>
    </header>
    <main class="main">
      <slot />
    </main>
  </div>
  <SiteFooter></SiteFooter>
  <script src="/js/theme-switch.js"></script>
  <script defer src="/js/copy-code.js"></script>
  <script>
  (function () {
    var button = document.getElementById("random-post-button");
    var pool = document.querySelectorAll("#random-post-pool a");
    if (!button || !pool.length) return;
    button.addEventListener("click", function (event) {
      event.preventDefault();
      var index = Math.floor(Math.random() * pool.length);
      var target = pool[index].getAttribute("href");
      if (target) window.location.href = target;
    });
  })();
  </script>
</body>
</html>
</template>
