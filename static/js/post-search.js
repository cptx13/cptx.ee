(function() {
  var toggle = document.getElementById('search-toggle');
  var bar = document.getElementById('search-bar');
  var input = document.getElementById('post-search');
  var list = document.getElementById('post-list');
  var noResults = document.getElementById('no-results');
  if (!toggle || !bar || !input || !list) return;

  var articles = list.querySelectorAll('article');

  toggle.addEventListener('click', function() {
    var open = bar.style.display !== 'none';
    bar.style.display = open ? 'none' : '';
    if (!open) {
      input.focus();
    } else {
      input.value = '';
      input.dispatchEvent(new Event('input'));
    }
  });

  input.addEventListener('input', function() {
    var q = input.value.toLowerCase().trim();
    var visible = 0;
    for (var i = 0; i < articles.length; i++) {
      var text = articles[i].textContent.toLowerCase();
      var show = !q || text.indexOf(q) !== -1;
      articles[i].style.display = show ? '' : 'none';
      if (show) visible++;
    }
    if (noResults) noResults.style.display = (q && visible === 0) ? '' : 'none';
  });

  // Escape key closes search
  input.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
      input.value = '';
      input.dispatchEvent(new Event('input'));
      bar.style.display = 'none';
    }
  });
})();
