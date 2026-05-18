import {
  Box,
  Divider,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Tooltip,
  Typography,
} from '@mui/material'
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline'
import RefreshIcon from '@mui/icons-material/Refresh'
import TableChartIcon from '@mui/icons-material/TableChart'

interface TablesPanelProps {
  tables: string[]
  // onSelect is called with a table name; the caller queries it.
  onSelect: (name: string) => void
  // onDrop is called with a table name; the caller drops it.
  onDrop: (name: string) => void
  onRefresh: () => void
}

// TablesPanel lists the tables and views in the session. Selecting one
// queries its current contents; the trash button drops it.
export function TablesPanel({
  tables,
  onSelect,
  onDrop,
  onRefresh,
}: TablesPanelProps) {
  return (
    <Box display="flex" flexDirection="column" minHeight={0}>
      <Box
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        px={1}
        py={0.5}
      >
        <Typography variant="subtitle2">Tables</Typography>
        <Tooltip title="Refresh">
          <IconButton size="small" onClick={onRefresh}>
            <RefreshIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      </Box>
      <Divider />
      <Box sx={{ overflow: 'auto' }}>
        {tables.length === 0 ? (
          <Typography variant="body2" color="text.secondary" sx={{ p: 1 }}>
            No tables yet — create one with CREATE TABLE.
          </Typography>
        ) : (
          <List dense disablePadding>
            {tables.map((name) => (
              <ListItem
                key={name}
                disablePadding
                secondaryAction={
                  <Tooltip title="Drop table">
                    <IconButton
                      edge="end"
                      size="small"
                      aria-label={`drop ${name}`}
                      onClick={() => onDrop(name)}
                    >
                      <DeleteOutlineIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                }
              >
                <ListItemButton onClick={() => onSelect(name)}>
                  <TableChartIcon
                    fontSize="small"
                    sx={{ mr: 1, color: 'text.secondary' }}
                  />
                  <ListItemText
                    primary={name}
                    primaryTypographyProps={{ noWrap: true }}
                  />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        )}
      </Box>
    </Box>
  )
}
