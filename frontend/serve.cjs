/* Optional zero-dependency local preview; not a production server. */
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');

const root = __dirname;
const host = process.env.FRONTEND_HOST || '127.0.0.1';
const port = parseInt(process.env.FRONTEND_PORT || '4173', 10);
const proxyTarget = process.env.API_PROXY_TARGET || '';

const types = {
  '.html': 'text/html; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.woff2': 'font/woff2',
  '.md': 'text/plain; charset=utf-8',
};

function shouldProxy(pathname) {
  return pathname.startsWith('/api/') || pathname === '/healthz' || pathname === '/readyz';
}

function proxyRequest(req, res, pathname, search) {
  const target = new URL(pathname + search, proxyTarget);
  const headers = { ...req.headers, host: target.host };
  delete headers.connection;

  const proxyReq = http.request(
    {
      hostname: target.hostname,
      port: target.port,
      path: target.pathname + target.search,
      method: req.method,
      headers,
    },
    (proxyRes) => {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.pipe(res);
    },
  );

  proxyReq.on('error', () => {
    res.writeHead(502, { 'Content-Type': 'text/plain; charset=utf-8' });
    res.end('Bad gateway: backend unreachable');
  });

  req.pipe(proxyReq);
}

function serveStatic(req, res, pathname) {
  const file = path.resolve(root, '.' + (pathname === '/' ? '/index.html' : pathname));
  if (!file.startsWith(root + path.sep)) {
    res.writeHead(403);
    return res.end('Forbidden');
  }
  fs.readFile(file, (err, body) => {
    if (err) {
      res.writeHead(404);
      return res.end('Not found');
    }
    res.writeHead(200, {
      'Content-Type': types[path.extname(file)] || 'application/octet-stream',
      'Cache-Control': 'no-store',
    });
    res.end(body);
  });
}

http
  .createServer((req, res) => {
    let parsed;
    try {
      parsed = new URL(req.url, `http://${host}`);
    } catch {
      res.writeHead(400);
      return res.end('Bad request');
    }

    const pathname = decodeURIComponent(parsed.pathname);
    if (proxyTarget && shouldProxy(pathname)) {
      return proxyRequest(req, res, pathname, parsed.search);
    }
    return serveStatic(req, res, pathname);
  })
  .listen(port, host, () => {
    const proxyNote = proxyTarget ? ` — API proxied to ${proxyTarget}` : '';
    console.log(`Preview: http://${host}:${port}${proxyNote}`);
  });
