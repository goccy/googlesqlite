import { existsSync, readFileSync } from 'node:fs'
import type { IncomingMessage, ServerResponse } from 'node:http'
import { fileURLToPath } from 'node:url'
import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'

const wasmGzFile = fileURLToPath(
  new URL('./public/googlesqlite.wasm.gz', import.meta.url),
)

// rawGzPlugin serves the compressed engine binary (*.wasm.gz) as an
// opaque binary, without a Content-Encoding header. Vite would
// otherwise treat a .gz file as transport-compressed and the browser
// would decompress it transparently — but a plain static host (and
// GitHub Pages) serves a .gz file as-is. Forcing the opaque form keeps
// the dev server, the preview server and the deploy consistent: the
// worker always receives the raw gzip bytes and decompresses them
// itself.
function rawGzPlugin(): Plugin {
  const middleware = (
    req: IncomingMessage,
    res: ServerResponse,
    next: () => void,
  ): void => {
    const path = (req.url ?? '').split('?')[0]
    if (!path.endsWith('googlesqlite.wasm.gz') || !existsSync(wasmGzFile)) {
      next()
      return
    }
    const body = readFileSync(wasmGzFile)
    res.setHeader('Content-Type', 'application/octet-stream')
    res.setHeader('Content-Length', String(body.length))
    res.end(body)
  }
  return {
    name: 'raw-gz',
    configureServer(server) {
      server.middlewares.use(middleware)
    },
    configurePreviewServer(server) {
      server.middlewares.use(middleware)
    },
  }
}

// The site is published under https://goccy.github.io/googlesqlite/,
// so every asset URL is prefixed with /googlesqlite/. The dev server
// serves under the same base, i.e. http://localhost:5173/googlesqlite/.
export default defineConfig({
  base: '/googlesqlite/',
  plugins: [react(), rawGzPlugin()],
  worker: {
    format: 'es',
  },
  server: {
    // Bind to every IPv4 interface so the dev server is reachable from
    // other devices on the local network. Plain `host: true` binds the
    // IPv6 wildcard, which is not reachable over IPv4 on every host.
    host: '0.0.0.0',
  },
})
