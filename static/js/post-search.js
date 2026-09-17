(function() {
  var input = document.getElementById('post-search');
  var list = document.getElementById('post-list');
  var noResults = document.getElementById('no-results');
  if (!input || !list) return;
  var articles = list.querySelectorAll('article');
  input.addEventListener('input', function() {
    var q = input.value.toLowerCase().trim();
    var visible = 0;
    for (var i = 0; i < articles.length; i++) {
      var title = articles[i].textContent.toLowerCase();
      var show = !q || title.indexOf(q) !== -1;
      articles[i].style.display = show ? '' : 'none';
      if (show) visible++;
    }
    if (noResults) noResults.style.display = (q && visible === 0) ? '' : 'none';
  });
})();
