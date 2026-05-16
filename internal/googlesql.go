// Small Go-side helpers built on top of go-googlesql. The go-googlesql
// generator is the source of truth for the wasm-bridged surface; this
// file only adds:
//
//   - tf(): a process-wide TypeFactory singleton for the few callers
//     that don't already hold one.
//   - m1[T]: an error-discarding wrapper used at call sites where
//     a googlesql accessor's (T, error) return is statically known to
//     succeed (the error path is exercised through the surrounding
//     analyzer/codec checks).
//   - ResolvedBaseFunctionCallNode: a short alias for the long
//     canonical ResolvedFunctionCallBase type so per-spec call sites
//     read more cleanly.
//   - ASTWalk: depth-first visitor over a googlesql AST subtree.
package internal

import (
	"fmt"
	"sync"

	googlesql "github.com/goccy/go-googlesql"
)

// tf returns a process-wide TypeFactory that's lazily created on first
// use. Lifetime is managed by the wasm module so holding onto it for
// the process lifetime is safe. The factory is initialized exactly
// once even under concurrent callers (sync.OnceValue) — otherwise
// parallel `NewCatalog` paths race on the underlying pointer.
var tf = sync.OnceValue(func() *googlesql.TypeFactory {
	f, err := googlesql.NewTypeFactory()
	if err != nil {
		panic(fmt.Errorf("failed to initialize type factory: %w", err))
	}
	return f
})

// m1 drops the error from a (T, error) pair returned by a go-googlesql
// accessor where the error is known to be nil at this call site.
func m1[T any](v T, _ error) T { return v }

// ResolvedBaseFunctionCallNode is a short alias for the canonical
// long-named ResolvedFunctionCallBase type.
type ResolvedBaseFunctionCallNode = googlesql.ResolvedFunctionCallBase

// resolvedCreateScope returns the scope (DEFAULT / TEMP / PRIVATE /
// PUBLIC) of any ResolvedCreateStatement-derived node. Every
// CREATE-family resolved node embeds *ResolvedCreateStatement, so the
// CreateScope() accessor is method-promoted; the interface assertion
// below picks it up regardless of the concrete subtype.
func resolvedCreateScope(h any) googlesql.ResolvedCreateStatementEnums_CreateScope {
	type scopeGetter interface {
		CreateScope() (googlesql.ResolvedCreateStatementEnums_CreateScope, error)
	}
	if g, ok := h.(scopeGetter); ok {
		if v, err := g.CreateScope(); err == nil {
			return v
		}
	}
	return googlesql.ResolvedCreateStatementEnums_CreateScopeCreateDefaultScope
}

// ASTWalk traverses an ASTNode subtree depth-first via Child(i) /
// NumChildren(), invoking visit on each node. Returns the first error
// the visitor produces.
func astWalk(root googlesql.ASTNode, visit func(googlesql.ASTNode) error) error {
	if root == nil {
		return nil
	}
	var walk func(n googlesql.ASTNode) error
	walk = func(n googlesql.ASTNode) error {
		if n == nil {
			return nil
		}
		if err := visit(n); err != nil {
			return err
		}
		num, err := n.NumChildren()
		if err != nil {
			return nil
		}
		for i := range num {
			child, err := n.Child(i)
			if err != nil || child == nil {
				continue
			}
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(root)
}
