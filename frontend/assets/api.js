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
    classes: (schoolId) =>
      request('/classes' + (schoolId ? '?schoolId=' + enc(String(schoolId)) : '')),
    class: (classId) => request('/class?classId=' + enc(classId)),
    createClass: (body, schoolId) =>
      request('/class' + (schoolId ? '?schoolId=' + enc(String(schoolId)) : ''), { method: 'POST', body }),
    saveClass: (body) =>
      request('/class?classId=' + enc(body.id || ''), {
        method: 'PATCH',
        body: { title: body.title, grade: body.grade, room: body.room || '', id: body.id },
      }),
    deleteClass: (classId) => request('/class?classId=' + enc(classId), { method: 'DELETE', body: {} }),
    sessions: (schoolId, classId) => {
      const q = [];
      if (classId) q.push('classId=' + enc(String(classId)));
      else if (schoolId) q.push('schoolId=' + enc(String(schoolId)));
      if (!q.length) return Promise.resolve([]);
      return request('/sessions?' + q.join('&'));
    },
    weeklySession: (sessionId) => request('/session?sessionId=' + enc(sessionId)),
    createSession: (body, classId) => {
      const q = ['classId=' + enc(String(classId))];
      if (activeSchoolId) q.push('schoolId=' + enc(activeSchoolId));
      return request('/session?' + q.join('&'), {
        method: 'POST',
        body: {
          sessionDate: body.sessionDate,
          day: body.day,
          startTime: body.startTime,
          endTime: body.endTime,
          subject: body.subject || '',
        },
      });
    },
    saveSession: (body) =>
      request('/session?sessionId=' + enc(body.id || ''), {
        method: 'PATCH',
        body: {
          sessionDate: body.sessionDate,
          day: body.day,
          startTime: body.startTime,
          endTime: body.endTime,
          subject: body.subject || '',
          id: body.id,
        },
      }),
    deleteSession: (sessionId) => request('/session?sessionId=' + enc(sessionId), { method: 'DELETE', body: {} }),
    courses: () => Promise.resolve([]),
    students: (classId) => {
      if (!classId) return Promise.resolve([]);
      return request('/students?classId=' + enc(String(classId)));
    },
    student: (studentId) => request('/student?studentId=' + enc(studentId)),
    createStudent: (body, classId) =>
      request('/student?classId=' + enc(String(classId)), { method: 'POST', body }),
    saveStudent: (body) =>
      request('/student?studentId=' + enc(body.id || ''), {
        method: 'PATCH',
        body: {
          firstName: body.firstName,
          lastName: body.lastName,
          nationalCode: body.nationalCode,
          fatherName: body.fatherName || '',
          studentNumber: body.studentNumber || '',
          id: body.id,
        },
      }),
    deleteStudent: (studentId) => request('/student?studentId=' + enc(studentId), { method: 'DELETE', body: {} }),
    attendance: (sessionId) => request('/attendance?sessionId=' + enc(String(sessionId))),
    setAttendance: (sessionId, studentId, status) =>
      request('/attendance?sessionId=' + enc(String(sessionId)) + '&studentId=' + enc(String(studentId)), {
        method: 'PATCH',
        body: { status },
      }),
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
