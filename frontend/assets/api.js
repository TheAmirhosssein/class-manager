/* Shared API boundary. UI never calls fetch directly. Response contracts: docs/API.md. */
(() => {
  'use strict';
  const C = window.APP_CONFIG;
  const TOKEN_KEY = 'classino.accessToken';
  const token = () => { try { return sessionStorage.getItem(TOKEN_KEY) || ''; } catch { return ''; } };
  const setToken = (t) => { try { if (t) sessionStorage.setItem(TOKEN_KEY, t); else sessionStorage.removeItem(TOKEN_KEY); } catch {} };
  const enc = encodeURIComponent;
  let activeSchoolId = '';
  const schoolQ = () => (activeSchoolId ? '?schoolId=' + enc(activeSchoolId) : '');
  const notImplemented = (name) => () => Promise.reject(new Error(`ماژول ${name} هنوز روی سرور فعال نیست.`));

  async function request(path, { method = 'GET', body } = {}) {
    const headers = { Accept: 'application/json' };
    const access = token();
    if (access) headers.Authorization = 'Bearer ' + access;
    if (body !== undefined) headers['Content-Type'] = 'application/json';
    if (method !== 'GET') {
      const item = document.cookie.split('; ').find((x) => x.startsWith(C.csrfCookie + '='));
      if (item) headers[C.csrfHeader] = decodeURIComponent(item.slice(C.csrfCookie.length + 1));
    }
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 15000);
    try {
      const response = await fetch(C.apiBase.replace(/\/$/, '') + path, {
        method,
        headers,
        credentials: C.credentials,
        body: body === undefined ? undefined : JSON.stringify(body),
        signal: controller.signal,
      });
      const data = response.status === 204 ? null : await response.json().catch(() => null);
      if (!response.ok) {
        throw new Error(
          data?.message ||
            (response.status === 401
              ? 'برای ادامه وارد حساب خود شوید.'
              : response.status === 403
                ? 'اجازه دسترسی به این بخش را ندارید.'
                : 'دریافت یا ذخیره اطلاعات انجام نشد. دوباره تلاش کنید.')
        );
      }
      if (response.status !== 204 && data === null) throw new Error('پاسخ سرور باید JSON معتبر باشد.');
      return data;
    } catch (e) {
      if (e.name === 'AbortError') throw new Error('زمان پاسخ‌گویی سرور تمام شد. دوباره تلاش کنید.');
      if (e instanceof TypeError) throw new Error('ارتباط با سرور برقرار نشد. اتصال و آدرس API را بررسی کنید.');
      throw e;
    } finally {
      clearTimeout(timeout);
    }
  }

  window.API = {
    setActiveSchool(id) {
      activeSchoolId = id ? String(id) : '';
    },
    activeSchool() {
      return activeSchoolId;
    },
    schools: () => request('/schools'),
    school: async () => {
      const data = await request('/school' + schoolQ());
      if (data?.id) activeSchoolId = String(data.id);
      return data;
    },
    createSchool: async (body) => {
      const data = await request('/school', { method: 'POST', body });
      if (data?.id) activeSchoolId = String(data.id);
      return data;
    },
    saveSchool: (body) =>
      request('/school' + schoolQ(), { method: 'PATCH', body: { ...body, id: activeSchoolId || body.id } }),
    classes: () => Promise.resolve([]),
    schedule: () => Promise.resolve([]),
    courses: () => Promise.resolve([]),
    students: () => Promise.resolve([]),
    addSchedule: notImplemented('برنامه کلاسی'),
    removeSchedule: notImplemented('برنامه کلاسی'),
    attendance: notImplemented('حضور و غیاب'),
    saveAttendance: notImplemented('حضور و غیاب'),
    discipline: () => Promise.resolve([]),
    addDiscipline: notImplemented('انضباط'),
    session: () => request('/auth/session'),
    login: async (body) => {
      const data = await request('/auth/login', { method: 'POST', body });
      if (data?.token) setToken(data.token);
      return data;
    },
    register: (body) => request('/auth/register', { method: 'POST', body }),
    logout: async () => {
      try {
        await request('/auth/logout', { method: 'POST', body: {} });
      } finally {
        setToken('');
      }
    },
  };
})();
