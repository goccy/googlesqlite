import { useState } from 'react'
import type { MouseEvent } from 'react'
import {
  Box,
  Divider,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  Menu,
  MenuItem,
  Tooltip,
  Typography,
} from '@mui/material'
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep'
import DownloadIcon from '@mui/icons-material/Download'
import type { ExportFormat, QueryResult } from '../api/types'

interface HistoryPanelProps {
  entries: QueryResult[]
  // onSelect is called with the full entry so the caller can restore
  // both the query text and its recorded result.
  onSelect: (entry: QueryResult) => void
  onClear: () => void
  onExport: (format: ExportFormat) => void
}

const EXPORT_FORMATS: { format: ExportFormat; label: string }[] = [
  { format: 'json', label: 'JSON' },
  { format: 'sql', label: 'SQL script' },
  { format: 'markdown', label: 'Markdown' },
  { format: 'csv', label: 'CSV' },
]

// HistoryPanel lists previously executed statements, newest first, and
// offers history export and clear actions.
export function HistoryPanel({
  entries,
  onSelect,
  onClear,
  onExport,
}: HistoryPanelProps) {
  const [anchor, setAnchor] = useState<HTMLElement | null>(null)

  const openMenu = (e: MouseEvent<HTMLElement>) => setAnchor(e.currentTarget)
  const closeMenu = () => setAnchor(null)

  return (
    <Box display="flex" flexDirection="column" height="100%">
      <Box
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        px={1}
        py={0.5}
      >
        <Typography variant="subtitle2">History</Typography>
        <Box>
          <Tooltip title="Export history">
            <span>
              <IconButton
                size="small"
                onClick={openMenu}
                disabled={entries.length === 0}
              >
                <DownloadIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
          <Tooltip title="Clear history">
            <span>
              <IconButton
                size="small"
                onClick={onClear}
                disabled={entries.length === 0}
              >
                <DeleteSweepIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
        </Box>
      </Box>
      <Menu anchorEl={anchor} open={anchor !== null} onClose={closeMenu}>
        {EXPORT_FORMATS.map((item) => (
          <MenuItem
            key={item.format}
            onClick={() => {
              onExport(item.format)
              closeMenu()
            }}
          >
            {item.label}
          </MenuItem>
        ))}
      </Menu>
      <Divider />
      <Box flex={1} overflow="auto">
        {entries.length === 0 ? (
          <Typography variant="body2" color="text.secondary" sx={{ p: 1 }}>
            No history yet.
          </Typography>
        ) : (
          <List dense disablePadding>
            {entries
              .map((entry, index) => ({ entry, index }))
              .reverse()
              .map(({ entry, index }) => (
                <ListItemButton
                  key={index}
                  onClick={() => onSelect(entry)}
                >
                  <ListItemText
                    primary={entry.statement}
                    secondary={historySummary(entry)}
                    primaryTypographyProps={{ noWrap: true }}
                    secondaryTypographyProps={{ noWrap: true }}
                  />
                </ListItemButton>
              ))}
          </List>
        )}
      </Box>
    </Box>
  )
}

function historySummary(entry: QueryResult): string {
  const time = formatTimestamp(entry.timestamp)
  const detail = entry.error
    ? `error: ${entry.error}`
    : `${
        entry.isQuery
          ? `${entry.rows.length} row(s)`
          : `${entry.rowsAffected} affected`
      } · ${entry.elapsedMs.toFixed(0)} ms`
  return time ? `${time} · ${detail}` : detail
}

// formatTimestamp renders an ISO timestamp as a local time-of-day
// string, or "" when the value cannot be parsed.
function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp)
  return Number.isNaN(date.getTime()) ? '' : date.toLocaleTimeString()
}
