// F-02: theme follows the system until the reader picks one; the choice is remembered.
// The initial class is set by an inline script in <head> (no flash).
(function () {
  var root = document.documentElement;
  var button = document.getElementById("theme-toggle");
  function apply(theme) {
    root.classList.toggle("dark", theme === "dark");
    if (button) button.setAttribute("aria-pressed", theme === "dark" ? "true" : "false");
  }
  apply(root.classList.contains("dark") ? "dark" : "light");
  if (button) {
    button.addEventListener("click", function () {
      var next = root.classList.contains("dark") ? "light" : "dark";
      apply(next);
      try { localStorage.setItem("theme", next); } catch (e) {}
    });
  }
  if (window.matchMedia) {
    var mq = matchMedia("(prefers-color-scheme: dark)");
    var onChange = function (e) {
      try { if (localStorage.getItem("theme")) return; } catch (_) {}
      apply(e.matches ? "dark" : "light");
    };
    if (mq.addEventListener) mq.addEventListener("change", onChange);
  }
})();
