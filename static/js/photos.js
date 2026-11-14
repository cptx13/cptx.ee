// Photo viewer: opens a photo over the page, ← → to move, Esc to close.
// Without JS every photo is a plain link to the full image.
(function () {
  var items = [], index = 0, opener = null, box, img, cap, orig;

  function build() {
    box = document.createElement('div');
    box.className = 'ph-box';
    box.setAttribute('role', 'dialog');
    box.setAttribute('aria-modal', 'true');
    box.setAttribute('aria-label', 'Photo viewer');
    box.hidden = true;
    box.innerHTML =
      '<button type="button" class="ph-box-btn ph-box-close" aria-label="Close">×</button>' +
      '<button type="button" class="ph-box-btn ph-box-prev" aria-label="Previous photo">‹</button>' +
      '<figure class="ph-box-fig"><img class="ph-box-img" alt=""><figcaption class="ph-box-cap">' +
      '<span class="ph-box-count"></span><a class="ph-box-orig" target="_blank" rel="noopener">open original ↗</a>' +
      '</figcaption></figure>' +
      '<button type="button" class="ph-box-btn ph-box-next" aria-label="Next photo">›</button>';
    document.body.appendChild(box);
    img = box.querySelector('.ph-box-img');
    cap = box.querySelector('.ph-box-count');
    orig = box.querySelector('.ph-box-orig');
    box.querySelector('.ph-box-close').addEventListener('click', close);
    box.querySelector('.ph-box-prev').addEventListener('click', function () { go(-1); });
    box.querySelector('.ph-box-next').addEventListener('click', function () { go(1); });
    box.addEventListener('click', function (e) { if (e.target === box) close(); });
    var x0 = null;
    box.addEventListener('touchstart', function (e) { x0 = e.touches[0].clientX; }, { passive: true });
    box.addEventListener('touchend', function (e) {
      if (x0 === null) return;
      var dx = e.changedTouches[0].clientX - x0;
      if (Math.abs(dx) > 40) go(dx < 0 ? 1 : -1);
      x0 = null;
    });
  }

  function render() {
    var a = items[index], thumb = a.querySelector('img');
    img.src = a.href;
    img.alt = thumb ? thumb.alt : '';
    cap.textContent = (index + 1) + ' / ' + items.length;
    orig.href = a.href;
    [index - 1, index + 1].forEach(function (i) {
      if (items[i]) { var p = new Image(); p.src = items[i].href; }
    });
  }

  function go(d) {
    index = (index + d + items.length) % items.length;
    render();
  }

  function open(a) {
    if (!box) build();
    items = Array.prototype.slice.call(a.closest('.ph-masonry').querySelectorAll('.ph-item'));
    index = items.indexOf(a);
    opener = a;
    render();
    box.hidden = false;
    document.documentElement.classList.add('ph-lock');
    box.querySelector('.ph-box-close').focus();
  }

  function close() {
    box.hidden = true;
    img.removeAttribute('src');
    document.documentElement.classList.remove('ph-lock');
    if (opener) opener.focus();
  }

  document.addEventListener('click', function (e) {
    var a = e.target.closest && e.target.closest('.ph-item');
    if (!a || e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
    e.preventDefault();
    open(a);
  });

  document.addEventListener('keydown', function (e) {
    if (!box || box.hidden) return;
    if (e.key === 'Escape') close();
    else if (e.key === 'ArrowLeft') go(-1);
    else if (e.key === 'ArrowRight') go(1);
    else if (e.key === 'Tab') {
      var f = box.querySelectorAll('button, a[href]'), first = f[0], last = f[f.length - 1];
      if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
      else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
    }
  });
})();
