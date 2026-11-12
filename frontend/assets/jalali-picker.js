/* Jalali date picker — stores Gregorian ISO (YYYY-MM-DD, Asia/Tehran day) for the API. */
(() => {
  'use strict';
  const fa = (v) => String(v).replace(/\d/g, (x) => '۰۱۲۳۴۵۶۷۸۹'[x]);
  const pad2 = (n) => String(n).padStart(2, '0');
  const WEEK = ['ش', 'ی', 'د', 'س', 'چ', 'پ', 'ج'];
  let openPicker = null;

  function emitChange(root, iso) {
    root.dispatchEvent(new CustomEvent('jalali-change', { bubbles: true, detail: { iso } }));
  }

  function setValue(root, iso) {
    const hidden = root.querySelector('input[type="hidden"]');
    const display = root.querySelector('.jalali-picker-display');
    if (!hidden || !display) return;
    hidden.value = iso;
    const j = Jalali.jalaaliFromIso(iso);
    if (j) display.value = fa(`${j.jy}/${pad2(j.jm)}/${pad2(j.jd)}`);
    else display.value = '';
    emitChange(root, iso);
  }

  function closePopover() {
    if (openPicker) {
      const pop = openPicker.querySelector('.jalali-picker-popover');
      if (pop) pop.hidden = true;
      openPicker = null;
    }
  }

  function buildGrid(root) {
    const grid = root.querySelector('.jalali-picker-grid');
    if (!grid) return;
    const viewJy = root._viewJy;
    const viewJm = root._viewJm;
    const selectedIso = root.querySelector('input[type="hidden"]')?.value || '';
    const firstIso = Jalali.isoFromJalaali(viewJy, viewJm, 1);
    const dow = new Date(`${firstIso}T12:00:00+03:30`).getDay();
    const offset = (dow + 1) % 7;
    const len = Jalali.jalaaliMonthLength(viewJy, viewJm);
    let html = WEEK.map((w) => `<span class="jalali-dow">${w}</span>`).join('');
    for (let i = 0; i < offset; i++) html += '<span class="jalali-day jalali-day--pad"></span>';
    for (let d = 1; d <= len; d++) {
      const iso = Jalali.isoFromJalaali(viewJy, viewJm, d);
      const sel = iso === selectedIso ? ' jalali-day--selected' : '';
      html += `<button type="button" class="jalali-day${sel}" data-iso="${iso}">${fa(d)}</button>`;
    }
    grid.innerHTML = html;
  }

  function refreshMonth(root) {
    const title = root.querySelector('.jalali-picker-title');
    if (title) title.textContent = `${Jalali.monthNames[root._viewJm - 1]} ${fa(root._viewJy)}`;
    buildGrid(root);
  }

  function openPopover(root) {
    closePopover();
    const pop = root.querySelector('.jalali-picker-popover');
    if (!pop) return;
    pop.hidden = false;
    openPicker = root;
    const hidden = root.querySelector('input[type="hidden"]');
    const iso = hidden?.value || Jalali.tehranTodayIso();
    let j = Jalali.jalaaliFromIso(iso) || Jalali.jalaaliFromIso(Jalali.tehranTodayIso());
    root._viewJy = j.jy;
    root._viewJm = j.jm;
    refreshMonth(root);
  }

  function mount(host) {
    if (!host || host.dataset.jalaliMounted) return host;
    const name = host.dataset.jalaliDate || host.getAttribute('name') || 'date';
    const required = host.hasAttribute('data-required');
    const initial = host.dataset.value || Jalali.tehranTodayIso();
    const id = host.id || '';
    const root = document.createElement('div');
    root.className = 'jalali-date-picker';
    root.innerHTML = `<div class="jalali-picker-input-wrap"><input type="text" class="jalali-picker-display" readonly placeholder="انتخاب تاریخ" autocomplete="off" aria-haspopup="dialog"><button type="button" class="jalali-picker-toggle" aria-label="باز کردن تقویم">📅</button><input type="hidden" name="${name}" value="${initial}" ${required ? 'required' : ''}></div><div class="jalali-picker-popover" hidden role="dialog" aria-label="تقویم شمسی"><div class="jalali-picker-head"><button type="button" class="jalali-nav" data-nav="prev" aria-label="ماه قبل">‹</button><span class="jalali-picker-title"></span><button type="button" class="jalali-nav" data-nav="next" aria-label="ماه بعد">›</button></div><div class="jalali-picker-grid"></div></div>`;
    host.replaceWith(root);
    if (id) root.id = id;
    setValue(root, initial);
    root.querySelector('.jalali-picker-display')?.addEventListener('click', () => openPopover(root));
    root.querySelector('.jalali-picker-toggle')?.addEventListener('click', () => openPopover(root));
    root.addEventListener('click', (e) => {
      const nav = e.target.closest('[data-nav]');
      if (nav) {
        e.preventDefault();
        if (nav.dataset.nav === 'prev') {
          root._viewJm -= 1;
          if (root._viewJm < 1) {
            root._viewJm = 12;
            root._viewJy -= 1;
          }
        } else {
          root._viewJm += 1;
          if (root._viewJm > 12) {
            root._viewJm = 1;
            root._viewJy += 1;
          }
        }
        refreshMonth(root);
        return;
      }
      const day = e.target.closest('[data-iso]');
      if (day) {
        setValue(root, day.dataset.iso);
        closePopover();
      }
    });
    root.dataset.jalaliMounted = '1';
    return root;
  }

  document.addEventListener('click', (e) => {
    if (!openPicker) return;
    if (openPicker.contains(e.target)) return;
    closePopover();
  });
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closePopover();
  });

  window.JalaliPicker = {
    mount,
    upgradeAll(root = document) {
      [...root.querySelectorAll('[data-jalali-date]:not([data-jalali-mounted])')].forEach((el) => mount(el));
    },
    getIso(host) {
      const picker = host?.classList?.contains('jalali-date-picker')
        ? host
        : host?.closest?.('.jalali-date-picker') || host;
      return picker?.querySelector?.('input[type="hidden"]')?.value || '';
    },
    setIso(host, iso) {
      const picker = host?.classList?.contains('jalali-date-picker')
        ? host
        : host?.closest?.('.jalali-date-picker') || host;
      if (picker) setValue(picker, iso);
    },
    setDisabled(host, disabled) {
      const picker = host?.classList?.contains('jalali-date-picker')
        ? host
        : host?.closest?.('.jalali-date-picker') || host;
      if (!picker) return;
      picker.querySelectorAll('input,button').forEach((el) => {
        el.disabled = !!disabled;
      });
    },
  };
})();
