/* Optional zero-dependency local preview; not a production server. */
const http=require('node:http');
const fs=require('node:fs');
const path=require('node:path');
const root=__dirname;
const types={'.html':'text/html; charset=utf-8','.css':'text/css; charset=utf-8','.js':'text/javascript; charset=utf-8','.svg':'image/svg+xml','.woff2':'font/woff2','.md':'text/plain; charset=utf-8'};
http.createServer((req,res)=>{
  let decoded;
  try{decoded=decodeURIComponent(new URL(req.url,'http://localhost').pathname);}catch{res.writeHead(400);return res.end('Bad request');}
  const file=path.resolve(root,'.'+(decoded==='/'?'/index.html':decoded));
  if(!file.startsWith(root+path.sep)){res.writeHead(403);return res.end('Forbidden');}
  fs.readFile(file,(err,body)=>{if(err){res.writeHead(404);return res.end('Not found');}res.writeHead(200,{'Content-Type':types[path.extname(file)]||'application/octet-stream','Cache-Control':'no-store'});res.end(body);});
}).listen(4173,'127.0.0.1',()=>console.log('Preview: http://localhost:4173 — press Ctrl+C to stop'));
