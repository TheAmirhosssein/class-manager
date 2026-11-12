/* Only public configuration belongs here. Never place secrets in frontend code. */
window.APP_CONFIG = Object.freeze({
  apiBase: '/api/v1',
  credentials: 'same-origin',
  csrfCookie: 'csrftoken',
  csrfHeader: 'X-CSRFToken',
  // Must match server admin.path_secret / ADMIN_PATH_SECRET
  adminPathSecret: 'sys-ingress',
});
