(() => {
  'use strict';
  if (!('serviceWorker' in navigator)) return;

  window.addEventListener('load', () => {
    navigator.serviceWorker
      .register('/sw.js', { scope: '/' })
      .catch(() => {
        /* Optional preview / insecure context */
      });
  });
})();
