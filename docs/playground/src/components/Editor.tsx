import { useCallback } from 'react'
import type { KeyboardEvent } from 'react'
import CodeMirror, { EditorView } from '@uiw/react-codemirror'
import { sql } from '@codemirror/lang-sql'
import type { ThemeMode } from '../theme'

interface EditorProps {
  value: string
  onChange: (value: string) => void
  onRun: () => void
  mode: ThemeMode
}

// editorExtensions are created once:
//  - the SQL language mode;
//  - line wrapping, so long statements wrap instead of overflowing;
//  - a 16px font size. iOS Safari auto-zooms when a focused field has
//    a font smaller than 16px (and does not zoom back out), so keeping
//    the editor at 16px avoids that zoom entirely.
const editorExtensions = [
  sql(),
  EditorView.lineWrapping,
  EditorView.theme({ '&': { fontSize: '16px' } }),
]

// Editor is the CodeMirror 6 SQL input. Cmd/Ctrl+Enter runs the query.
export function Editor({ value, onChange, onRun, mode }: EditorProps) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLDivElement>) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault()
        onRun()
      }
    },
    [onRun],
  )

  return (
    <div onKeyDown={handleKeyDown} style={{ height: '100%', overflow: 'hidden' }}>
      <CodeMirror
        value={value}
        height="100%"
        theme={mode}
        extensions={editorExtensions}
        onChange={onChange}
      />
    </div>
  )
}
