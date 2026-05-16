package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	upstreamRepo    = "https://github.com/google/googlesql.git"
	upstreamSubpath = "docs"
	vendorPath      = "docs/third_party/googlesql-docs"
	upstreamFile    = "UPSTREAM.txt"
)

func init() {
	register(command{
		name:    "upstream-sync",
		summary: "refresh docs/third_party/googlesql-docs/ from upstream and report a diff",
		run:     runUpstreamSync,
	})
}

func runUpstreamSync(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("upstream-sync", flag.ContinueOnError)
	repoFlag := flags.String("repo", upstreamRepo, "git URL of the upstream repository")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}

	tmp, err := os.MkdirTemp("", "googlesqlite-upstream-*")
	if err != nil {
		return fmt.Errorf("mkdir tmp: %w", err)
	}
	defer os.RemoveAll(tmp)

	if err := runGit(ctx, "", "clone",
		"--depth=1", "--filter=blob:none", "--sparse",
		*repoFlag, tmp,
	); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	if err := runGit(ctx, tmp, "sparse-checkout", "set", "--no-cone", upstreamSubpath, "LICENSE"); err != nil {
		return fmt.Errorf("git sparse-checkout: %w", err)
	}

	sha, err := captureGit(ctx, tmp, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("git rev-parse: %w", err)
	}
	commitDate, err := captureGit(ctx, tmp, "log", "-1", "--format=%cI")
	if err != nil {
		return fmt.Errorf("git log: %w", err)
	}

	wantFiles, err := collectMarkdown(filepath.Join(tmp, upstreamSubpath))
	if err != nil {
		return err
	}
	wantLicense, err := os.ReadFile(filepath.Join(tmp, "LICENSE"))
	if err != nil {
		return fmt.Errorf("read upstream LICENSE: %w", err)
	}

	if err := os.MkdirAll(vendorPath, 0o755); err != nil {
		return fmt.Errorf("mkdir vendor: %w", err)
	}

	currentFiles, err := collectMarkdown(vendorPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	added, removed, modified := diffSets(currentFiles, wantFiles)
	if err := replaceMarkdown(vendorPath, currentFiles, wantFiles); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(vendorPath, "LICENSE"), wantLicense, 0o644); err != nil {
		return fmt.Errorf("write LICENSE: %w", err)
	}

	upstream := upstreamFileContents(*repoFlag, strings.TrimSpace(sha), strings.TrimSpace(commitDate))
	if err := os.WriteFile(filepath.Join(vendorPath, upstreamFile), []byte(upstream), 0o644); err != nil {
		return fmt.Errorf("write UPSTREAM.txt: %w", err)
	}

	printDiffSummary(added, removed, modified)
	return nil
}

func runGit(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func captureGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// collectMarkdown reads every *.md file directly inside dir and returns
// {basename → bytes}.
func collectMarkdown(dir string) (map[string][]byte, error) {
	out := map[string][]byte{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		out[name] = data
	}
	return out, nil
}

func diffSets(current, want map[string][]byte) (added, removed, modified []string) {
	for name := range want {
		if _, ok := current[name]; !ok {
			added = append(added, name)
			continue
		}
		if !bytes.Equal(current[name], want[name]) {
			modified = append(modified, name)
		}
	}
	for name := range current {
		if _, ok := want[name]; !ok {
			removed = append(removed, name)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(modified)
	return
}

func replaceMarkdown(dir string, current, want map[string][]byte) error {
	for name := range current {
		if _, keep := want[name]; keep {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return fmt.Errorf("remove %s: %w", name, err)
		}
	}
	for name, data := range want {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}

func upstreamFileContents(repoURL, sha, commitDate string) string {
	now := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf(
		`upstream: %s
path: %s/
commit: %s
commit_date: %s
synced_at: %s
license: Apache-2.0 (see LICENSE in this directory)

This directory holds a snapshot of the GoogleSQL documentation. It is the
canonical input to `+"`cmd/specctl normalize`"+`, which produces the
per-function/per-type files under `+"`docs/specs/googlesql/`"+`.

To refresh: run `+"`make spec/upstream-sync`"+`. Do not edit files here by
hand — local edits will be overwritten on the next sync.
`,
		strings.TrimSuffix(repoURL, ".git"),
		upstreamSubpath, sha, commitDate, now,
	)
}

func printDiffSummary(added, removed, modified []string) {
	fmt.Printf("upstream-sync: %s/ updated\n", vendorPath)
	fmt.Printf("  added:    %d\n", len(added))
	for _, n := range added {
		fmt.Printf("    + %s\n", n)
	}
	fmt.Printf("  removed:  %d\n", len(removed))
	for _, n := range removed {
		fmt.Printf("    - %s\n", n)
	}
	fmt.Printf("  modified: %d\n", len(modified))
	for _, n := range modified {
		fmt.Printf("    M %s\n", n)
	}
}
