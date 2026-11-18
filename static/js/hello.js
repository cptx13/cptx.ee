// AD-07/AD-09: says "a person is here" once per page, so the readers page
// counts people, not scrapers. Sent on the reader's first scroll, tap, click
// or key press. No timers: nothing measures how long anyone stays. Automated
// browsers (navigator.webdriver) never send it. The beacon carries only the
// page's path. No cookies, nothing stored.
(function () {
  if (!navigator.sendBeacon || navigator.webdriver) return;
  var events = ["scroll", "wheel", "keydown", "pointerdown", "touchstart"];
  function hello() {
    events.forEach(function (e) { removeEventListener(e, hello, true); });
    var p = location.pathname;
    try { p = decodeURIComponent(p); } catch (e) {}
    navigator.sendBeacon("/_hello", new Blob([p], { type: "text/plain" }));
  }
  events.forEach(function (e) { addEventListener(e, hello, { capture: true, passive: true }); });
})();
