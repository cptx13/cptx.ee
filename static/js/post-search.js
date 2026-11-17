// Home: the search field filters the post titles as you type; enter
// submits to /search/ for the full text (the form works without JS too).
(function() {
  var input = document.getElementById('post-search');
  var list = document.getElementById('post-list');
  if (!input || !list) return;
  var noResults = document.getElementById('no-results');
  var full = document.getElementById('search-full');
  var fullLink = document.getElementById('search-full-link');
  var articles = list.querySelectorAll('article');

  function filter() {
    var q = input.value.toLowerCase().trim();
    var visible = 0;
    for (var i = 0; i < articles.length; i++) {
      var show = !q || articles[i].textContent.toLowerCase().indexOf(q) !== -1;
      articles[i].style.display = show ? '' : 'none';
      if (show) visible++;
    }
    if (noResults) noResults.hidden = !(q && visible === 0);
    if (full) full.hidden = !q;
    if (fullLink) fullLink.href = '/search/?q=' + encodeURIComponent(input.value.trim());
  }

  input.addEventListener('input', filter);
  input.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') { input.value = ''; filter(); input.blur(); }
  });

  // "/" focuses the field from anywhere on the page
  document.addEventListener('keydown', function(e) {
    if (e.key !== '/' || e.ctrlKey || e.metaKey || e.altKey) return;
    var t = e.target;
    if (t && (t.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName))) return;
    e.preventDefault();
    input.focus();
  });

  if (input.value) filter();
})();
