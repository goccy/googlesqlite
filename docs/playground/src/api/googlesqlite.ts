// googlesqlite.ts is the UI-thread client for the worker. It owns the
// Worker instance, correlates requests with responses by id, and
// exposes a small promise-based API. A single shared client is
// exported so the worker (and the heavy wasm load it triggers) starts
// as soon as the module is imported.

import type {
  ExecResponse,
  ExportFormat,
  LoadProgress,
  QueryResult,
  WorkerRequest,
  WorkerResponse,
} from './types'

type Pending = {
  resolve: (value: unknown) => void
  reject: (reason: Error) => void
}

// DistributiveOmit applies Omit to each member of a union separately.
// A plain Omit over a union collapses to the members' shared keys,
// which would drop the per-request payload fields (sql, format).
type DistributiveOmit<T, K extends keyof never> = T extends unknown
  ? Omit<T, K>
  : never

export class GoogleSQLiteClient {
  private readonly worker: Worker
  private seq = 0
  private readonly pending = new Map<number, Pending>()
  private resolveReady!: () => void
  private rejectReady!: (reason: Error) => void
  private progressListener: ((p: LoadProgress) => void) | null = null

  // ready resolves once the wasm engine has finished initialising.
  readonly ready: Promise<void>

  constructor() {
    this.ready = new Promise<void>((resolve, reject) => {
      this.resolveReady = resolve
      this.rejectReady = reject
    })
    this.worker = new Worker(new URL('./worker.ts', import.meta.url), {
      type: 'module',
    })
    this.worker.addEventListener(
      'message',
      (event: MessageEvent<WorkerResponse>) => this.onMessage(event.data),
    )
  }

  // onProgress registers a listener for engine-initialisation progress.
  onProgress(listener: (p: LoadProgress) => void): void {
    this.progressListener = listener
  }

  private onMessage(msg: WorkerResponse): void {
    switch (msg.type) {
      case 'ready':
        this.resolveReady()
        return
      case 'init-error':
        this.rejectReady(new Error(msg.error))
        return
      case 'progress':
        this.progressListener?.({
          phase: msg.phase,
          loaded: msg.loaded,
          total: msg.total,
        })
        return
      case 'result': {
        const pending = this.pending.get(msg.id)
        if (!pending) {
          return
        }
        this.pending.delete(msg.id)
        if (msg.ok) {
          pending.resolve(msg.data)
        } else {
          pending.reject(new Error(msg.error))
        }
        return
      }
    }
  }

  private call<T>(req: DistributiveOmit<WorkerRequest, 'id'>): Promise<T> {
    const id = ++this.seq
    return new Promise<T>((resolve, reject) => {
      this.pending.set(id, {
        resolve: resolve as (value: unknown) => void,
        reject,
      })
      this.worker.postMessage({ ...req, id })
    })
  }

  // exec runs one or more statements and returns their results.
  exec(sql: string): Promise<ExecResponse> {
    return this.call<ExecResponse>({ kind: 'exec', sql })
  }

  // getHistory returns every executed statement recorded so far.
  getHistory(): Promise<QueryResult[]> {
    return this.call<QueryResult[]>({ kind: 'getHistory' })
  }

  // clearHistory drops the recorded history.
  clearHistory(): Promise<void> {
    return this.call<void>({ kind: 'clearHistory' })
  }

  // exportHistory renders the history in the given format.
  exportHistory(format: ExportFormat): Promise<string> {
    return this.call<string>({ kind: 'exportHistory', format })
  }

  // listTables returns the names of the tables and views that exist.
  listTables(): Promise<string[]> {
    return this.call<string[]>({ kind: 'listTables' })
  }
}

// client is the shared instance used by the whole app.
export const client = new GoogleSQLiteClient()
