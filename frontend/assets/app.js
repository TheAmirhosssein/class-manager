(() => {
  'use strict';
  const $ = (s, root = document) => root.querySelector(s);
  const $$ = (s, root = document) => [...root.querySelectorAll(s)];
  const esc = (v) =>
    String(v ?? '').replace(/[&<>"']/g, (x) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[x]));
  const fa = (v) => String(v).replace(/\d/g, (x) => '۰۱۲۳۴۵۶۷۸۹'[x]);
  const days = ['شنبه', 'یکشنبه', 'دوشنبه', 'سه‌شنبه', 'چهارشنبه', 'پنجشنبه', 'جمعه'];
  const today = () => {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  };
  const prettyDate = (v) => {
    try {
      return new Intl.DateTimeFormat('fa-IR', { dateStyle: 'medium' }).format(new Date(v + 'T12:00:00'));
    } catch {
      return v;
    }
  };
  const BRAND = 'کلاسینو';
  const MODULE_SOON = 'این بخش هنوز به API وصل نشده است.';
  let classes = [],
    schedule = [],
    schools = [],
    school = { id: '', name: '' },
    routeVersion = 0,
    currentModalOp = 0,
    toastTimer,
    user = null,
    classesPageSchoolId = '',
    classesPageCache = [],
    classesSearch = '',
    classesListPage = 1;
  const CLASSES_PAGE_SIZE = 10;
  const classBy = (id) => classes.find((x) => String(x.id) === String(id));
  const classOptions = () => classes.map((x) => `<option value="${esc(x.id)}">${esc(x.title)}</option>`).join('');
  const gradeOptions = (selected = '') =>
    ['10', '11', '12']
      .map((g) => `<option value="${g}" ${String(selected) === g ? 'selected' : ''}>پایه ${fa(g)}</option>`)
      .join('');
  const noClassesNotice = () =>
    `<div class="notice">هنوز کلاسی ثبت نشده است.</div><p class="help">از بخش «کلاس‌ها» برای یکی از مدارس خود کلاس بسازید.</p><a href="#panel/classes" class="btn btn-green">مدیریت کلاس‌ها ←</a>`;
  const dayOptions = () => days.map((x, i) => `<option value="${i}">${x}</option>`).join('');
  const error = (e) => `<div class="error-message" role="alert">${esc(e.message || 'خطایی رخ داد.')}</div>`;
  const moduleSoon = () => `<p class="help">${MODULE_SOON}</p>`;

  function toast(message) {
    $('#toast').textContent = message;
    $('#toast').hidden = false;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => ($('#toast').hidden = true), 4500);
  }
  function showModal(html) {
    currentModalOp++;
    $('#modal-content').innerHTML = html;
    if (!$('#modal').open) $('#modal').showModal();
    return currentModalOp;
  }
  function closeModal() {
    currentModalOp++;
    $('#modal').close();
  }
  $('.modal-close')?.addEventListener('click', closeModal);
  $('#modal')?.addEventListener('click', (e) => {
    if (e.target === $('#modal')) {
      const r = e.target.getBoundingClientRect();
      if (e.clientX < r.left || e.clientX > r.right || e.clientY < r.top || e.clientY > r.bottom) closeModal();
    }
  });
  $('#modal')?.addEventListener('cancel', () => currentModalOp++);

  async function reloadClasses(schoolId) {
    const id = schoolId || classesPageSchoolId || school?.id;
    classes = id ? await API.classes(id) : [];
  }

  const schoolSelectOptions = (selectedId) =>
    schools.map((s) => `<option value="${esc(s.id)}" ${String(s.id) === String(selectedId) ? 'selected' : ''}>${esc(s.name)}</option>`).join('');
  const schoolFilterOptions = (selectedId) =>
    `<option value="" ${selectedId === '' || selectedId === undefined ? 'selected' : ''}>همه مدارس</option>${schoolSelectOptions(selectedId)}`;

  async function fetchClassesForFilter(schoolId) {
    if (schoolId) {
      const name = schools.find((s) => String(s.id) === String(schoolId))?.name || '';
      return (await API.classes(schoolId)).map((c) => ({ ...c, schoolName: name }));
    }
    const batches = await Promise.all(
      schools.map(async (s) => {
        const list = await API.classes(s.id);
        return list.map((c) => ({ ...c, schoolName: s.name }));
      })
    );
    return batches.flat();
  }

  function filteredClassesRows() {
    const q = classesSearch.trim();
    if (!q) return classesPageCache;
    return classesPageCache.filter((c) => String(c.title || '').includes(q));
  }

  function paintClassesTable() {
    const host = $('#classes-table-host');
    const pager = $('#classes-pagination');
    if (!host || !pager) return;
    const rows = filteredClassesRows();
    const total = rows.length;
    const totalPages = Math.max(1, Math.ceil(total / CLASSES_PAGE_SIZE));
    if (classesListPage > totalPages) classesListPage = totalPages;
    if (classesListPage < 1) classesListPage = 1;
    const start = (classesListPage - 1) * CLASSES_PAGE_SIZE;
    const pageRows = rows.slice(start, start + CLASSES_PAGE_SIZE);
    const showSchool = !classesPageSchoolId;
    host.innerHTML = pageRows.length
      ? `<div class="table-wrap"><table class="classes-table"><thead><tr><th>ردیف</th>${showSchool ? '<th>مدرسه</th>' : ''}<th>عنوان</th><th>پایه</th><th>اتاق</th><th>دانش‌آموز</th><th>عملیات</th></tr></thead><tbody>${pageRows
          .map((c, i) => {
            const room = c.room ? esc(c.room) : '—';
            return `<tr data-class-id="${esc(c.id)}"><td>${fa(start + i + 1)}</td>${showSchool ? `<td>${esc(c.schoolName || '')}</td>` : ''}<td>${esc(c.title)}</td><td>${fa(c.grade)}</td><td>${room}</td><td>${fa(c.studentCount ?? 0)}</td><td><div class="table-actions"><button type="button" class="btn btn-outline btn-small" data-attend-class="${esc(c.id)}">حضور</button><button type="button" class="btn btn-outline btn-small" data-edit-class="${esc(c.id)}">ویرایش</button><button type="button" class="danger-link" data-delete-class="${esc(c.id)}">حذف</button></div></td></tr>`;
          })
          .join('')}</tbody></table></div>`
      : `<p class="empty">${total ? 'نتیجه‌ای برای جستجو یافت نشد.' : 'کلاسی ثبت نشده است.'}</p>`;
    pager.innerHTML =
      total > 0
        ? `<span class="page-meta">${fa(total)} کلاس • صفحه ${fa(classesListPage)} از ${fa(totalPages)}</span><button type="button" class="btn btn-outline btn-small" id="classes-page-prev" ${classesListPage <= 1 ? 'disabled' : ''}>قبلی</button>${Array.from({ length: totalPages }, (_, i) => i + 1)
            .filter((p) => p === 1 || p === totalPages || Math.abs(p - classesListPage) <= 1)
            .reduce((acc, p, idx, arr) => {
              if (idx && p - arr[idx - 1] > 1) acc.push('…');
              acc.push(p);
              return acc;
            }, [])
            .map((p) =>
              typeof p === 'number'
                ? `<button type="button" class="btn btn-outline btn-small classes-page-num ${p === classesListPage ? 'active' : ''}" data-page="${p}">${fa(p)}</button>`
                : `<span class="page-ellipsis">${p}</span>`
            )
            .join('')}<button type="button" class="btn btn-outline btn-small" id="classes-page-next" ${classesListPage >= totalPages ? 'disabled' : ''}>بعدی</button>`
        : '';
    bindClassesTableActions();
    $('#classes-page-prev')?.addEventListener('click', () => {
      if (classesListPage > 1) {
        classesListPage--;
        paintClassesTable();
      }
    });
    $('#classes-page-next')?.addEventListener('click', () => {
      if (classesListPage < totalPages) {
        classesListPage++;
        paintClassesTable();
      }
    });
    $$('.classes-page-num', pager).forEach((b) =>
      b.addEventListener('click', () => {
        classesListPage = Number(b.dataset.page) || 1;
        paintClassesTable();
      })
    );
  }

  function bindClassesTableActions() {
    $$('[data-attend-class]').forEach(
      (b) =>
        (b.onclick = () => {
          sessionStorage.setItem('classino.selectedClass', b.dataset.attendClass);
          location.hash = 'panel/attendance';
        })
    );
    $$('[data-edit-class]').forEach((b) => (b.onclick = () => openClassEditor(classBy(b.dataset.editClass))));
    $$('[data-delete-class]').forEach(
      (b) =>
        (b.onclick = async () => {
          const c = classBy(b.dataset.deleteClass);
          if (!c || !confirm(`کلاس «${c.title}» حذف شود؟ دانش‌آموزان و سوابق وابسته هم حذف می‌شوند.`)) return;
          b.disabled = true;
          try {
            await API.deleteClass(c.id);
            classesPageCache = await fetchClassesForFilter(classesPageSchoolId);
            classes = classesPageCache;
            paintClassesTable();
            toast('کلاس حذف شد.');
          } catch (err) {
            toast(err.message);
            b.disabled = false;
          }
        })
    );
  }

  async function loadPanelData() {
    try {
      schools = await API.schools();
      const stored = sessionStorage.getItem('classino.schoolId');
      const pick = schools.find((s) => String(s.id) === stored) || schools[0];
      if (pick?.id) {
        API.setActiveSchool(pick.id);
        sessionStorage.setItem('classino.schoolId', String(pick.id));
        school = await API.school();
      } else {
        school = await API.school();
        if (school?.id) {
          API.setActiveSchool(school.id);
          sessionStorage.setItem('classino.schoolId', String(school.id));
        }
      }
      classes = school?.id ? await API.classes(school.id) : [];
      schedule = [];
      return !!school?.id;
    } catch (e) {
      if (/مدرسه‌ای برای حساب|یافت نشد/i.test(e.message)) {
        school = { id: '', name: '' };
        schools = [];
        classes = [];
        schedule = [];
        API.setActiveSchool('');
        sessionStorage.removeItem('classino.schoolId');
        return false;
      }
      throw e;
    }
  }

  function isLoggedIn() {
    return !!user && user.role === 'teacher';
  }

  function setView(mode) {
    const panel = mode === 'panel';
    document.body.classList.toggle('app-panel', panel);
    document.body.classList.toggle('app-auth', !panel);
    $('#login-view').hidden = panel;
    $('#panel-view').hidden = !panel;
  }

  function updateHeaderAuth() {
    const btn = $('#login-button');
    const brand = $('#brand-link');
    const headerUser = $('#header-user');
    const headerName = $('#header-user-name');
    const sidebarName = $('#sidebar-user-name');
    if (isLoggedIn()) {
      if (headerName) headerName.textContent = user.name;
      if (sidebarName) sidebarName.textContent = user.name;
      if (headerUser) headerUser.hidden = false;
      if (btn) btn.hidden = true;
      if (brand) {
        brand.setAttribute('href', '#panel');
        brand.setAttribute('aria-label', 'کلاسینو، پنل مدیریت');
      }
    } else {
      if (headerUser) headerUser.hidden = true;
      if (btn) {
        btn.hidden = false;
        btn.innerHTML = 'ورود <span aria-hidden="true">↗</span>';
        btn.setAttribute('href', '#login');
      }
      if (sidebarName) sidebarName.textContent = '—';
      if (brand) {
        brand.setAttribute('href', '#login');
        brand.setAttribute('aria-label', 'کلاسینو، صفحه ورود');
      }
    }
  }

  async function refreshSession() {
    try {
      const session = await API.session();
      user = session?.user || null;
    } catch {
      user = null;
    }
  }

  function completeLogin(nextUser) {
    user = nextUser;
    updateHeaderAuth();
    const next = sessionStorage.getItem('classino.afterLogin') || 'panel';
    sessionStorage.removeItem('classino.afterLogin');
    location.hash = next.startsWith('panel') ? next : 'panel';
  }

  async function logout() {
    try {
      await API.logout();
    } catch (e) {
      toast(e.message);
      return;
    }
    user = null;
    API.setActiveSchool('');
    sessionStorage.removeItem('classino.schoolId');
    school = { id: '', name: '' };
    schools = [];
    classes = [];
    schedule = [];
    classesPageSchoolId = '';
    classesPageCache = [];
    classesSearch = '';
    classesListPage = 1;
    updateHeaderAuth();
    location.hash = 'login';
  }

  function renderLoginPage(sub) {
    setView('login');
    document.title = (sub === 'account' ? 'حساب کاربری' : 'ورود') + ' | ' + BRAND;
    const logged = isLoggedIn();
    const account = sub === 'account' && logged;
    $('#login-account').hidden = !account;
    $('#login-forms').hidden = account;
    if (account) {
      $('#account-name').textContent = user.name;
      return;
    }
    $('#login-heading').textContent = 'ورود به کلاسینو';
    $('#auth-status').textContent = '';
    $('#auth-form').reset();
  }

  async function submitAuth(e) {
    e.preventDefault();
    const btn = $('#auth-submit');
    const status = $('#auth-status');
    btn.disabled = true;
    status.textContent = '';
    try {
      const payload = Object.fromEntries(new FormData(e.target));
      const result = await API.login(payload);
      if (!result.user) {
        status.textContent = result.message || 'ورود انجام نشد. دوباره تلاش کنید.';
        return;
      }
      completeLogin(result.user);
    } catch (err) {
      status.textContent = err.message;
    } finally {
      btn.disabled = false;
    }
  }

  async function routePanel() {
    const version = ++routeVersion;
    let hash = location.hash.slice(1);
    if (!hash.startsWith('panel')) hash = 'panel';
    let tab = hash.split('/')[1] || 'overview';
    const titles = {
      overview: 'نمای کلی',
      classes: 'کلاس‌ها',
      attendance: 'حضور و غیاب',
      discipline: 'انضباط',
      schedule: 'برنامه کلاسی',
      settings: 'تنظیمات مدرسه',
    };
    if (!titles[tab]) tab = 'overview';
    setView('panel');
    $('#panel-title').textContent = titles[tab];
    document.title = titles[tab] + ' | ' + BRAND;
    $$('[data-panel]').forEach((x) => x.classList.toggle('active', x.dataset.panel === tab));
    const body = $('#panel-body');
    body.innerHTML = '<p class="help" role="status">در حال دریافت اطلاعات…</p>';
    try {
      await refreshSession();
      if (version !== routeVersion) return;
      if (!isLoggedIn()) {
        sessionStorage.setItem('classino.afterLogin', hash);
        location.replace('#login');
        return;
      }
      updateHeaderAuth();
      await loadPanelData();
      if (version !== routeVersion) return;
      if (tab === 'overview') renderOverview();
      if (tab === 'classes') await renderClasses();
      if (tab === 'attendance') await renderAttendance(version);
      if (tab === 'discipline') await renderDiscipline(version);
      if (tab === 'schedule') await renderSchedule(version);
      if (tab === 'settings') await renderSettings();
    } catch (e) {
      if (version === routeVersion)
        body.innerHTML =
          error(e) +
          '<div class="quick-links"><button class="btn btn-green" id="panel-retry">تلاش دوباره</button><a href="#login" class="btn btn-outline">ورود به حساب</a></div>';
      $('#panel-retry')?.addEventListener('click', route);
    }
  }

  async function route() {
    const hash = location.hash.slice(1) || 'login';
    const parts = hash.split('/');
    if (parts[0] === 'login') {
      if (parts[1] === 'logout') {
        await logout();
        return;
      }
      if (parts[1] === 'register') {
        location.replace('#login');
        return;
      }
      await refreshSession();
      if (isLoggedIn()) {
        location.replace('#panel');
        return;
      }
      renderLoginPage('login');
      updateHeaderAuth();
      return;
    }
    if (parts[0] === 'panel' || hash === 'home') {
      if (hash === 'home') {
        location.replace('#login');
        return;
      }
      await routePanel();
      return;
    }
    location.replace('#login');
  }

  function renderOverview() {
    if (!school.id) {
      $('#panel-body').innerHTML =
        '<div class="notice">هنوز مدرسه‌ای نساخته‌اید. از بخش «نام مدرسه» یک مدرسه بسازید.</div><a href="#panel/settings" class="btn btn-green">ساخت مدرسه ←</a>';
      return;
    }
  const total = classes.reduce((n, c) => n + c.studentCount, 0);
    $('#panel-body').innerHTML = `<div class="stats"><div class="stat"><b>${fa(classes.length)}</b><span>کلاس آموزشی</span></div><div class="stat"><b>${fa(total)}</b><span>عضویت در کلاس‌ها</span></div><div class="stat"><b>${fa(schedule.length)}</b><span>بازه در برنامه هفتگی</span></div></div><div class="surface"><div class="eyebrow">سلام، ${esc(user.name)}</div><h2>یک روز تازه، یک فرصت برای یادگیری</h2><p>کلاس‌های مدرسهٔ «${esc(school.name)}» را از بخش کلاس‌ها مدیریت کنید؛ حضور، برنامه و انضباط به‌زودی به API وصل می‌شوند.</p><div class="quick-links"><a href="#panel/classes" class="btn btn-green">کلاس‌ها ←</a><a href="#panel/settings" class="btn btn-outline">تنظیمات مدرسه</a></div></div>`;
  }

  function classFormFields(c = {}) {
    return `<label>عنوان کلاس<input name="title" value="${esc(c.title || '')}" minlength="2" maxlength="100" required placeholder="مثلاً ریاضی ۳"></label><label>پایه<select name="grade" required>${gradeOptions(c.grade || '10')}</select></label><label>کلاس / اتاق<input name="room" value="${esc(c.room || '')}" maxlength="80" placeholder="مثلاً کلاس ۱۰۲"></label>`;
  }

  function bindClassForm(form, onSuccess) {
    form.onsubmit = async (e) => {
      e.preventDefault();
      const d = new FormData(form);
      const body = {
        title: String(d.get('title') || '').trim(),
        grade: d.get('grade'),
        room: String(d.get('room') || '').trim(),
      };
      if (body.title.length < 2) {
        toast('عنوان کلاس باید حداقل ۲ حرف باشد.');
        return;
      }
      const btn = $('button[type="submit"]', form);
      btn.disabled = true;
      try {
        await onSuccess(body, form);
      } catch (err) {
        toast(err.message);
      } finally {
        btn.disabled = false;
      }
    };
  }

  function openClassEditor(classData) {
    const op = showModal(
      `<h2 id="modal-title">ویرایش کلاس</h2><form id="class-edit-form" class="form-stack">${classFormFields(classData)}<div class="quick-links"><button type="submit" class="btn btn-green">ذخیره</button><button type="button" class="btn btn-outline modal-close">انصراف</button></div></form>`
    );
    const form = $('#class-edit-form');
    bindClassForm(form, async (body) => {
      await API.saveClass({ ...body, id: classData.id });
      if (op !== currentModalOp) return;
      closeModal();
      classesPageCache = await fetchClassesForFilter(classesPageSchoolId);
      classes = classesPageCache;
      paintClassesTable();
      toast('کلاس به‌روز شد.');
    });
    $$('.modal-close', form).forEach((b) => (b.onclick = closeModal));
  }

  function openClassCreateModal() {
    const defaultSchool = classesPageSchoolId || schools[0]?.id || '';
    const op = showModal(
      `<h2 id="modal-title">کلاس جدید</h2><p>مدرسه و مشخصات کلاس را وارد کنید.</p><form id="class-create-form" class="form-stack"><label>مدرسه<select name="schoolId" required>${schoolSelectOptions(defaultSchool)}</select></label>${classFormFields()}<div class="quick-links"><button type="submit" class="btn btn-green">ساخت کلاس</button><button type="button" class="btn btn-outline modal-close">انصراف</button></div></form>`
    );
    const form = $('#class-create-form');
    bindClassForm(form, async (formBody) => {
      const schoolId = form.querySelector('[name="schoolId"]').value;
      await API.createClass(formBody, schoolId);
      if (op !== currentModalOp) return;
      closeModal();
      classesPageCache = await fetchClassesForFilter(classesPageSchoolId);
      classes = classesPageCache;
      classesListPage = 1;
      paintClassesTable();
      toast('کلاس ایجاد شد.');
    });
    $$('.modal-close', form).forEach((b) => (b.onclick = closeModal));
  }

  async function renderClasses() {
    const body = $('#panel-body');
    try {
      schools = await API.schools();
    } catch (e) {
      body.innerHTML = error(e);
      return;
    }
    if (!schools.length) {
      body.innerHTML =
        '<div class="notice">برای ثبت کلاس، ابتدا یک مدرسه بسازید.</div><a href="#panel/settings" class="btn btn-green">تنظیمات مدرسه ←</a>';
      return;
    }
    if (classesPageSchoolId && !schools.some((s) => String(s.id) === String(classesPageSchoolId))) {
      classesPageSchoolId = '';
    }
    try {
      classesPageCache = await fetchClassesForFilter(classesPageSchoolId);
      classes = classesPageCache;
    } catch (e) {
      body.innerHTML = error(e);
      return;
    }
    body.innerHTML = `<div class="surface classes-page"><div class="classes-toolbar"><div class="form-row classes-filters"><label>مدرسه<select id="class-filter-school">${schoolFilterOptions(classesPageSchoolId)}</select></label><label class="classes-search-wrap">جستجوی عنوان<div class="search classes-search"><span aria-hidden="true">⌕</span><input id="class-search" type="search" placeholder="نام کلاس…" value="${esc(classesSearch)}" autocomplete="off"></div></label></div><button type="button" class="btn btn-green" id="class-create-open">کلاس جدید +</button></div><div id="classes-table-host"></div><div id="classes-pagination" class="pagination"></div></div>`;
    $('#class-filter-school').onchange = async () => {
      classesPageSchoolId = $('#class-filter-school').value;
      classesListPage = 1;
      try {
        classesPageCache = await fetchClassesForFilter(classesPageSchoolId);
        classes = classesPageCache;
        paintClassesTable();
      } catch (err) {
        toast(err.message);
      }
    };
    $('#class-search').oninput = (e) => {
      classesSearch = e.target.value;
      classesListPage = 1;
      paintClassesTable();
    };
    $('#class-create-open').onclick = openClassCreateModal;
    paintClassesTable();
  }

  function selectedClass() {
    try {
      const id = sessionStorage.getItem('classino.selectedClass');
      return classBy(id) ? id : classes[0]?.id || '';
    } catch {
      return classes[0]?.id || '';
    }
  }

  async function renderAttendance(version) {
    if (!school.id) {
      $('#panel-body').innerHTML =
        '<div class="notice">برای حضور و غیاب ابتدا مدرسه را تنظیم کنید.</div><a href="#panel/settings" class="btn btn-green">تنظیمات مدرسه ←</a>';
      return;
    }
    if (!classes.length) {
      $('#panel-body').innerHTML = noClassesNotice();
      return;
    }
    $('#panel-body').innerHTML = `<div class="form-row"><label>انتخاب کلاس<select id="attendance-class">${classOptions()}</select></label><label>تاریخ جلسه<input type="date" id="attendance-date" value="${today()}" required></label><button class="btn btn-outline" id="all-present">همه حاضر</button></div><div id="attendance-caption" class="help"></div><div id="attendance-table"></div><div class="save-bar"><button class="btn btn-green" id="attendance-save" disabled>ذخیره حضور و غیاب</button><span id="attendance-status" role="status">وضعیت را برای همه دانش‌آموزان مشخص کنید.</span></div>`;
    $('#attendance-class').value = selectedClass();
    let loadId = 0,
      roster = [],
      savedClass = '',
      savedDate = '',
      dirty = false;
    async function load() {
      const op = ++loadId,
        id = $('#attendance-class').value,
        date = $('#attendance-date').value;
      $('#attendance-save').disabled = true;
      $('#attendance-table').innerHTML = '<p class="help">در حال دریافت فهرست…</p>';
      $('#attendance-status').textContent = '';
      if (!id || !date) {
        $('#attendance-table').innerHTML = '<p class="empty">کلاس و تاریخ معتبر انتخاب کنید.</p>';
        return;
      }
      try {
        const [students, existing] = await Promise.all([API.students(id), API.attendance(id, date)]);
        if (version !== routeVersion || op !== loadId) return;
        roster = students;
        savedClass = id;
        savedDate = date;
        dirty = false;
        $('#attendance-caption').textContent = 'تاریخ جلسه: ' + prettyDate(date);
        $('#attendance-table').innerHTML = students.length
          ? `<div class="table-wrap"><table><thead><tr><th>ردیف</th><th>نام دانش‌آموز</th><th>وضعیت حضور</th></tr></thead><tbody>${students
              .map((s, i) => {
                const status = existing.records.find((x) => x.studentId === s.id)?.status || '';
                return `<tr><td>${fa(i + 1)}</td><td>${esc(s.name)}</td><td><select data-student="${esc(s.id)}" aria-label="وضعیت ${esc(s.name)}"><option value="">انتخاب وضعیت</option>${[
                  ['present', 'حاضر'],
                  ['absent', 'غایب'],
                  ['late', 'تأخیر'],
                  ['excused', 'غیبت موجه'],
                ]
                  .map(([v, l]) => `<option value="${v}" ${v === status ? 'selected' : ''}>${l}</option>`)
                  .join('')}</select></td></tr>`;
              })
              .join('')}</tbody></table></div>`
          : '<p class="empty">دانش‌آموزی در این کلاس ثبت نشده است.</p>';
        $('#attendance-save').disabled = !students.length;
        $('#attendance-status').textContent = existing.records.length
          ? 'اطلاعات ثبت‌شده بارگذاری شد.'
          : 'هنوز برای این جلسه حضوری ثبت نشده است.';
      } catch (e) {
        if (version === routeVersion && op === loadId) $('#attendance-table').innerHTML = error(e);
      }
    }
    $('#attendance-table').onchange = () => {
      dirty = true;
      $('#attendance-status').textContent = 'تغییرات هنوز ذخیره نشده‌اند.';
    };
    let previousClass = $('#attendance-class').value,
      previousDate = $('#attendance-date').value;
    const change = () => {
      if (dirty && !confirm('تغییرات ذخیره نشده‌اند. از تغییر کلاس یا تاریخ مطمئن هستید؟')) {
        $('#attendance-class').value = previousClass;
        $('#attendance-date').value = previousDate;
        return;
      }
      previousClass = $('#attendance-class').value;
      previousDate = $('#attendance-date').value;
      load();
    };
    $('#attendance-class').onchange = change;
    $('#attendance-date').onchange = change;
    $('#all-present').onclick = () => {
      $$('[data-student]').forEach((x) => (x.value = 'present'));
      dirty = true;
      $('#attendance-status').textContent = 'همه حاضر انتخاب شدند؛ برای ثبت، ذخیره را بزنید.';
    };
    $('#attendance-save').onclick = async () => {
      const records = $$('[data-student]').map((x) => ({ studentId: x.dataset.student, status: x.value }));
      if (records.length !== roster.length || records.some((x) => !x.status)) {
        toast('وضعیت همه دانش‌آموزان را مشخص کنید.');
        return;
      }
      const btn = $('#attendance-save');
      btn.disabled = true;
      $('#attendance-class').disabled = true;
      $('#attendance-date').disabled = true;
      $('#all-present').disabled = true;
      try {
        await API.saveAttendance({ classId: savedClass, date: savedDate, records });
        if (version === routeVersion) {
          dirty = false;
          $('#attendance-status').textContent = 'ذخیره شد.';
          toast('حضور و غیاب ذخیره شد.');
        }
      } catch (e) {
        toast(e.message);
      } finally {
        if (version === routeVersion) {
          btn.disabled = false;
          $('#attendance-class').disabled = false;
          $('#attendance-date').disabled = false;
          $('#all-present').disabled = false;
        }
      }
    };
    await load();
  }

  async function renderDiscipline(version) {
    if (!school.id) {
      $('#panel-body').innerHTML =
        '<div class="notice">برای ثبت انضباط ابتدا مدرسه را تنظیم کنید.</div><a href="#panel/settings" class="btn btn-green">تنظیمات مدرسه ←</a>';
      return;
    }
    if (!classes.length) {
      $('#panel-body').innerHTML = noClassesNotice();
      return;
    }
    let records;
    try {
      records = await API.discipline();
    } catch (e) {
      $('#panel-body').innerHTML = error(e);
      return;
    }
    if (version !== routeVersion) return;
    $('#panel-body').innerHTML = `<div class="surface"><h2>ثبت یادداشت انضباطی</h2><form id="discipline-form"><div class="form-row"><label>کلاس<select id="discipline-class" required>${classOptions()}</select></label><label>دانش‌آموز<select id="discipline-student" required></select></label><label>نوع یادداشت<select name="type"><option value="positive">تشویق</option><option value="attention">نیاز به پیگیری</option></select></label><label>تاریخ<input name="date" type="date" value="${today()}" required></label></div><label class="form-stack">شرح یادداشت<textarea name="note" rows="3" minlength="3" maxlength="1000" placeholder="برای مثال: مشارکت خوب در حل تمرین گروهی" required></textarea></label><div class="save-bar"><button class="btn btn-green" id="discipline-save" type="submit" disabled>ثبت یادداشت</button><span id="discipline-form-status" role="status"></span></div></form></div><div class="surface"><h2>یادداشت‌های ثبت‌شده</h2><div id="discipline-records"></div></div>`;
    let studentLoad = 0;
    const loadStudents = async () => {
      const op = ++studentLoad;
      $('#discipline-student').disabled = true;
      $('#discipline-save').disabled = true;
      try {
        const data = await API.students($('#discipline-class').value);
        if (version !== routeVersion || op !== studentLoad) return;
        $('#discipline-student').innerHTML = data.length
          ? data.map((s) => `<option value="${esc(s.id)}">${esc(s.name)}</option>`).join('')
          : '<option value="">بدون دانش‌آموز</option>';
        $('#discipline-student').disabled = false;
        $('#discipline-save').disabled = !data.length;
        $('#discipline-form-status').textContent = '';
      } catch (e) {
        if (version === routeVersion) $('#discipline-form-status').textContent = e.message;
      }
    };
    $('#discipline-class').onchange = loadStudents;
    const renderRecords = () => {
      $('#discipline-records').innerHTML = records.length
        ? records
            .map(
              (r) =>
                `<article class="record"><div class="record-head"><strong>${esc(r.studentName)}</strong><span class="badge">${r.type === 'positive' ? 'تشویق' : 'نیاز به پیگیری'}</span></div><small>${esc(classBy(r.classId)?.title || '')} • ${esc(prettyDate(r.date))}</small><p>${esc(r.note)}</p></article>`
            )
            .join('')
        : '<p class="help">هنوز یادداشتی ثبت نشده است.</p>';
    };
    renderRecords();
    $('#discipline-form').onsubmit = async (e) => {
      e.preventDefault();
      const f = e.target,
        d = new FormData(f),
        note = d.get('note').trim();
      if (note.length < 3) {
        toast('شرح یادداشت باید حداقل ۳ حرف داشته باشد.');
        return;
      }
      const btn = $('#discipline-save');
      btn.disabled = true;
      try {
        const r = await API.addDiscipline({
          classId: $('#discipline-class').value,
          studentId: $('#discipline-student').value,
          studentName: $('#discipline-student').selectedOptions[0].textContent,
          type: d.get('type'),
          date: d.get('date'),
          note,
        });
        if (version === routeVersion) {
          records.unshift(r);
          renderRecords();
          f.elements.note.value = '';
          toast('یادداشت ثبت شد.');
        }
      } catch (e) {
        toast(e.message);
      } finally {
        if (version === routeVersion) btn.disabled = false;
      }
    };
    await loadStudents();
  }

  async function renderSchedule(version) {
    if (!school.id) {
      $('#panel-body').innerHTML =
        '<div class="notice">برای برنامهٔ هفتگی ابتدا مدرسه را تنظیم کنید.</div><a href="#panel/settings" class="btn btn-green">تنظیمات مدرسه ←</a>';
      return;
    }
    if (!classes.length) {
      $('#panel-body').innerHTML = noClassesNotice();
      return;
    }
    try {
      schedule = await API.schedule();
    } catch (e) {
      $('#panel-body').innerHTML = error(e);
      return;
    }
    if (version !== routeVersion) return;
    $('#panel-body').innerHTML = `<div class="surface"><h2>افزودن به برنامه هفتگی</h2><form id="schedule-form"><div class="form-row"><label>کلاس<select name="classId" required>${classOptions()}</select></label><label>روز<select name="day">${dayOptions()}</select></label><label>از ساعت<input name="start" type="time" value="08:00" required></label><label>تا ساعت<input name="end" type="time" value="09:30" required></label><label>موضوع جلسه<input name="topic" maxlength="100" required placeholder="مثلاً حل تمرین تابع"></label><button type="submit" class="btn btn-green">افزودن</button></div></form></div><div id="schedule-table"></div>`;
    const table = () => {
      $('#schedule-table').innerHTML = schedule.length
        ? `<div class="table-wrap"><table><thead><tr><th>روز</th><th>کلاس</th><th>ساعت</th><th>موضوع</th><th>عملیات</th></tr></thead><tbody>${[...schedule]
            .sort((a, b) => a.day - b.day || a.start.localeCompare(b.start))
            .map(
              (s) =>
                `<tr><td>${days[s.day]}</td><td>${esc(classBy(s.classId)?.title || '')}</td><td>${fa(s.start)} تا ${fa(s.end)}</td><td>${esc(s.topic)}</td><td><button class="danger-link" data-remove="${esc(s.id)}" aria-label="حذف جلسه ${esc(s.topic)}">حذف</button></td></tr>`
            )
            .join('')}</tbody></table></div>`
        : '<p class="empty">هنوز برنامه‌ای ثبت نشده است.</p>';
    };
    table();
    $('#schedule-form').onsubmit = async (e) => {
      e.preventDefault();
      const d = new FormData(e.target),
        body = {
          classId: d.get('classId'),
          day: Number(d.get('day')),
          start: d.get('start'),
          end: d.get('end'),
          topic: d.get('topic').trim(),
        };
      if (!body.topic) {
        toast('موضوع جلسه را وارد کنید.');
        return;
      }
      if (body.start >= body.end) {
        toast('ساعت پایان باید بعد از ساعت شروع باشد.');
        return;
      }
      if (schedule.some((s) => s.day === body.day && s.start < body.end && body.start < s.end)) {
        toast('این بازه با یکی از کلاس‌های مدرس تداخل دارد.');
        return;
      }
      const btn = $('button[type="submit"]', e.target);
      btn.disabled = true;
      try {
        const item = await API.addSchedule(body);
        schedule.push(item);
        if (version === routeVersion) {
          table();
          toast('جلسه به برنامه اضافه شد.');
        }
      } catch (e) {
        toast(e.message);
      } finally {
        btn.disabled = false;
      }
    };
    $('#schedule-table').onclick = async (e) => {
      const btn = e.target.closest('[data-remove]');
      if (!btn) return;
      if (!confirm('این جلسه از برنامه هفتگی حذف شود؟')) return;
      btn.disabled = true;
      try {
        await API.removeSchedule(btn.dataset.remove);
        schedule = schedule.filter((x) => x.id !== btn.dataset.remove);
        if (version === routeVersion) {
          table();
          toast('جلسه حذف شد.');
        }
      } catch (e) {
        toast(e.message);
        btn.disabled = false;
      }
    };
  }

  async function renderSettings() {
    try {
      schools = await API.schools();
    } catch (e) {
      $('#panel-body').innerHTML = error(e);
      return;
    }
    const creating = !school.id;
    const list = schools.length
      ? `<div class="surface"><h2>مدارس شما</h2><p>فقط مدرسه‌هایی که خودتان ساخته‌اید در اینجا دیده می‌شوند.</p><ul class="school-list">${schools
          .map(
            (s) =>
              `<li><button type="button" class="btn btn-outline btn-small school-pick" data-school-id="${esc(s.id)}" ${String(s.id) === String(school.id) ? 'aria-current="true"' : ''}>${esc(s.name)}</button></li>`
          )
          .join('')}</ul></div>`
      : '';
    $('#panel-body').innerHTML = `${list}<div class="surface"><h2>${creating ? 'ساخت مدرسه' : 'ویرایش مدرسه فعال'}</h2><p>${creating ? 'با ساخت مدرسه، مالک آن می‌شوید و دیگر کاربران به آن دسترسی ندارند.' : 'نام مدرسهٔ انتخاب‌شده در سایدبار نمایش داده می‌شود.'}</p><form id="school-form" class="form-stack"><label>نام مدرسه<input name="name" value="${esc(school.name || '')}" minlength="2" maxlength="80" required placeholder="مثلاً دبیرستان نمونه"></label><div class="quick-links"><button class="btn btn-green" type="submit">${creating ? 'ساخت مدرسه' : 'ذخیره نام'}</button>${!creating ? '<button type="button" class="btn btn-outline" id="school-create-another">مدرسهٔ دیگر</button>' : ''}</div></form></div>`;
    $$('.school-pick').forEach(
      (b) =>
        (b.onclick = async () => {
          API.setActiveSchool(b.dataset.schoolId);
          sessionStorage.setItem('classino.schoolId', b.dataset.schoolId);
          await loadPanelData();
          await renderSettings();
          toast('مدرسه فعال تغییر کرد.');
        })
    );
    let forceCreate = false;
    $('#school-create-another')?.addEventListener('click', () => {
      forceCreate = true;
      const input = $('input[name="name"]', $('#school-form'));
      input.value = '';
      input.placeholder = 'نام مدرسهٔ جدید';
      input.focus();
      toast('نام مدرسهٔ جدید را وارد کنید و «ساخت مدرسه» بزنید.');
    });
    $('#school-form').onsubmit = async (e) => {
      e.preventDefault();
      const name = new FormData(e.target).get('name').trim();
      if (name.length < 2) {
        toast('نام مدرسه باید حداقل ۲ حرف باشد.');
        return;
      }
      const btn = $('button[type="submit"]', e.target);
      btn.disabled = true;
      const doCreate = creating || forceCreate;
      try {
        if (doCreate) {
          school = await API.createSchool({ name });
          forceCreate = false;
          toast('مدرسه ساخته شد.');
        } else {
          school = await API.saveSchool({ name, id: school.id });
          toast('نام مدرسه ذخیره شد.');
        }
        if (school?.id) {
          API.setActiveSchool(school.id);
          sessionStorage.setItem('classino.schoolId', String(school.id));
        }
        await loadPanelData();
        await renderSettings();
      } catch (err) {
        toast(err.message);
      } finally {
        btn.disabled = false;
      }
    };
  }

  $('#auth-form').addEventListener('submit', submitAuth);
  $('#logout-button')?.addEventListener('click', logout);

  async function init() {
    await refreshSession();
    if (isLoggedIn()) {
      try {
        await loadPanelData();
      } catch {}
    }
    updateHeaderAuth();
    if (!location.hash || location.hash === '#' || location.hash === '#home') {
      location.hash = isLoggedIn() ? 'panel' : 'login';
    }
    window.addEventListener('hashchange', route);
    await route();
  }
  init();
})();
