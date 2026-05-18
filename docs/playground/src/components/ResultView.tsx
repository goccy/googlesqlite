import {
  Alert,
  Box,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'
import type { QueryResult } from '../api/types'

interface ResultViewProps {
  results: QueryResult[]
  showDebug: boolean
}

// ResultView renders the results of the most recent run, one block
// per statement.
export function ResultView({ results, showDebug }: ResultViewProps) {
  if (results.length === 0) {
    return (
      <Typography color="text.secondary">
        No results yet. Write a query and press Run (⌘/Ctrl+Enter).
      </Typography>
    )
  }
  return (
    <Box display="flex" flexDirection="column" gap={2}>
      {results.map((result, i) => (
        <ResultBlock key={i} result={result} showDebug={showDebug} />
      ))}
    </Box>
  )
}

function ResultBlock({
  result,
  showDebug,
}: {
  result: QueryResult
  showDebug: boolean
}) {
  return (
    <Paper variant="outlined" sx={{ p: 1.5 }}>
      <Typography
        variant="caption"
        component="pre"
        sx={{ whiteSpace: 'pre-wrap', m: 0, mb: 1, fontFamily: 'monospace' }}
      >
        {result.statement}
      </Typography>
      {showDebug && result.sqliteQuery && (
        <Typography
          variant="caption"
          component="pre"
          sx={{
            whiteSpace: 'pre-wrap',
            m: 0,
            mb: 1,
            color: 'info.main',
            fontFamily: 'monospace',
          }}
        >
          {`-- sqlite: ${result.sqliteQuery}`}
        </Typography>
      )}
      {result.error ? (
        <Alert severity="error">{result.error}</Alert>
      ) : result.isQuery ? (
        <QueryTable
          columns={result.columns}
          rows={result.rows}
          elapsedMs={result.elapsedMs}
        />
      ) : (
        <Alert severity="success">
          {`Query OK, ${result.rowsAffected} row(s) affected (${result.elapsedMs.toFixed(1)} ms)`}
        </Alert>
      )}
    </Paper>
  )
}

function QueryTable({
  columns,
  rows,
  elapsedMs,
}: {
  columns: string[]
  rows: string[][]
  elapsedMs: number
}) {
  if (columns.length === 0) {
    return <Alert severity="info">Empty result.</Alert>
  }
  return (
    <Box>
      <Box sx={{ overflow: 'auto' }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              {columns.map((name, i) => (
                <TableCell key={i} sx={{ fontWeight: 'bold' }}>
                  {name}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.map((row, ri) => (
              <TableRow key={ri}>
                {row.map((cell, ci) => (
                  <TableCell key={ci} sx={{ whiteSpace: 'pre' }}>
                    {cell}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
      <Typography variant="caption" color="text.secondary">
        {`${rows.length} row(s) in set (${elapsedMs.toFixed(1)} ms)`}
      </Typography>
    </Box>
  )
}
