// wasmfs.ts provides the in-memory filesystem the googlesqlite wasm
// engine runs against in the browser, installed as globalThis.fs
// before the Go runtime starts.
//
// It is a small, self-contained, browser-native implementation (no
// Node Buffer, no dependencies). It needs to be only:
//   - read-only — the engine reads the IANA timezone database from
//     /usr/share/zoneinfo and never writes host files (the SQLite
//     database lives in the OPFS-backed VFS, not here);
//   - asynchronous — Go's js/wasm bridge requires fs callbacks to fire
//     on a later microtask, never synchronously.
// Writes to fd 1 / 2 are routed to console.log / console.error.

// errnoError builds an Error carrying a POSIX-style code, the shape
// the Go js/wasm runtime inspects.
function errnoError(code: string): Error & { code: string } {
  const err = new Error(code) as Error & { code: string }
  err.code = code
  return err
}

type FileNode = { kind: 'file'; ino: number; data: Uint8Array }
type DirNode = { kind: 'dir'; ino: number; children: Map<string, Node> }
type Node = FileNode | DirNode

const EPOCH_MS = 0

let inoSeq = 1

function newDir(): DirNode {
  return { kind: 'dir', ino: inoSeq++, children: new Map() }
}

// splitPath turns an absolute path into its non-empty components,
// resolving "." and "..".
function splitPath(path: string): string[] {
  const parts: string[] = []
  for (const part of path.split('/')) {
    if (part === '' || part === '.') {
      continue
    }
    if (part === '..') {
      parts.pop()
    } else {
      parts.push(part)
    }
  }
  return parts
}

// fileTree is the root of the in-memory filesystem.
const root = newDir()

function mkdirp(path: string): DirNode {
  let dir = root
  for (const name of splitPath(path)) {
    let child = dir.children.get(name)
    if (!child) {
      child = newDir()
      dir.children.set(name, child)
    }
    if (child.kind !== 'dir') {
      throw errnoError('ENOTDIR')
    }
    dir = child
  }
  return dir
}

function addFile(path: string, data: Uint8Array): void {
  const parts = splitPath(path)
  const name = parts.pop()
  if (name === undefined) {
    throw errnoError('EINVAL')
  }
  const dir = mkdirp('/' + parts.join('/'))
  dir.children.set(name, { kind: 'file', ino: inoSeq++, data })
}

function resolve(path: string): Node {
  let node: Node = root
  for (const name of splitPath(path)) {
    if (node.kind !== 'dir') {
      throw errnoError('ENOTDIR')
    }
    const child = node.children.get(name)
    if (!child) {
      throw errnoError('ENOENT')
    }
    node = child
  }
  return node
}

// Stats is the object stat / fstat / lstat resolve with. It carries
// the numeric fields the Go runtime copies into a Stat_t, plus the
// node fs.Stats predicate methods (isDirectory / isFile / …) that the
// wasm fs bridge invokes.
interface Stats {
  dev: number
  ino: number
  mode: number
  nlink: number
  uid: number
  gid: number
  rdev: number
  size: number
  blksize: number
  blocks: number
  atimeMs: number
  mtimeMs: number
  ctimeMs: number
  isDirectory(): boolean
  isFile(): boolean
  isBlockDevice(): boolean
  isCharacterDevice(): boolean
  isSymbolicLink(): boolean
  isFIFO(): boolean
  isSocket(): boolean
}

function stat(node: Node): Stats {
  const isDir = node.kind === 'dir'
  const size = isDir ? 0 : node.data.length
  return {
    dev: 1,
    ino: node.ino,
    // S_IFDIR | 0755 or S_IFREG | 0644.
    mode: isDir ? 0o040000 | 0o755 : 0o100000 | 0o644,
    nlink: 1,
    uid: 0,
    gid: 0,
    rdev: 0,
    size,
    blksize: 4096,
    blocks: Math.ceil(size / 512),
    atimeMs: EPOCH_MS,
    mtimeMs: EPOCH_MS,
    ctimeMs: EPOCH_MS,
    isDirectory: () => isDir,
    isFile: () => !isDir,
    isBlockDevice: () => false,
    isCharacterDevice: () => false,
    isSymbolicLink: () => false,
    isFIFO: () => false,
    isSocket: () => false,
  }
}

// Open file descriptors. Descriptors 0-2 are reserved for the
// standard streams.
type Descriptor = { node: Node; position: number }
const descriptors = new Map<number, Descriptor>()
let fdSeq = 3

// defer invokes cb asynchronously. Go's js/wasm fs bridge deadlocks if
// a callback runs synchronously within the originating call.
function defer(cb: unknown, ...args: unknown[]): void {
  if (typeof cb === 'function') {
    queueMicrotask(() => (cb as (...a: unknown[]) => void)(...args))
  }
}

// outputBuffers accumulate partial stdout / stderr lines.
const outputBuffers: Record<number, string> = { 1: '', 2: '' }
const decoder = new TextDecoder('utf-8')

function writeConsole(fd: number, chunk: Uint8Array): number {
  outputBuffers[fd] += decoder.decode(chunk, { stream: true })
  const nl = outputBuffers[fd].lastIndexOf('\n')
  if (nl !== -1) {
    const line = outputBuffers[fd].slice(0, nl)
    if (fd === 1) {
      console.log(line)
    } else {
      console.error(line)
    }
    outputBuffers[fd] = outputBuffers[fd].slice(nl + 1)
  }
  return chunk.length
}

// wasmFS is the object installed as globalThis.fs. Its method set and
// callback conventions match what the Go js/wasm runtime expects.
const wasmFS = {
  constants: {
    O_WRONLY: 1,
    O_RDWR: 2,
    O_CREAT: 64,
    O_TRUNC: 512,
    O_APPEND: 1024,
    O_EXCL: 128,
    O_DIRECTORY: 65536,
  },

  writeSync(fd: number, buf: Uint8Array): number {
    if (fd === 1 || fd === 2) {
      return writeConsole(fd, buf)
    }
    throw errnoError('EBADF')
  },

  write(
    fd: number,
    buf: Uint8Array,
    offset: number,
    length: number,
    _position: number | null,
    callback: (err: unknown, n?: number) => void,
  ): void {
    if (fd === 1 || fd === 2) {
      if (offset !== 0 || length !== buf.length) {
        defer(callback, errnoError('ENOSYS'))
        return
      }
      const n = writeConsole(fd, buf)
      defer(callback, null, n)
      return
    }
    defer(callback, errnoError('EBADF'))
  },

  open(
    path: string,
    _flags: number,
    _mode: number,
    callback: (err: unknown, fd?: number) => void,
  ): void {
    let node: Node
    try {
      node = resolve(path)
    } catch (e) {
      defer(callback, e)
      return
    }
    const fd = fdSeq++
    descriptors.set(fd, { node, position: 0 })
    defer(callback, null, fd)
  },

  close(fd: number, callback: (err: unknown) => void): void {
    descriptors.delete(fd)
    defer(callback, null)
  },

  fsync(_fd: number, callback: (err: unknown) => void): void {
    defer(callback, null)
  },

  read(
    fd: number,
    buffer: Uint8Array,
    offset: number,
    length: number,
    position: number | null,
    callback: (err: unknown, n?: number) => void,
  ): void {
    const desc = descriptors.get(fd)
    if (!desc || desc.node.kind !== 'file') {
      defer(callback, errnoError('EBADF'))
      return
    }
    const start = position === null ? desc.position : position
    const data = desc.node.data
    const end = Math.min(start + length, data.length)
    const n = Math.max(0, end - start)
    if (n > 0) {
      buffer.set(data.subarray(start, end), offset)
    }
    if (position === null) {
      desc.position = start + n
    }
    defer(callback, null, n)
  },

  fstat(
    fd: number,
    callback: (err: unknown, stats?: Stats) => void,
  ): void {
    const desc = descriptors.get(fd)
    if (!desc) {
      defer(callback, errnoError('EBADF'))
      return
    }
    defer(callback, null, stat(desc.node))
  },

  stat(
    path: string,
    callback: (err: unknown, stats?: Stats) => void,
  ): void {
    try {
      defer(callback, null, stat(resolve(path)))
    } catch (e) {
      defer(callback, e)
    }
  },

  lstat(
    path: string,
    callback: (err: unknown, stats?: Stats) => void,
  ): void {
    wasmFS.stat(path, callback)
  },

  readdir(
    path: string,
    callback: (err: unknown, names?: string[]) => void,
  ): void {
    let node: Node
    try {
      node = resolve(path)
    } catch (e) {
      defer(callback, e)
      return
    }
    if (node.kind !== 'dir') {
      defer(callback, errnoError('ENOTDIR'))
      return
    }
    defer(callback, null, [...node.children.keys()])
  },

  // Mutating and symlink operations are unsupported: the engine never
  // writes host files. They report ENOSYS like the wasm_exec.js stub.
  mkdir(_p: string, _m: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  rmdir(_p: string, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  unlink(_p: string, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  rename(_a: string, _b: string, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  truncate(_p: string, _l: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  ftruncate(_fd: number, _l: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  readlink(_p: string, cb: (e: unknown) => void): void {
    defer(cb, errnoError('EINVAL'))
  },
  symlink(_t: string, _p: string, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  link(_a: string, _b: string, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  chmod(_p: string, _m: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  fchmod(_fd: number, _m: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  chown(_p: string, _u: number, _g: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  fchown(_fd: number, _u: number, _g: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  lchown(_p: string, _u: number, _g: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
  utimes(_p: string, _a: number, _m: number, cb: (e: unknown) => void): void {
    defer(cb, errnoError('ENOSYS'))
  },
}

function base64ToBytes(b64: string): Uint8Array {
  const binary = atob(b64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

// installWasmFS fetches the timezone database, populates the in-memory
// filesystem and installs it as globalThis.fs.
export async function installWasmFS(baseURL: string): Promise<void> {
  mkdirp('/tmp')
  mkdirp('/usr/share/zoneinfo')

  const resp = await fetch(`${baseURL}zoneinfo.json`)
  if (!resp.ok) {
    throw new Error(`failed to load zoneinfo.json: HTTP ${resp.status}`)
  }
  const files = (await resp.json()) as Record<string, string>
  for (const [rel, b64] of Object.entries(files)) {
    addFile(`/usr/share/zoneinfo/${rel}`, base64ToBytes(b64))
  }

  ;(globalThis as { fs?: unknown }).fs = wasmFS
}
