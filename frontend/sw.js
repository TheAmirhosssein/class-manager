/* Service worker — cache shell + static assets; API always uses network. */
const CACHE = 'classino-shell-v1';

const PRECACHE_URLS = [
  '/',
  '/index.html',
  '/assets/styles.css',
  '/assets/config.js',
  '/assets/icons.js',
  '/assets/jalali.js',
  '/assets/jalali-picker.js',
  '/assets/api.js',
  '/assets/app.js',
  '/assets/favicon.svg',
  '/manifest.webmanifest',
  '/assets/fonts/YekanBakh-Regular.woff2',
  '/assets/fonts/YekanBakh-SemiBold.woff2',
  '/assets/fonts/YekanBakh-Bold.woff2',
];

function isApiRequest(url) {
  return url.pathname.startsWith('/api/') || url.pathname === '/healthz' || url.pathname === '/readyz';
}

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches
      .open(CACHE)
      .then((cache) => cache.addAll(PRECACHE_URLS))
      .then(() => self.skipWaiting())
      .catch(() => self.skipWaiting()),
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
      .then(() => self.clients.claim()),
  );
});

self.addEventListener('fetch', (event) => {
  if (event.request.method !== 'GET') return;

  const url = new URL(event.request.url);
  if (url.origin !== self.location.origin) return;
  if (isApiRequest(url)) return;

  if (event.request.mode === 'navigate') {
    event.respondWith(
      fetch(event.request)
        .then((response) => {
          if (response.ok) {
            const copy = response.clone();
            caches.open(CACHE).then((cache) => cache.put('/index.html', copy));
          }
          return response;
        })
        .catch(() =>
          caches.match('/index.html').then((cached) => cached || caches.match('/')),
        ),
    );
    return;
  }

  event.respondWith(
    caches.match(event.request).then((cached) => {
      if (cached) return cached;
      return fetch(event.request).then((response) => {
        if (!response || response.status !== 200 || response.type === 'opaque') return response;
        const copy = response.clone();
        caches.open(CACHE).then((cache) => cache.put(event.request, copy));
        return response;
      });
    }),
  );
});
