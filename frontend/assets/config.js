/* Only public configuration belongs here. Never place secrets in frontend code. */
window.APP_CONFIG = Object.freeze({
  apiBase: '/api/v1',
  credentials: 'same-origin',
  csrfCookie: 'csrftoken',
  csrfHeader: 'X-CSRFToken',
});
