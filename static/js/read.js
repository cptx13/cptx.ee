// A-02: counts a "read" once the end of a post has been on screen and enough
// time has passed to have read it (about a third of the reading time, 5–90 s).
// Sends one small beacon with the post's path. No cookies, nothing stored.
(function () {
  var m = document.getElementById("read-marker");
  if (!m || !navigator.sendBeacon || !("IntersectionObserver" in window)) return;
  var words = parseInt(m.getAttribute("data-words"), 10) || 0;
  var need = Math.min(90, Math.max(5, (words / 230) * 60 / 3)) * 1000;
  var start = Date.now(), seenEnd = false, sent = false;
  function check() {
    if (sent || !seenEnd || Date.now() - start < need) return;
    sent = true;
    navigator.sendBeacon("/_read", new Blob([m.getAttribute("data-path")], { type: "text/plain" }));
  }
  new IntersectionObserver(function (es) {
    es.forEach(function (e) { if (e.isIntersecting) { seenEnd = true; check(); } });
  }).observe(m);
  var t = setInterval(function () { check(); if (sent) clearInterval(t); }, 2000);
  document.addEventListener("visibilitychange", function () { if (document.visibilityState === "hidden") check(); });
})();
