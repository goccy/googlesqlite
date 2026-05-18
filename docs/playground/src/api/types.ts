// Types shared between the UI, the worker client and the worker.

// QueryResult mirrors one entry of the array returned by the wasm
// `exec` call (see cmd/googlesqlite/main_js.go: entryToJS).
export interface QueryResult {
  statement: string
  sqliteQuery: string
  isQuery: boolean
  columns: string[]
  rows: string[][]
  rowsAffected: number
  error: string
  elapsedMs: number
  timestamp: string
}

// ExecResponse is the value the wasm `exec` call returns on success.
export interface ExecResponse {
  output: string
  results: QueryResult[]
}

// ExportFormat enumerates the history export formats the engine
// supports.
export type ExportFormat = 'json' | 'sql' | 'markdown' | 'csv'

// WorkerRequest is a message sent from the UI thread to the worker.
// Every request carries a correlation id echoed back in the response.
export type WorkerRequest =
  | { id: number; kind: 'exec'; sql: string }
  | { id: number; kind: 'getHistory' }
  | { id: number; kind: 'clearHistory' }
  | { id: number; kind: 'exportHistory'; format: ExportFormat }
  | { id: number; kind: 'listTables' }

// LoadPhase names a stage of engine initialisation.
export type LoadPhase = 'timezone' | 'download' | 'start'

// LoadProgress reports initialisation progress. loaded / total are
// meaningful (bytes) only for the 'download' phase; they are 0
// otherwise, signalling an indeterminate stage.
export interface LoadProgress {
  phase: LoadPhase
  loaded: number
  total: number
}

// WorkerResponse is a message sent from the worker to the UI thread.
export type WorkerResponse =
  | { type: 'ready' }
  | { type: 'init-error'; error: string }
  | ({ type: 'progress' } & LoadProgress)
  | { type: 'result'; id: number; ok: true; data: unknown }
  | { type: 'result'; id: number; ok: false; error: string }
