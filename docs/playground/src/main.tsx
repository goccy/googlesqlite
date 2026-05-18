import { StrictMode, useCallback, useMemo, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { CssBaseline, ThemeProvider } from '@mui/material'
import { App } from './App'
import {
  createAppTheme,
  initialThemeMode,
  storeThemeMode,
  type ThemeMode,
} from './theme'
import './index.css'

// Root owns the theme mode so the whole app can switch between light
// and dark at runtime.
function Root() {
  const [mode, setMode] = useState<ThemeMode>(initialThemeMode)
  const theme = useMemo(() => createAppTheme(mode), [mode])
  const toggleMode = useCallback(() => {
    setMode((current) => {
      const next: ThemeMode = current === 'dark' ? 'light' : 'dark'
      storeThemeMode(next)
      return next
    })
  }, [])

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <App mode={mode} onToggleMode={toggleMode} />
    </ThemeProvider>
  )
}

const container = document.getElementById('root')
if (!container) {
  throw new Error('root container not found')
}

createRoot(container).render(
  <StrictMode>
    <Root />
  </StrictMode>,
)
