import { useCallback, useEffect, useRef, useState } from 'react'
import type { ChangeEvent } from 'react'
import {
  Alert,
  AppBar,
  Box,
  Button,
  Divider,
  Drawer,
  FormControlLabel,
  IconButton,
  LinearProgress,
  Menu,
  MenuItem,
  Paper,
  Switch,
  Toolbar,
  Tooltip,
  Typography,
  useMediaQuery,
  useTheme,
} from '@mui/material'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import HistoryIcon from '@mui/icons-material/History'
import LightModeIcon from '@mui/icons-material/LightMode'
import LockIcon from '@mui/icons-material/Lock'
import MenuBookIcon from '@mui/icons-material/MenuBook'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import UploadFileIcon from '@mui/icons-material/UploadFile'
import { Editor } from './components/Editor'
import { ResultView } from './components/ResultView'
import { HistoryPanel } from './components/HistoryPanel'
import { TablesPanel } from './components/TablesPanel'
import { client } from './api/googlesqlite'
import { EXAMPLES } from './data/examples'
import type { ExportFormat, LoadProgress, QueryResult } from './api/types'
import type { ThemeMode } from './theme'

const SAMPLE_QUERY = `-- Welcome to the GoogleSQLite Playground.
-- Run with the Run button or Cmd/Ctrl+Enter.
SELECT 'こんにちは' AS greeting, 1 + 1 AS sum;`

const HISTORY_WIDTH = 300

type Status = 'loading' | 'ready' | 'error'

interface AppProps {
  mode: ThemeMode
  onToggleMode: () => void
}

export function App({ mode, onToggleMode }: AppProps) {
  const theme = useTheme()
  const isMobile = useMediaQuery(theme.breakpoints.down('md'))

  const [status, setStatus] = useState<Status>('loading')
  const [initError, setInitError] = useState('')
  const [progress, setProgress] = useState<LoadProgress | null>(null)
  const [sql, setSql] = useState(SAMPLE_QUERY)
  const [results, setResults] = useState<QueryResult[]>([])
  const [history, setHistory] = useState<QueryResult[]>([])
  const [tables, setTables] = useState<string[]>([])
  const [running, setRunning] = useState(false)
  const [debug, setDebug] = useState(false)
  const [runError, setRunError] = useState('')
  const [historyOpen, setHistoryOpen] = useState(false)
  const [examplesAnchor, setExamplesAnchor] = useState<HTMLElement | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    let cancelled = false
    client.onProgress((p) => {
      if (!cancelled) {
        setProgress(p)
      }
    })
    client.ready
      .then(() => Promise.all([client.getHistory(), client.listTables()]))
      .then(([entries, tableNames]) => {
        if (cancelled) {
          return
        }
        setHistory(entries)
        setTables(tableNames)
        setStatus('ready')
      })
      .catch((e: unknown) => {
        if (cancelled) {
          return
        }
        setInitError(e instanceof Error ? e.message : String(e))
        setStatus('error')
      })
    return () => {
      cancelled = true
    }
  }, [])

  // runSql executes an explicit statement (not necessarily the editor
  // content) and refreshes the results, history and table list.
  const runSql = useCallback(
    async (text: string) => {
      if (running || status !== 'ready') {
        return
      }
      setRunning(true)
      setRunError('')
      try {
        const response = await client.exec(text)
        setResults(response.results)
        const [entries, tableNames] = await Promise.all([
          client.getHistory(),
          client.listTables(),
        ])
        setHistory(entries)
        setTables(tableNames)
      } catch (e: unknown) {
        setRunError(e instanceof Error ? e.message : String(e))
      } finally {
        setRunning(false)
      }
    },
    [running, status],
  )

  const run = useCallback(() => {
    void runSql(sql)
  }, [runSql, sql])

  const loadAndRun = useCallback(
    (text: string) => {
      setSql(text)
      void runSql(text)
    },
    [runSql],
  )

  const refreshTables = useCallback(() => {
    client
      .listTables()
      .then(setTables)
      .catch(() => undefined)
  }, [])

  const clearHistory = useCallback(async () => {
    await client.clearHistory()
    setHistory([])
  }, [])

  const exportHistory = useCallback(async (format: ExportFormat) => {
    const content = await client.exportHistory(format)
    downloadHistory(content, format)
  }, [])

  const handleFileUpload = useCallback(
    async (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0]
      e.target.value = '' // allow re-selecting the same file
      if (!file) {
        return
      }
      loadAndRun(await file.text())
    },
    [loadAndRun],
  )

  if (status !== 'ready') {
    return <LoadingScreen status={status} error={initError} progress={progress} />
  }

  const debugToggle = (
    <FormControlLabel
      control={
        <Switch
          size="small"
          checked={debug}
          onChange={(e) => setDebug(e.target.checked)}
        />
      }
      label="Show translated SQLite"
    />
  )

  const sidePanel = (
    <Box display="flex" flexDirection="column" height="100%" minHeight={0}>
      <Box sx={{ flexShrink: 0, maxHeight: '45%', display: 'flex' }}>
        <TablesPanel
          tables={tables}
          onSelect={(name) => {
            loadAndRun(`SELECT * FROM \`${name}\``)
            setHistoryOpen(false)
          }}
          onDrop={(name) => {
            if (
              window.confirm(`Drop table "${name}"? This cannot be undone.`)
            ) {
              void runSql(`DROP TABLE \`${name}\``)
            }
          }}
          onRefresh={refreshTables}
        />
      </Box>
      <Divider />
      <Box flex={1} minHeight={0}>
        <HistoryPanel
          entries={history}
          onSelect={(entry) => {
            setSql(entry.statement)
            setResults([entry])
            setRunError('')
            setHistoryOpen(false)
          }}
          onClear={clearHistory}
          onExport={exportHistory}
        />
      </Box>
    </Box>
  )

  return (
    <Box
      display="flex"
      flexDirection="column"
      height="100%"
      width="100%"
      overflow="hidden"
    >
      <AppBar position="static" color="default" elevation={1}>
        <Toolbar variant="dense" sx={{ gap: 1 }}>
          <Typography variant="h6" noWrap sx={{ flexGrow: 1 }}>
            {isMobile ? 'GoogleSQLite' : 'GoogleSQLite Playground'}
          </Typography>
          {!isMobile && debugToggle}
          <Tooltip
            title={mode === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
          >
            <IconButton aria-label="toggle theme" onClick={onToggleMode}>
              {mode === 'dark' ? <LightModeIcon /> : <DarkModeIcon />}
            </IconButton>
          </Tooltip>
          <Button
            variant="contained"
            startIcon={<PlayArrowIcon />}
            onClick={run}
            disabled={running}
          >
            {running ? 'Running…' : 'Run'}
          </Button>
          {isMobile && (
            <IconButton
              edge="end"
              aria-label="history"
              onClick={() => setHistoryOpen(true)}
            >
              <HistoryIcon />
            </IconButton>
          )}
        </Toolbar>
      </AppBar>

      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: 0.75,
          px: 2,
          py: 0.5,
          bgcolor: 'action.hover',
          borderBottom: 1,
          borderColor: 'divider',
        }}
      >
        <LockIcon sx={{ fontSize: 16 }} color="action" />
        <Typography variant="caption" color="text.secondary">
          Every query runs entirely in your browser — your queries and data are
          never uploaded to a server.
        </Typography>
      </Box>

      <Box
        display="flex"
        flexDirection={isMobile ? 'column' : 'row'}
        flex={1}
        minHeight={0}
      >
        <Box
          sx={{
            width: isMobile ? '100%' : '40%',
            minWidth: isMobile ? 'auto' : 320,
            height: isMobile ? '45vh' : 'auto',
            flexShrink: 0,
            display: 'flex',
            flexDirection: 'column',
            borderRight: isMobile ? 0 : 1,
            borderBottom: isMobile ? 1 : 0,
            borderColor: 'divider',
          }}
        >
          <Box
            sx={{
              display: 'flex',
              gap: 1,
              p: 0.5,
              borderBottom: 1,
              borderColor: 'divider',
            }}
          >
            <Button
              size="small"
              startIcon={<MenuBookIcon />}
              onClick={(e) => setExamplesAnchor(e.currentTarget)}
            >
              Examples
            </Button>
            <Button
              size="small"
              startIcon={<UploadFileIcon />}
              onClick={() => fileInputRef.current?.click()}
            >
              Upload .sql
            </Button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".sql,text/plain"
              style={{ display: 'none' }}
              onChange={handleFileUpload}
            />
            <Menu
              anchorEl={examplesAnchor}
              open={examplesAnchor !== null}
              onClose={() => setExamplesAnchor(null)}
            >
              {EXAMPLES.map((example) => (
                <MenuItem
                  key={example.label}
                  onClick={() => {
                    setExamplesAnchor(null)
                    // Load the example into the editor; the user runs it.
                    setSql(example.sql)
                  }}
                >
                  {example.label}
                </MenuItem>
              ))}
            </Menu>
          </Box>
          <Box flex={1} minHeight={0}>
            <Editor value={sql} onChange={setSql} onRun={run} mode={mode} />
          </Box>
        </Box>

        <Box flex={1} minWidth={0} sx={{ overflow: 'auto', p: 2 }}>
          {isMobile && <Box sx={{ mb: 1 }}>{debugToggle}</Box>}
          {runError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {runError}
            </Alert>
          )}
          <ResultView results={results} showDebug={debug} />
        </Box>

        {!isMobile && (
          <Paper
            square
            elevation={0}
            sx={{ width: HISTORY_WIDTH, borderLeft: 1, borderColor: 'divider' }}
          >
            {sidePanel}
          </Paper>
        )}
      </Box>

      {isMobile && (
        <Drawer
          anchor="right"
          open={historyOpen}
          onClose={() => setHistoryOpen(false)}
        >
          <Box sx={{ width: 'min(320px, 85vw)', height: '100%' }}>
            {sidePanel}
          </Box>
        </Drawer>
      )}
    </Box>
  )
}

const PHASE_LABEL: Record<LoadProgress['phase'], string> = {
  timezone: 'Loading timezone data…',
  download: 'Downloading the engine…',
  start: 'Starting the engine…',
}

function LoadingScreen({
  status,
  error,
  progress,
}: {
  status: Status
  error: string
  progress: LoadProgress | null
}) {
  if (status === 'error') {
    return (
      <Box
        display="flex"
        alignItems="center"
        justifyContent="center"
        height="100%"
        p={4}
      >
        <Alert severity="error" sx={{ maxWidth: 640 }}>
          Failed to start the GoogleSQLite engine: {error}
        </Alert>
      </Box>
    )
  }

  const determinate = progress?.phase === 'download' && progress.total > 0
  const pct = determinate
    ? Math.floor((progress.loaded / progress.total) * 100)
    : 0
  const label = progress ? PHASE_LABEL[progress.phase] : 'Loading…'

  return (
    <Box
      display="flex"
      flexDirection="column"
      alignItems="center"
      justifyContent="center"
      gap={2}
      height="100%"
      px={3}
    >
      <Typography variant="h6">GoogleSQLite Playground</Typography>
      <Box sx={{ width: 'min(360px, 80vw)' }}>
        <LinearProgress
          variant={determinate ? 'determinate' : 'indeterminate'}
          value={pct}
        />
      </Box>
      <Typography color="text.secondary">
        {label}
        {determinate ? ` ${pct}%` : ''}
      </Typography>
      <Typography variant="caption" color="text.secondary" textAlign="center">
        The first load downloads a large WebAssembly module; it is then
        cached by the browser.
      </Typography>
    </Box>
  )
}

const FILE_EXTENSION: Record<ExportFormat, string> = {
  json: 'json',
  sql: 'sql',
  markdown: 'md',
  csv: 'csv',
}

const MIME_TYPE: Record<ExportFormat, string> = {
  json: 'application/json',
  sql: 'application/sql',
  markdown: 'text/markdown',
  csv: 'text/csv',
}

function downloadHistory(content: string, format: ExportFormat): void {
  const blob = new Blob([content], { type: MIME_TYPE[format] })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `googlesqlite-history.${FILE_EXTENSION[format]}`
  anchor.click()
  URL.revokeObjectURL(url)
}
