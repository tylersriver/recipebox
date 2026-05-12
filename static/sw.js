const VERSION = 'v1';
const STATIC_CACHE = `recipebox-static-${VERSION}`;
const RUNTIME_CACHE = `recipebox-runtime-${VERSION}`;
const OFFLINE_URL = '/static/offline.html';

const PRECACHE_URLS = [
  OFFLINE_URL,
  '/static/style.css',
  '/static/icon.svg',
  '/static/icon-maskable.svg',
  '/static/manifest.webmanifest',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(STATIC_CACHE).then((cache) => cache.addAll(PRECACHE_URLS))
      .then(() => self.skipWaiting())
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((keys) => Promise.all(
        keys
          .filter((key) => key !== STATIC_CACHE && key !== RUNTIME_CACHE)
          .map((key) => caches.delete(key))
      ))
      .then(() => self.clients.claim())
  );
});

function isStaticAsset(url) {
  return url.pathname.startsWith('/static/');
}

function isSameOrigin(request) {
  return new URL(request.url).origin === self.location.origin;
}

self.addEventListener('fetch', (event) => {
  const request = event.request;

  if (request.method !== 'GET' || !isSameOrigin(request)) {
    return;
  }

  const url = new URL(request.url);

  // Don't intercept Datastar SSE streams or search-results live endpoints.
  if (request.headers.get('accept')?.includes('text/event-stream')) {
    return;
  }
  if (url.pathname === '/search/results' || url.pathname === '/import/submit') {
    return;
  }

  if (request.mode === 'navigate') {
    event.respondWith(networkFirstNavigation(request));
    return;
  }

  if (isStaticAsset(url)) {
    event.respondWith(cacheFirst(request, STATIC_CACHE));
    return;
  }
});

async function networkFirstNavigation(request) {
  try {
    const response = await fetch(request);
    if (response && response.ok) {
      const cache = await caches.open(RUNTIME_CACHE);
      cache.put(request, response.clone()).catch(() => {});
    }
    return response;
  } catch (err) {
    const cached = await caches.match(request);
    if (cached) return cached;
    const offline = await caches.match(OFFLINE_URL);
    return offline || new Response('Offline', { status: 503, statusText: 'Offline' });
  }
}

async function cacheFirst(request, cacheName) {
  const cached = await caches.match(request);
  if (cached) return cached;
  try {
    const response = await fetch(request);
    if (response && response.ok) {
      const cache = await caches.open(cacheName);
      cache.put(request, response.clone()).catch(() => {});
    }
    return response;
  } catch (err) {
    return new Response('Offline', { status: 503, statusText: 'Offline' });
  }
}
