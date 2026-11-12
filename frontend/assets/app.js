(() => {
  'use strict';
  const $ = (s, root = document) => root.querySelector(s);
  const $$ = (s, root = document) => [...root.querySelectorAll(s)];
  const esc = (v) =>
    String(v ?? '').replace(/[&<>"']/g, (x) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[x]));
  const fa = (v) => String(v).replace(/\d/g, (x) => '۰۱۲۳۴۵۶۷۸۹'[x]);
  const days = ['شنبه', 'یکشنبه', 'دوشنبه', 'سه‌شنبه', 'چهارشنبه', 'پنجشنبه', 'جمعه'];
  const today = () => (typeof Jalali !== 'undefined' ? Jalali.tehranTodayIso() : new Date().toISOString().slice(0, 10));
  const tehranDateFormatter = new Intl.DateTimeFormat('fa-IR', { dateStyle: 'medium', timeZone: 'Asia/Tehran' });
  const tehranWeekdayFormatter = new Intl.DateTimeFormat('en-US', { weekday: 'short', timeZone: 'Asia/Tehran' });
  const tehranWeekdayMap = { Sat: 0, Sun: 1, Mon: 2, Tue: 3, Wed: 4, Thu: 5, Fri: 6 };
  const prettyDate = (v) => {
    try {
      return tehranDateFormatter.format(new Date(String(v) + 'T12:00:00+03:30'));
    } catch {
      return v;
    }
  };
  const tehranWeekday = (dateStr) => {
    try {
      const key = tehranWeekdayFormatter.format(new Date(String(dateStr) + 'T12:00:00+03:30'));
      return tehranWeekdayMap[key] ?? 0;
    } catch {
      return 0;
    }
  };
  const prettyTime = (v) => fa(String(v || '').slice(0, 5));
  const CLOCK_HOURS = Array.from({ length: 24 }, (_, i) => String(i).padStart(2, '0'));
  const CLOCK_MINUTES = Array.from({ length: 60 }, (_, i) => String(i).padStart(2, '0'));

  function parseClock(value) {
    const v = normalizeDigits(String(value || '').trim());
    const m = v.match(/^(\d{1,2}):(\d{2})$/);
    if (!m) return { h: '08', m: '00' };
    const h = Number(m[1]);
    const min = Number(m[2]);
    if (h > 23 || min > 59) return { h: '08', m: '00' };
    return { h: String(h).padStart(2, '0'), m: String(min).padStart(2, '0') };
  }

  function clockSelectPair(name, value) {
    const { h, m } = parseClock(value);
    const hourOpts = CLOCK_HOURS.map(
      (x) => `<option value="${x}" ${x === h ? 'selected' : ''}>${fa(x)}</option>`
    ).join('');
    const minOpts = CLOCK_MINUTES.map(
      (x) => `<option value="${x}" ${x === m ? 'selected' : ''}>${fa(x)}</option>`
    ).join('');
    return `<span class="clock-24" data-clock="${name}"><select name="${name}Hour" class="clock-hour" aria-label="ساعت" required>${hourOpts}</select><span class="clock-sep" aria-hidden="true">:</span><select name="${name}Minute" class="clock-minute" aria-label="دقیقه" required>${minOpts}</select></span>`;
  }

  function readClockFromForm(form, name) {
    const h = form.querySelector(`[name="${name}Hour"]`)?.value;
    const m = form.querySelector(`[name="${name}Minute"]`)?.value;
    return `${h || '00'}:${m || '00'}`;
  }
  const BRAND = 'کلاسینو';
  const MODULE_SOON = 'این بخش هنوز به API وصل نشده است.';
  const ic = (name, cls = 'icon') => (typeof Icons !== 'undefined' ? Icons.svg(name, cls) : '');
  const mountIcons = (root) => {
    if (typeof Icons !== 'undefined') Icons.mount(root || document);
  };
  const panelLink = (href, text, cls = 'btn btn-green') =>
    `<a href="${href}" class="${cls} btn-with-icon"><span>${text}</span>${ic('arrow-left', 'icon')}</a>`;
  const searchFieldIcon = () => `<span data-icon="search" aria-hidden="true"></span>`;
  const btnNew = (text) => `${ic('plus', 'icon')}<span>${text}</span>`;
  let classes = [],
    sessionsOverviewCount = 0,
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
  let studentsPageClassId = '',
    studentsPageCache = [],
    studentsClassesCache = [],
    studentsSearch = '',
    studentsListPage = 1;
  const STUDENTS_PAGE_SIZE = 10;
  let sessionsPageClassId = '',
    sessionsPageCache = [],
    sessionsClassesCache = [],
    sessionsSearch = '',
    sessionsListPage = 1;
  const SESSIONS_PAGE_SIZE = 10;
  const ATTENDANCE_STATUSES = [
    ['present', 'حاضر'],
    ['absent', 'غایب'],
    ['late', 'تأخیر'],
    ['excused', 'موجه'],
  ];
  const faStudentCollator = new Intl.Collator('fa', { sensitivity: 'base', numeric: true });
  const studentsScopeSchoolId = () => API.activeSchool() || school?.id || schools[0]?.id || '';
  const sortStudentsPersian = (list) =>
    [...list].sort((a, b) => {
      const byLast = faStudentCollator.compare(a.lastName || '', b.lastName || '');
      if (byLast !== 0) return byLast;
      return faStudentCollator.compare(a.firstName || '', b.firstName || '');
    });
  const classBy = (id) => classes.find((x) => String(x.id) === String(id));
  const studentDisplayName = (s) =>
    String(s?.name || '').trim() || [s?.firstName, s?.lastName].filter(Boolean).join(' ').trim();
  const classOptions = () => classes.map((x) => `<option value="${esc(x.id)}">${esc(x.title)}</option>`).join('');
  const gradeOptions = (selected = '') =>
    ['7', '8', '9', '10', '11', '12']
      .map((g) => `<option value="${g}" ${String(selected) === g ? 'selected' : ''}>پایه ${fa(g)}</option>`)
      .join('');
  const noClassesNotice = () =>
    `<div class="notice">هنوز کلاسی ثبت نشده است.</div><p class="help">از بخش «کلاس‌ها» برای یکی از مدارس خود کلاس بسازید.</p>${panelLink('#panel/classes', 'مدیریت کلاس‌ها')}`;
  const dayOptions = () => days.map((x, i) => `<option value="${i}">${x}</option>`).join('');
  const error = (e) => `<div class="error-message" role="alert">${esc(e.message || 'خطایی رخ داد.')}</div>`;
  const moduleSoon = () => `<p class="help">${MODULE_SOON}</p>`;

  function hideModalAlert() {
    const el = $('#modal-inline-alert', $('#modal-content'));
    if (el) el.hidden = true;
  }

  function toast(message) {
    clearTimeout(toastTimer);
    const modal = $('#modal');
    const pageToast = $('#toast');
    if (modal?.open) {
      if (pageToast) pageToast.hidden = true;
      const content = $('#modal-content');
      let alert = $('#modal-inline-alert', content);
      if (!alert && content) {
        alert = document.createElement('div');
        alert.id = 'modal-inline-alert';
        alert.className = 'error-message modal-inline-alert';
        alert.setAttribute('role', 'alert');
        const actions = content.querySelector('.quick-links');
        if (actions) actions.before(alert);
        else content.appendChild(alert);
      }
      if (alert) {
        alert.textContent = message;
        alert.hidden = false;
        alert.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
      }
      toastTimer = setTimeout(hideModalAlert, 4500);
      return;
    }
    hideModalAlert();
    if (pageToast) {
      pageToast.textContent = message;
      pageToast.hidden = false;
      toastTimer = setTimeout(() => {
        pageToast.hidden = true;
      }, 4500);
    }
  }
  function showModal(html) {
    currentModalOp++;
    $('#modal-content').innerHTML = html;
    upgradeJalaliFields($('#modal-content'));
    if (!$('#modal').open) $('#modal').showModal();
    return currentModalOp;
  }
  function closeModal() {
    currentModalOp++;
    hideModalAlert();
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

  async function fetchStudentsForFilter(schoolId, classId) {
    studentsClassesCache = await fetchClassesForFilter(schoolId);
    const scope = classId
      ? studentsClassesCache.filter((c) => String(c.id) === String(classId))
      : studentsClassesCache;
    if (!scope.length) return [];
    const batches = await Promise.all(
      scope.map(async (c) => {
        const list = await API.students(c.id);
        return list.map((s) => ({
          ...s,
          classTitle: c.title,
          classGrade: c.grade,
          schoolName: c.schoolName || '',
        }));
      })
    );
    return sortStudentsPersian(batches.flat());
  }

  const studentClassFilterOptions = (selectedClassId) =>
    `<option value="" ${selectedClassId === '' || selectedClassId === undefined ? 'selected' : ''}>همه کلاس‌ها</option>${studentsClassesCache
      .map(
        (c) =>
          `<option value="${esc(c.id)}" ${String(c.id) === String(selectedClassId) ? 'selected' : ''}>${esc(c.title)}</option>`
      )
      .join('')}`;

  function filteredStudentsRows() {
    const q = studentsSearch.trim();
    const rows = q
      ? studentsPageCache.filter((s) => {
          const name = studentDisplayName(s);
          return String(name).includes(q) || String(s.firstName || '').includes(q) || String(s.lastName || '').includes(q);
        })
      : studentsPageCache;
    return sortStudentsPersian(rows);
  }

  function paintStudentsTable() {
    const host = $('#students-table-host');
    const pager = $('#students-pagination');
    if (!host || !pager) return;
    const rows = filteredStudentsRows();
    const total = rows.length;
    const totalPages = Math.max(1, Math.ceil(total / STUDENTS_PAGE_SIZE));
    if (studentsListPage > totalPages) studentsListPage = totalPages;
    if (studentsListPage < 1) studentsListPage = 1;
    const start = (studentsListPage - 1) * STUDENTS_PAGE_SIZE;
    const pageRows = rows.slice(start, start + STUDENTS_PAGE_SIZE);
    host.innerHTML = pageRows.length
      ? `<div class="table-wrap"><table class="classes-table students-table"><thead><tr><th>ردیف</th><th>کلاس</th><th>نام</th><th>نام خانوادگی</th><th>کد ملی</th><th>نام پدر</th><th>شماره</th><th>عملیات</th></tr></thead><tbody>${pageRows
          .map((s, i) => {
            const nc = s.nationalCode ? fa(s.nationalCode) : '—';
            const father = s.fatherName ? esc(s.fatherName) : '—';
            const num = s.studentNumber ? fa(s.studentNumber) : '—';
            return `<tr data-student-id="${esc(s.id)}"><td>${fa(start + i + 1)}</td><td>${esc(s.classTitle || '')}</td><td>${esc(s.firstName || '')}</td><td>${esc(s.lastName || '')}</td><td>${nc}</td><td>${father}</td><td>${num}</td><td><div class="table-actions"><button type="button" class="btn btn-outline btn-small" data-edit-student="${esc(s.id)}">ویرایش</button><button type="button" class="danger-link" data-delete-student="${esc(s.id)}">حذف</button></div></td></tr>`;
          })
          .join('')}</tbody></table></div>`
      : `<p class="empty">${total ? 'نتیجه‌ای برای فیلترها یافت نشد.' : 'دانش‌آموزی ثبت نشده است.'}</p>`;
    pager.innerHTML =
      total > 0
        ? `<span class="page-meta">${fa(total)} دانش‌آموز • صفحه ${fa(studentsListPage)} از ${fa(totalPages)}</span><button type="button" class="btn btn-outline btn-small" id="students-page-prev" ${studentsListPage <= 1 ? 'disabled' : ''}>قبلی</button>${Array.from({ length: totalPages }, (_, i) => i + 1)
            .filter((p) => p === 1 || p === totalPages || Math.abs(p - studentsListPage) <= 1)
            .reduce((acc, p, idx, arr) => {
              if (idx && p - arr[idx - 1] > 1) acc.push('…');
              acc.push(p);
              return acc;
            }, [])
            .map((p) =>
              typeof p === 'number'
                ? `<button type="button" class="btn btn-outline btn-small students-page-num ${p === studentsListPage ? 'active' : ''}" data-page="${p}">${fa(p)}</button>`
                : `<span class="page-ellipsis">${p}</span>`
            )
            .join('')}<button type="button" class="btn btn-outline btn-small" id="students-page-next" ${studentsListPage >= totalPages ? 'disabled' : ''}>بعدی</button>`
        : '';
    bindStudentsTableActions();
    $('#students-page-prev')?.addEventListener('click', () => {
      if (studentsListPage > 1) {
        studentsListPage--;
        paintStudentsTable();
      }
    });
    $('#students-page-next')?.addEventListener('click', () => {
      if (studentsListPage < totalPages) {
        studentsListPage++;
        paintStudentsTable();
      }
    });
    $$('.students-page-num', pager).forEach((b) =>
      b.addEventListener('click', () => {
        studentsListPage = Number(b.dataset.page) || 1;
        paintStudentsTable();
      })
    );
  }

  function bindStudentsTableActions() {
    const reload = async () => {
      const sid = studentsScopeSchoolId();
      studentsPageCache = await fetchStudentsForFilter(sid, studentsPageClassId);
      classesPageCache = await fetchClassesForFilter(sid);
      classes = classesPageCache;
      paintStudentsTable();
    };
    $$('[data-edit-student]').forEach((b) => {
      b.onclick = () => openStudentEditor(studentsPageCache.find((x) => String(x.id) === b.dataset.editStudent), reload);
    });
    $$('[data-delete-student]').forEach((b) => {
      b.onclick = async () => {
        const s = studentsPageCache.find((x) => String(x.id) === b.dataset.deleteStudent);
        if (!s || !confirm(`«${studentDisplayName(s)}» حذف شود؟`)) return;
        b.disabled = true;
        try {
          await API.deleteStudent(s.id);
          await reload();
          toast('دانش‌آموز حذف شد.');
        } catch (err) {
          toast(err.message);
          b.disabled = false;
        }
      };
    });
  }

  function openStudentEditor(studentData, onSuccess) {
    if (!studentData?.id) return;
    const op = showModal(
      `<h2 id="modal-title">ویرایش دانش‌آموز</h2><form id="student-edit-form" class="form-stack">${studentFormFields(studentData)}<div class="quick-links"><button type="submit" class="btn btn-green">ذخیره</button><button type="button" class="btn btn-outline modal-close">انصراف</button></div></form>`
    );
    const form = $('#student-edit-form');
    form.onsubmit = async (e) => {
      e.preventDefault();
      const body = readStudentBody(form);
      if (!nationalCodeClientOk(body.nationalCode)) {
        toast('کد ملی باید ۱۰ رقم باشد یا خالی بماند.');
        return;
      }
      const submit = $('button[type="submit"]', form);
      submit.disabled = true;
      try {
        await API.saveStudent({ ...body, id: studentData.id });
        if (op !== currentModalOp) return;
        closeModal();
        await onSuccess?.();
        toast('دانش‌آموز به‌روز شد.');
      } catch (err) {
        toast(err.message);
      } finally {
        submit.disabled = false;
      }
    };
    $$('.modal-close', form).forEach((b) => (b.onclick = closeModal));
  }

  function openStudentCreateModal() {
    const defaultClass = studentsPageClassId || studentsClassesCache[0]?.id || '';
    const op = showModal(
      `<h2 id="modal-title">دانش‌آموز جدید</h2><p>کلاس و مشخصات دانش‌آموز را وارد کنید.</p><form id="student-create-form" class="form-stack"><label>کلاس<select name="classId" required>${studentClassFilterOptions(defaultClass)}</select></label>${studentFormFields()}<div class="quick-links"><button type="submit" class="btn btn-green">ثبت دانش‌آموز</button><button type="button" class="btn btn-outline modal-close">انصراف</button></div></form>`
    );
    const form = $('#student-create-form');
    form.onsubmit = async (e) => {
      e.preventDefault();
      const classId = form.querySelector('[name="classId"]').value;
      const body = readStudentBody(form);
      if (!nationalCodeClientOk(body.nationalCode)) {
        toast('کد ملی باید ۱۰ رقم باشد یا خالی بماند.');
        return;
      }
      const btn = $('button[type="submit"]', form);
      btn.disabled = true;
      try {
        await API.createStudent(body, classId);
        if (op !== currentModalOp) return;
        closeModal();
        const sid = studentsScopeSchoolId();
        studentsPageCache = await fetchStudentsForFilter(sid, studentsPageClassId);
        classesPageCache = await fetchClassesForFilter(sid);
        classes = classesPageCache;
        studentsListPage = 1;
        paintStudentsTable();
        toast('دانش‌آموز ثبت شد.');
      } catch (err) {
        toast(err.message);
      } finally {
        btn.disabled = false;
      }
    };
    $$('.modal-close', form).forEach((b) => (b.onclick = closeModal));
  }

  const sessionsScopeSchoolId = () => API.activeSchool() || school?.id || schools[0]?.id || '';

  const normalizeSessionRow = (s) => ({
    ...s,
    startTime: s.startTime || s.start || '',
    endTime: s.endTime || s.end || '',
  });

  async function fetchSessionsForFilter(schoolId, classId) {
    const sid = schoolId || sessionsScopeSchoolId();
    sessionsClassesCache = await fetchClassesForFilter(schoolId);
    if (!sessionsClassesCache.length) return [];
    const byClass = Object.fromEntries(sessionsClassesCache.map((c) => [String(c.id), c]));
    const enrich = (list) =>
      list.map((raw) => {
        const s = normalizeSessionRow(raw);
        const c = byClass[String(s.classId)] || {};
        return { ...s, classTitle: c.title || '', schoolName: c.schoolName || '' };
      });
    if (classId) {
      return enrich(await API.sessions(sid, classId)).sort(
        (a, b) =>
          String(b.sessionDate).localeCompare(String(a.sessionDate)) || String(a.startTime).localeCompare(String(b.startTime))
      );
    }
    return enrich(await API.sessions(sid, '')).sort(
      (a, b) =>
        String(b.sessionDate).localeCompare(String(a.sessionDate)) || String(a.startTime).localeCompare(String(b.startTime))
    );
  }

  const sessionClassFilterOptions = (selectedClassId) =>
    `<option value="" ${selectedClassId === '' || selectedClassId === undefined ? 'selected' : ''}>همه کلاس‌ها</option>${sessionsClassesCache
      .map(
        (c) =>
          `<option value="${esc(c.id)}" ${String(c.id) === String(selectedClassId) ? 'selected' : ''}>${esc(c.title)}</option>`
      )
      .join('')}`;

  function sessionFormFields(s = {}) {
    const dateVal = s.sessionDate || today();
    const dayVal = s.day !== undefined && s.day !== '' ? s.day : tehranWeekday(dateVal);
    return `<label>کلاس<select name="classId" ${s.id ? 'disabled' : ''} required>${sessionClassFilterOptions(s.classId || sessionsPageClassId || '')}</select></label><label>تاریخ جلسه<span data-jalali-date="sessionDate" data-value="${esc(dateVal)}" data-required></span></label><label>روز هفته<select name="day">${days
      .map((x, i) => `<option value="${i}" ${Number(dayVal) === i ? 'selected' : ''}>${x}</option>`)
      .join('')}</select></label><label>از ساعت (۲۴ ساعته)${clockSelectPair('startTime', s.startTime || '08:00')}</label><label>تا ساعت (۲۴ ساعته)${clockSelectPair('endTime', s.endTime || '09:30')}</label><label>موضوع (اختیاری)<input name="subject" value="${esc(s.subject || '')}" maxlength="100" placeholder="مثلاً حل تمرین"></label>`;
  }

  function readSessionBody(form) {
    const d = new FormData(form);
    const sessionDate = String(d.get('sessionDate') || '').trim();
    return {
      sessionDate,
      day: Number(d.get('day')),
      startTime: readClockFromForm(form, 'startTime'),
      endTime: readClockFromForm(form, 'endTime'),
      subject: String(d.get('subject') || '').trim(),
    };
  }

  function bindSessionDateDaySync(form) {
    const daySelect = form.querySelector('[name="day"]');
    const picker = form.querySelector('.jalali-date-picker');
    if (!daySelect || !picker) return;
    const sync = (iso) => {
      if (iso) daySelect.value = String(tehranWeekday(iso));
    };
    picker.addEventListener('jalali-change', (e) => sync(e.detail?.iso));
    sync(JalaliPicker.getIso(picker));
  }

  function upgradeJalaliFields(root) {
    mountIcons(root);
    if (typeof JalaliPicker !== 'undefined') JalaliPicker.upgradeAll(root);
  }

  function filteredSessionsRows() {
    const q = sessionsSearch.trim();
    const rows = q
      ? sessionsPageCache.filter((s) => {
          const subj = s.subject || '';
          return String(s.classTitle || '').includes(q) || subj.includes(q) || prettyDate(s.sessionDate).includes(q);
        })
      : sessionsPageCache;
    return rows;
  }

  function paintSessionsTable() {
    const host = $('#sessions-table-host');
    const pager = $('#sessions-pagination');
    if (!host || !pager) return;
    const rows = filteredSessionsRows();
    const total = rows.length;
    const totalPages = Math.max(1, Math.ceil(total / SESSIONS_PAGE_SIZE));
    if (sessionsListPage > totalPages) sessionsListPage = totalPages;
    if (sessionsListPage < 1) sessionsListPage = 1;
    const start = (sessionsListPage - 1) * SESSIONS_PAGE_SIZE;
    const pageRows = rows.slice(start, start + SESSIONS_PAGE_SIZE);
    host.innerHTML = pageRows.length
      ? `<div class="table-wrap"><table class="classes-table sessions-table"><thead><tr><th>ردیف</th><th>کلاس</th><th>تاریخ</th><th>روز</th><th>ساعت</th><th>موضوع</th><th>عملیات</th></tr></thead><tbody>${pageRows
          .map((s, i) => {
            const subj = s.subject ? esc(s.subject) : '—';
            return `<tr data-session-id="${esc(s.id)}"><td>${fa(start + i + 1)}</td><td>${esc(s.classTitle || '')}</td><td>${esc(prettyDate(s.sessionDate))}</td><td>${days[s.day] || '—'}</td><td><span title="Asia/Tehran">${prettyTime(s.startTime)} – ${prettyTime(s.endTime)}</span></td><td>${subj}</td><td><div class="table-actions"><button type="button" class="btn btn-outline btn-small" data-attend-session="${esc(s.id)}">حضور و غیاب</button><button type="button" class="btn btn-outline btn-small" data-edit-session="${esc(s.id)}">ویرایش</button><button type="button" class="danger-link" data-delete-session="${esc(s.id)}">حذف</button></div></td></tr>`;
          })
          .join('')}</tbody></table></div>`
      : `<p class="empty">${total ? 'نتیجه‌ای برای فیلترها یافت نشد.' : 'جلسه‌ای ثبت نشده است.'}</p>`;
    pager.innerHTML =
      total > 0
        ? `<span class="page-meta">${fa(total)} جلسه • صفحه ${fa(sessionsListPage)} از ${fa(totalPages)}</span><button type="button" class="btn btn-outline btn-small" id="sessions-page-prev" ${sessionsListPage <= 1 ? 'disabled' : ''}>قبلی</button>${Array.from({ length: totalPages }, (_, i) => i + 1)
            .filter((p) => p === 1 || p === totalPages || Math.abs(p - sessionsListPage) <= 1)
            .reduce((acc, p, idx, arr) => {
              if (idx && p - arr[idx - 1] > 1) acc.push('…');
              acc.push(p);
              return acc;
            }, [])
            .map((p) =>
              typeof p === 'number'
                ? `<button type="button" class="btn btn-outline btn-small sessions-page-num ${p === sessionsListPage ? 'active' : ''}" data-page="${p}">${fa(p)}</button>`
                : `<span class="page-ellipsis">${p}</span>`
            )
            .join('')}<button type="button" class="btn btn-outline btn-small" id="sessions-page-next" ${sessionsListPage >= totalPages ? 'disabled' : ''}>بعدی</button>`
        : '';
    bindSessionsTableActions();
    $('#sessions-page-prev')?.addEventListener('click', () => {
      if (sessionsListPage > 1) {
        sessionsListPage--;
        paintSessionsTable();
      }
    });
    $('#sessions-page-next')?.addEventListener('click', () => {
      if (sessionsListPage < totalPages) {
        sessionsListPage++;
        paintSessionsTable();
      }
    });
    $$('.sessions-page-num').forEach((b) =>
      b.addEventListener('click', () => {
        sessionsListPage = Number(b.dataset.page);
        paintSessionsTable();
      })
    );
  }

  function bindSessionsTableActions() {
    const reload = async () => {
      const sid = sessionsScopeSchoolId();
      sessionsPageCache = await fetchSessionsForFilter(sid, sessionsPageClassId);
      $('#session-filter-class') && ($('#session-filter-class').innerHTML = sessionClassFilterOptions(sessionsPageClassId));
      paintSessionsTable();
    };
    $$('[data-attend-session]').forEach((b) => {
      b.onclick = () => {
        sessionStorage.setItem('classino.attendanceSessionId', b.dataset.attendSession);
        location.hash = 'panel/attendance';
      };
    });
    $$('[data-edit-session]').forEach((b) => {
      b.onclick = () => openSessionEditor(sessionsPageCache.find((x) => String(x.id) === b.dataset.editSession), reload);
    });
    $$('[data-delete-session]').forEach((b) =>
      b.addEventListener('click', async () => {
        const item = sessionsPageCache.find((x) => String(x.id) === b.dataset.deleteSession);
        if (!item || !confirm('این جلسه از برنامه حذف شود؟')) return;
        b.disabled = true;
        try {
          await API.deleteSession(item.id);
          sessionsPageCache = sessionsPageCache.filter((x) => String(x.id) !== String(item.id));
          sessionsOverviewCount = Math.max(0, sessionsOverviewCount - 1);
          paintSessionsTable();
          toast('جلسه حذف شد.');
        } catch (err) {
          toast(err.message);
          b.disabled = false;
        }
      })
    );
  }

  function openSessionEditor(sessionData, onSuccess) {
    if (!sessionData) return;
    const op = showModal(
      `<h2 id="modal-title">ویرایش جلسه</h2><form id="session-edit-form" class="form-stack">${sessionFormFields({ ...sessionData, classId: sessionData.classId })}<div class="quick-links"><button type="submit" class="btn btn-green">ذخیره</button><button type="button" class="btn btn-outline modal-close">انصراف</button></div></form>`
    );
    const form = $('#session-edit-form');
    upgradeJalaliFields(form);
    bindSessionDateDaySync(form);
    form.onsubmit = async (e) => {
      e.preventDefault();
      const body = readSessionBody(form);
      if (body.startTime >= body.endTime) {
        toast('ساعت پایان باید بعد از ساعت شروع باشد.');
        return;
      }
      if (tehranWeekday(body.sessionDate) !== body.day) {
        toast('روز هفته با تاریخ جلسه (تقویم تهران) هم‌خوان نیست.');
        return;
      }
      const btn = $('button[type="submit"]', form);
      btn.disabled = true;
      try {
        await API.saveSession({ ...body, id: sessionData.id });
        if (op !== currentModalOp) return;
        closeModal();
        await onSuccess();
        toast('جلسه به‌روز شد.');
      } catch (err) {
        toast(err.message);
      } finally {
        btn.disabled = false;
      }
    };
    $$('.modal-close', form).forEach((b) => (b.onclick = closeModal));
  }

  function openSessionCreateModal() {
    const defaultClass = sessionsPageClassId || sessionsClassesCache[0]?.id || '';
    const op = showModal(
      `<h2 id="modal-title">جلسه جدید</h2><p>کلاس، تاریخ و بازهٔ زمانی جلسه را وارد کنید. زمان‌ها بر اساس منطقهٔ زمانی تهران هستند.</p><form id="session-create-form" class="form-stack">${sessionFormFields({ classId: defaultClass, sessionDate: today() })}<div class="quick-links"><button type="submit" class="btn btn-green">ثبت جلسه</button><button type="button" class="btn btn-outline modal-close">انصراف</button></div></form>`
    );
    const form = $('#session-create-form');
    upgradeJalaliFields(form);
    bindSessionDateDaySync(form);
    form.onsubmit = async (e) => {
      e.preventDefault();
      const classId = form.querySelector('[name="classId"]').value;
      const body = readSessionBody(form);
      if (body.startTime >= body.endTime) {
        toast('ساعت پایان باید بعد از ساعت شروع باشد.');
        return;
      }
      if (tehranWeekday(body.sessionDate) !== body.day) {
        toast('روز هفته با تاریخ جلسه (تقویم تهران) هم‌خوان نیست.');
        return;
      }
      const btn = $('button[type="submit"]', form);
      btn.disabled = true;
      try {
        await API.createSession(body, classId);
        if (op !== currentModalOp) return;
        closeModal();
        const sid = sessionsScopeSchoolId();
        sessionsPageCache = await fetchSessionsForFilter(sid, sessionsPageClassId);
        sessionsListPage = 1;
        sessionsOverviewCount += 1;
        paintSessionsTable();
        toast('جلسه ثبت شد.');
      } catch (err) {
        toast(err.message);
      } finally {
        btn.disabled = false;
      }
    };
    $$('.modal-close', form).forEach((b) => (b.onclick = closeModal));
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
          return `<tr data-class-id="${esc(c.id)}"><td>${fa(start + i + 1)}</td>${showSchool ? `<td>${esc(c.schoolName || '')}</td>` : ''}<td>${esc(c.title)}</td><td>${fa(c.grade)}</td><td>${room}</td><td>${fa(c.studentCount ?? 0)}</td><td><div class="table-actions"><button type="button" class="btn btn-outline btn-small" data-manage-students="${esc(c.id)}">دانش‌آموزان</button><button type="button" class="btn btn-outline btn-small" data-edit-class="${esc(c.id)}">ویرایش</button><button type="button" class="danger-link" data-delete-class="${esc(c.id)}">حذف</button></div></td></tr>`;
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
    $$('[data-manage-students]').forEach(
      (b) =>
        (b.onclick = () => {
          const c = classBy(b.dataset.manageStudents);
          if (!c) return;
          studentsPageClassId = String(c.id);
          studentsListPage = 1;
          sessionStorage.setItem('classino.selectedClass', studentsPageClassId);
          if (c.schoolId) {
            API.setActiveSchool(c.schoolId);
            sessionStorage.setItem('classino.schoolId', String(c.schoolId));
          }
          location.hash = 'panel/students';
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
      try {
        sessionsOverviewCount = school?.id ? (await API.sessions(school.id, '')).length : 0;
      } catch {
        sessionsOverviewCount = 0;
      }
      return !!school?.id;
    } catch (e) {
      if (/مدرسه‌ای برای حساب|یافت نشد/i.test(e.message)) {
        school = { id: '', name: '' };
        schools = [];
        classes = [];
        sessionsOverviewCount = 0;
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
        btn.innerHTML = `ورود <span data-icon="log-in" data-icon-class="icon-trailing" aria-hidden="true"></span>`;
        mountIcons(btn);
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
    sessionsOverviewCount = 0;
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
    mountIcons($('#login-view'));
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
      students: 'دانش‌آموزان',
      attendance: 'حضور و غیاب',
      discipline: 'انضباط',
      schedule: 'برنامه هفتگی',
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
      if (tab === 'students') await renderStudents();
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
      $('#panel-body').innerHTML = `<div class="notice">هنوز مدرسه‌ای نساخته‌اید. از بخش «نام مدرسه» یک مدرسه بسازید.</div>${panelLink('#panel/settings', 'ساخت مدرسه')}`;
      upgradeJalaliFields($('#panel-body'));
      return;
    }
    const total = classes.reduce((n, c) => n + c.studentCount, 0);
    $('#panel-body').innerHTML = `<div class="stats"><div class="stat"><b>${fa(classes.length)}</b><span>کلاس آموزشی</span></div><div class="stat"><b>${fa(total)}</b><span>عضویت در کلاس‌ها</span></div><div class="stat"><b>${fa(sessionsOverviewCount)}</b><span>جلسه در برنامه هفتگی</span></div></div><div class="surface"><div class="eyebrow">سلام، ${esc(user.name)}</div><h2>یک روز تازه، یک فرصت برای یادگیری</h2><p>کلاس‌ها و برنامهٔ هفتگی مدرسهٔ «${esc(school.name)}» را از منوی کناری مدیریت کنید.</p><div class="quick-links">${panelLink('#panel/classes', 'کلاس‌ها')}<a href="#panel/schedule" class="btn btn-outline btn-with-icon">${ic('calendar-days', 'icon')}<span>برنامه هفتگی</span></a></div></div>`;
    upgradeJalaliFields($('#panel-body'));
  }

  function classFormFields(c = {}) {
    return `<label>عنوان کلاس<input name="title" value="${esc(c.title || '')}" minlength="2" maxlength="100" required placeholder="مثلاً ریاضی ۳"></label><label>پایه<select name="grade" required>${gradeOptions(c.grade || '7')}</select></label><label>کلاس / اتاق<input name="room" value="${esc(c.room || '')}" maxlength="80" placeholder="مثلاً کلاس ۱۰۲"></label>`;
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

  function studentFormFields(s = {}) {
    return `<label>نام<input name="firstName" value="${esc(s.firstName || '')}" minlength="2" maxlength="60" required placeholder="مثلاً علی"></label><label>نام خانوادگی<input name="lastName" value="${esc(s.lastName || '')}" minlength="2" maxlength="60" required placeholder="مثلاً رضایی"></label><label>کد ملی<input name="nationalCode" value="${esc(s.nationalCode || '')}" inputmode="numeric" maxlength="10" placeholder="اختیاری — ۱۰ رقم"></label><label>نام پدر<input name="fatherName" value="${esc(s.fatherName || '')}" maxlength="60" placeholder="اختیاری"></label><label>شماره دانش‌آموزی<input name="studentNumber" value="${esc(s.studentNumber || '')}" maxlength="20" placeholder="اختیاری"></label>`;
  }

  function nationalCodeClientOk(code) {
    const n = normalizeDigits(code);
    if (!n) return true;
    return n.length === 10;
  }

  function normalizeDigits(v) {
    const persian = '۰۱۲۳۴۵۶۷۸۹';
    return String(v || '')
      .replace(/[۰-۹]/g, (d) => String(persian.indexOf(d)))
      .replace(/\D/g, '');
  }

  function readStudentBody(form) {
    const d = new FormData(form);
    return {
      firstName: String(d.get('firstName') || '').trim(),
      lastName: String(d.get('lastName') || '').trim(),
      nationalCode: normalizeDigits(d.get('nationalCode')),
      fatherName: String(d.get('fatherName') || '').trim(),
      studentNumber: String(d.get('studentNumber') || '').trim(),
    };
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
      body.innerHTML = `<div class="notice">برای ثبت کلاس، ابتدا یک مدرسه بسازید.</div>${panelLink('#panel/settings', 'تنظیمات مدرسه')}`;
      upgradeJalaliFields(body);
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
    body.innerHTML = `<div class="surface classes-page"><div class="classes-toolbar"><div class="form-row classes-filters"><label>مدرسه<select id="class-filter-school">${schoolFilterOptions(classesPageSchoolId)}</select></label><label class="classes-search-wrap">جستجوی عنوان<div class="search classes-search">${searchFieldIcon()}<input id="class-search" type="search" placeholder="نام کلاس…" value="${esc(classesSearch)}" autocomplete="off"></div></label></div><button type="button" class="btn btn-green btn-with-icon" id="class-create-open">${btnNew('کلاس جدید')}</button></div><div id="classes-table-host"></div><div id="classes-pagination" class="pagination"></div></div>`;
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
    upgradeJalaliFields(body);
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

  async function renderStudents() {
    const body = $('#panel-body');
    try {
      schools = await API.schools();
    } catch (e) {
      body.innerHTML = error(e);
      return;
    }
    if (!schools.length) {
      body.innerHTML = `<div class="notice">برای ثبت دانش‌آموز، ابتدا یک مدرسه بسازید.</div>${panelLink('#panel/settings', 'تنظیمات مدرسه')}`;
      upgradeJalaliFields(body);
      return;
    }
    const scopeSchoolId = studentsScopeSchoolId();
    if (!scopeSchoolId) {
      body.innerHTML = `<div class="notice">مدرسهٔ فعال مشخص نیست.</div>${panelLink('#panel/settings', 'انتخاب مدرسه')}`;
      upgradeJalaliFields(body);
      return;
    }
    if (
      studentsPageClassId &&
      studentsClassesCache.length &&
      !studentsClassesCache.some((c) => String(c.id) === String(studentsPageClassId))
    ) {
      studentsPageClassId = '';
    }
    try {
      studentsPageCache = await fetchStudentsForFilter(scopeSchoolId, studentsPageClassId);
    } catch (e) {
      body.innerHTML = error(e);
      return;
    }
    body.innerHTML = `<div class="surface classes-page"><div class="classes-toolbar"><div class="form-row classes-filters"><label>کلاس<select id="student-filter-class">${studentClassFilterOptions(studentsPageClassId)}</select></label><label class="classes-search-wrap">جستجوی نام<div class="search classes-search">${searchFieldIcon()}<input id="student-search" type="search" placeholder="نام یا نام خانوادگی…" value="${esc(studentsSearch)}" autocomplete="off"></div></label></div><button type="button" class="btn btn-green btn-with-icon" id="student-create-open">${btnNew('دانش‌آموز جدید')}</button></div><div id="students-table-host"></div><div id="students-pagination" class="pagination"></div></div>`;
    const reloadCache = async () => {
      const sid = studentsScopeSchoolId();
      studentsPageCache = await fetchStudentsForFilter(sid, studentsPageClassId);
      $('#student-filter-class').innerHTML = studentClassFilterOptions(studentsPageClassId);
      classes = await fetchClassesForFilter(sid);
      paintStudentsTable();
    };
    $('#student-filter-class').onchange = async () => {
      studentsPageClassId = $('#student-filter-class').value;
      studentsListPage = 1;
      try {
        await reloadCache();
      } catch (err) {
        toast(err.message);
      }
    };
    $('#student-search').oninput = (e) => {
      studentsSearch = e.target.value;
      studentsListPage = 1;
      paintStudentsTable();
    };
    $('#student-create-open').onclick = openStudentCreateModal;
    upgradeJalaliFields(body);
    paintStudentsTable();
  }

  function attendanceStatusButtons(studentId, active) {
    return `<div class="attendance-picks" data-student-id="${esc(studentId)}">${ATTENDANCE_STATUSES.map(
      ([v, l]) =>
        `<button type="button" class="btn btn-outline btn-small attendance-pick ${v === active ? 'active' : ''}" data-status="${v}">${l}</button>`
    ).join('')}</div>`;
  }

  async function renderAttendance(version) {
    const body = $('#panel-body');
    const sessionId = sessionStorage.getItem('classino.attendanceSessionId') || '';
    if (!school.id) {
      body.innerHTML = `<div class="notice">برای حضور و غیاب ابتدا مدرسه را تنظیم کنید.</div>${panelLink('#panel/settings', 'تنظیمات مدرسه')}`;
      upgradeJalaliFields(body);
      return;
    }
    if (!sessionId) {
      body.innerHTML = `<div class="notice">حضور و غیاب به جلسهٔ برنامه وصل است.</div><p class="help">از جدول «برنامه هفتگی» روی «حضور و غیاب» همان جلسه بزنید.</p>${panelLink('#panel/schedule', 'برنامه هفتگی')}`;
      upgradeJalaliFields(body);
      return;
    }
    body.innerHTML = `<div class="surface attendance-page"><div id="attendance-header" class="help">در حال بارگذاری جلسه…</div><p class="help">با انتخاب هر وضعیت، بلافاصله ذخیره می‌شود.</p><div id="attendance-table-host"></div><div class="quick-links">${panelLink('#panel/schedule', 'بازگشت به برنامه', 'btn btn-outline')}</div></div>`;
    let weeklySession = sessionsPageCache.find((s) => String(s.id) === String(sessionId));
    if (!weeklySession) {
      try {
        weeklySession = await API.weeklySession(sessionId);
      } catch (e) {
        body.innerHTML = error(e) + panelLink('#panel/schedule', 'برنامه هفتگی');
        upgradeJalaliFields(body);
        return;
      }
    }
    if (version !== routeVersion) return;
    upgradeJalaliFields(body);
    const classTitle =
      weeklySession.classTitle || classBy(weeklySession.classId)?.title || classes.find((c) => String(c.id) === String(weeklySession.classId))?.title || '';
    $('#attendance-header').innerHTML = `<strong>${esc(classTitle)}</strong> · ${esc(prettyDate(weeklySession.sessionDate))} · ${prettyTime(weeklySession.startTime)}–${prettyTime(weeklySession.endTime)}${weeklySession.subject ? ` · ${esc(weeklySession.subject)}` : ''}`;
    const statusMap = {};
    const markRow = (studentId, status) => {
      const row = $(`[data-attendance-row="${studentId}"]`);
      if (!row) return;
      $$('.attendance-pick', row).forEach((b) => b.classList.toggle('active', b.dataset.status === status));
    };
    const host = $('#attendance-table-host');
    host.innerHTML = '<p class="help">در حال دریافت فهرست…</p>';
    try {
      const [students, roll] = await Promise.all([API.students(weeklySession.classId), API.attendance(sessionId)]);
      if (version !== routeVersion) return;
      (roll.records || []).forEach((r) => {
        statusMap[r.studentId] = r.status;
      });
      host.innerHTML = students.length
        ? `<div class="table-wrap"><table class="classes-table attendance-table"><thead><tr><th>ردیف</th><th>نام</th><th>نام خانوادگی</th><th>وضعیت</th></tr></thead><tbody>${students
            .map((s, i) => {
              const st = statusMap[s.id] || '';
              return `<tr data-attendance-row="${esc(s.id)}"><td>${fa(i + 1)}</td><td>${esc(s.firstName || '')}</td><td>${esc(s.lastName || '')}</td><td>${attendanceStatusButtons(s.id, st)}</td></tr>`;
            })
            .join('')}</tbody></table></div>`
        : '<p class="empty">دانش‌آموزی در این کلاس ثبت نشده است.</p>';
    } catch (e) {
      if (version === routeVersion) host.innerHTML = error(e);
      return;
    }
    host.addEventListener('click', async (e) => {
      const btn = e.target.closest('.attendance-pick');
      if (!btn || btn.disabled) return;
      const row = btn.closest('[data-attendance-row]');
      const studentId = row?.dataset?.attendanceRow;
      const status = btn.dataset.status;
      if (!studentId || !status) return;
      const picks = $$('.attendance-pick', row);
      picks.forEach((b) => (b.disabled = true));
      try {
        await API.setAttendance(sessionId, studentId, status);
        if (version !== routeVersion) return;
        statusMap[studentId] = status;
        markRow(studentId, status);
      } catch (err) {
        toast(err.message);
      } finally {
        if (version === routeVersion) picks.forEach((b) => (b.disabled = false));
      }
    });
  }

  async function renderDiscipline(version) {
    if (!school.id) {
      $('#panel-body').innerHTML = `<div class="notice">برای ثبت انضباط ابتدا مدرسه را تنظیم کنید.</div>${panelLink('#panel/settings', 'تنظیمات مدرسه')}`;
      upgradeJalaliFields($('#panel-body'));
      return;
    }
    if (!classes.length) {
      $('#panel-body').innerHTML = noClassesNotice();
      upgradeJalaliFields($('#panel-body'));
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
    $('#panel-body').innerHTML = `<div class="surface"><h2>ثبت یادداشت انضباطی</h2><form id="discipline-form"><div class="form-row"><label>کلاس<select id="discipline-class" required>${classOptions()}</select></label><label>دانش‌آموز<select id="discipline-student" required></select></label><label>نوع یادداشت<select name="type"><option value="positive">تشویق</option><option value="attention">نیاز به پیگیری</option></select></label><label>تاریخ<span data-jalali-date="date" data-value="${today()}" data-required></span></label></div><label class="form-stack">شرح یادداشت<textarea name="note" rows="3" minlength="3" maxlength="1000" placeholder="برای مثال: مشارکت خوب در حل تمرین گروهی" required></textarea></label><div class="save-bar"><button class="btn btn-green" id="discipline-save" type="submit" disabled>ثبت یادداشت</button><span id="discipline-form-status" role="status"></span></div></form></div><div class="surface"><h2>یادداشت‌های ثبت‌شده</h2><div id="discipline-records"></div></div>`;
    upgradeJalaliFields($('#panel-body'));
    let studentLoad = 0;
    const loadStudents = async () => {
      const op = ++studentLoad;
      $('#discipline-student').disabled = true;
      $('#discipline-save').disabled = true;
      try {
        const data = await API.students($('#discipline-class').value);
        if (version !== routeVersion || op !== studentLoad) return;
        $('#discipline-student').innerHTML = data.length
          ? data.map((s) => `<option value="${esc(s.id)}">${esc(studentDisplayName(s))}</option>`).join('')
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
    const body = $('#panel-body');
    try {
      schools = await API.schools();
    } catch (e) {
      body.innerHTML = error(e);
      return;
    }
    if (!schools.length) {
      body.innerHTML = `<div class="notice">برای ثبت برنامه، ابتدا یک مدرسه بسازید.</div>${panelLink('#panel/settings', 'تنظیمات مدرسه')}`;
      upgradeJalaliFields(body);
      return;
    }
    const scopeSchoolId = sessionsScopeSchoolId();
    if (!scopeSchoolId) {
      body.innerHTML = `<div class="notice">مدرسهٔ فعال مشخص نیست.</div>${panelLink('#panel/settings', 'انتخاب مدرسه')}`;
      upgradeJalaliFields(body);
      return;
    }
    if (
      sessionsPageClassId &&
      sessionsClassesCache.length &&
      !sessionsClassesCache.some((c) => String(c.id) === String(sessionsPageClassId))
    ) {
      sessionsPageClassId = '';
    }
    try {
      sessionsPageCache = await fetchSessionsForFilter(scopeSchoolId, sessionsPageClassId);
    } catch (e) {
      body.innerHTML = error(e);
      return;
    }
    if (version !== routeVersion) return;
    if (!sessionsClassesCache.length) {
      body.innerHTML = noClassesNotice();
      upgradeJalaliFields(body);
      return;
    }
    body.innerHTML = `<div class="surface classes-page"><div class="classes-toolbar"><div class="form-row classes-filters"><label>کلاس<select id="session-filter-class">${sessionClassFilterOptions(sessionsPageClassId)}</select></label><label class="classes-search-wrap">جستجو<div class="search classes-search">${searchFieldIcon()}<input id="session-search" type="search" placeholder="کلاس، موضوع یا تاریخ…" value="${esc(sessionsSearch)}" autocomplete="off"></div></label></div><button type="button" class="btn btn-green btn-with-icon" id="session-create-open">${btnNew('جلسه جدید')}</button></div><div id="sessions-table-host"></div><div id="sessions-pagination" class="pagination"></div></div>`;
    const reloadCache = async () => {
      const sid = sessionsScopeSchoolId();
      sessionsPageCache = await fetchSessionsForFilter(sid, sessionsPageClassId);
      const sel = $('#session-filter-class');
      if (sel) sel.innerHTML = sessionClassFilterOptions(sessionsPageClassId);
      try {
        sessionsOverviewCount = (await API.sessions(sid, '')).length;
      } catch {
        /* keep previous count */
      }
      paintSessionsTable();
    };
    $('#session-filter-class').onchange = async () => {
      sessionsPageClassId = $('#session-filter-class').value;
      sessionsListPage = 1;
      try {
        await reloadCache();
      } catch (err) {
        toast(err.message);
      }
    };
    $('#session-search').oninput = (e) => {
      sessionsSearch = e.target.value;
      sessionsListPage = 1;
      paintSessionsTable();
    };
    $('#session-create-open').onclick = openSessionCreateModal;
    upgradeJalaliFields(body);
    paintSessionsTable();
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
    mountIcons(document);
    await refreshSession();
    if (isLoggedIn()) {
      try {
        await loadPanelData();
      } catch { }
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
