// AD-07: says "a person is here" once per page, so the readers page counts
// people, not scrapers. Sent after the page has been on screen for 1 s and the
// reader scrolls, taps, clicks or types, or after 10 s on screen without that.
// Automated browsers (navigator.webdriver) never send it. The beacon carries
// only the page's path. No cookies, nothing stored.
(function () {
  if (!navigator.sendBeacon || navigator.webdriver) return;
  var sent = false, shown = 0, acted = false;
  var events = ["scroll", "wheel", "keydown", "pointerdown", "touchstart"];
  function send() {
    if (sent) return;
    sent = true;
    clearInterval(timer);
    events.forEach(function (e) { removeEventListener(e, act, true); });
    var p = location.pathname;
    try { p = decodeURIComponent(p); } catch (e) {}
    navigator.sendBeacon("/_hello", new Blob([p], { type: "text/plain" }));
  }
  function check() {
    if ((acted && shown >= 1) || shown >= 10) send();
  }
  function act() { acted = true; check(); }
  events.forEach(function (e) { addEventListener(e, act, { capture: true, passive: true }); });
  var timer = setInterval(function () {
    if (document.visibilityState === "visible") shown++;
    check();
  }, 1000);
})();
