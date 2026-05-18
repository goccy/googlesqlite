import { createTheme, type Theme } from '@mui/material/styles'

export type ThemeMode = 'light' | 'dark'

const STORAGE_KEY = 'googlesqlite.theme'

// initialThemeMode resolves the start-up theme: a previously stored
// preference, or the operating-system colour scheme.
export function initialThemeMode(): ThemeMode {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved === 'light' || saved === 'dark') {
    return saved
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light'
}

// storeThemeMode persists the chosen theme.
export function storeThemeMode(mode: ThemeMode): void {
  localStorage.setItem(STORAGE_KEY, mode)
}

// createAppTheme builds the MUI theme for the given mode.
export function createAppTheme(mode: ThemeMode): Theme {
  return createTheme({
    palette: {
      mode,
      primary: { main: mode === 'dark' ? '#8ab4f8' : '#1a73e8' },
    },
    typography: {
      fontSize: 13,
    },
  })
}
