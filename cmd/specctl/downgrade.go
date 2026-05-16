package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

func init() {
	register(command{
		name:    "downgrade",
		summary: "downgrade a list of spec paths from implemented to partial (and stamp a note)",
		run: func(_ context.Context, args []string) error {
			return runDowngrade(args)
		},
	})
}

// runDowngrade reads spec-path lines from --list <file> (or stdin) and
// rewrites each spec's frontmatter so `status:` becomes `partial`. When
// the spec has no `notes:` field, --note <text> is inserted as a
// single-line note. Existing notes are left untouched.
//
// Input lines may be either:
//   - a spec path relative to the project root
//     (e.g. docs/specs/googlesql/functions/proto/extract.md)
//   - a TestSpec subtest path
//     (e.g. googlesql/functions/proto/extract)
//
// Lines starting with `#` are ignored.
func runDowngrade(args []string) error {
	listPath := ""
	note := ""
	target := "partial"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--list":
			if i+1 < len(args) {
				listPath = args[i+1]
				i++
			}
		case "--note":
			if i+1 < len(args) {
				note = args[i+1]
				i++
			}
		case "--to":
			// Target status (defaults to "partial"). Accepts any
			// status string; the spec is rewritten regardless of its
			// current state.
			if i+1 < len(args) {
				target = args[i+1]
				i++
			}
		}
	}
	var r *bufio.Scanner
	if listPath != "" {
		f, err := os.Open(listPath)
		if err != nil {
			return err
		}
		defer f.Close()
		r = bufio.NewScanner(f)
	} else {
		r = bufio.NewScanner(os.Stdin)
	}

	changed := 0
	skipped := 0
	for r.Scan() {
		line := strings.TrimSpace(r.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		path := normalizeSpecPath(line)
		ok, err := downgradeSpec(path, note, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn %s: %v\n", path, err)
			skipped++
			continue
		}
		if ok {
			changed++
		} else {
			skipped++
		}
	}
	fmt.Fprintf(os.Stderr, "downgrade: %d changed, %d skipped\n", changed, skipped)
	return r.Err()
}

func normalizeSpecPath(s string) string {
	if strings.HasPrefix(s, "docs/specs/") {
		return s
	}
	if strings.HasSuffix(s, ".md") {
		return s
	}
	return "docs/specs/" + s + ".md"
}

func downgradeSpec(path, note, target string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	body := string(data)
	if !strings.HasPrefix(body, "---\n") {
		return false, fmt.Errorf("missing frontmatter")
	}
	endIdx := strings.Index(body[4:], "\n---\n")
	if endIdx < 0 {
		return false, fmt.Errorf("frontmatter not closed")
	}
	head := body[:4+endIdx+1] // includes trailing newline before closing ---
	closing := body[4+endIdx+1:]

	lines := strings.Split(head, "\n")
	statusIdx := -1
	notesIdx := -1
	sourceIdx := -1
	for i, l := range lines {
		switch {
		case strings.HasPrefix(l, "status:"):
			statusIdx = i
		case strings.HasPrefix(l, "notes:"):
			notesIdx = i
		case strings.HasPrefix(l, "source_url:"):
			sourceIdx = i
		}
	}
	if statusIdx < 0 {
		return false, fmt.Errorf("status: line not found")
	}
	cur := strings.TrimSpace(strings.TrimPrefix(lines[statusIdx], "status:"))
	if cur == target {
		// Already at the target status. Still stamp a notes line if
		// requested and none present.
		if note == "" || notesIdx >= 0 {
			return false, nil
		}
	} else {
		lines[statusIdx] = "status: " + target
	}

	if note != "" && notesIdx < 0 {
		// Insert a notes line right before source_url (or at the end
		// of the frontmatter if source_url is missing).
		insertAt := sourceIdx
		if insertAt < 0 {
			insertAt = len(lines) - 1
		}
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:insertAt]...)
		newLines = append(newLines, "notes: "+note)
		newLines = append(newLines, lines[insertAt:]...)
		lines = newLines
	}
	newHead := strings.Join(lines, "\n")
	if newHead == head {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(newHead+closing), 0o644)
}
