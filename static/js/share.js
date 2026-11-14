// "share this view": native share sheet on phones, otherwise copy the link.
// Without JS it is a plain link to the view.
(function () {
  document.addEventListener("click", function (e) {
    var a = e.target.closest && e.target.closest(".share-link");
    if (!a) return;
    var url = a.getAttribute("data-share-url") || a.href;
    if (navigator.share) {
      e.preventDefault();
      navigator.share({ title: document.title, url: url }).catch(function () {});
      return;
    }
    if (navigator.clipboard) {
      e.preventDefault();
      navigator.clipboard.writeText(url).then(function () {
        var t = a.textContent;
        a.textContent = "link copied";
        setTimeout(function () { a.textContent = t; }, 1600);
      });
    }
  });
})();
