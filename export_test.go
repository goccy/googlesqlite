package googlesqlite

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	internal "github.com/goccy/googlesqlite/internal"
)

// This file is compiled only for tests. It re-exposes functionality
// that is intentionally not part of the public API but still needs
// black-box test coverage from package googlesqlite_test. The
// implementations live here rather than in the production sources so
// that the public surface stays minimal. If any of these is later
// promoted to the public API, move the declaration into a normal
// (non _test.go) file.

// WithCurrentTime overrides the wall clock seen by CURRENT_DATE,
// CURRENT_DATETIME, CURRENT_TIME and CURRENT_TIMESTAMP for queries run
// with the returned context.
func WithCurrentTime(ctx context.Context, now time.Time) context.Context {
	return internal.WithCurrentTime(ctx, now)
}

// CurrentTime returns the time injected by WithCurrentTime, or nil.
func CurrentTime(ctx context.Context) *time.Time {
	return internal.CurrentTime(ctx)
}

// ChangedCatalogFromResult retrieves modified catalog information from
// a sql.Result produced by the googlesqlite driver.
func ChangedCatalogFromResult(result sql.Result) (*ChangedCatalog, error) {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("googlesqlite: unexpected sql.Result layout. expected sql.Result type is struct but got %T", result)
	}
	resi := rv.FieldByName("resi")
	if !resi.IsValid() {
		return nil, fmt.Errorf("googlesqlite: unexpected sql.Result layout")
	}
	driverValue := resi.Elem()
	if driverValue.Type() != reflect.TypeFor[*internal.Result]() {
		return nil, fmt.Errorf("googlesqlite: sql.Result must be an instance created using the googlesqlite database driver")
	}
	googlesqliteResult := (*internal.Result)(driverValue.UnsafePointer())
	return googlesqliteResult.ChangedCatalog(), nil
}

// SetMaterializeCTE controls whether CTEs referenced more than once
// are emitted with the SQLite MATERIALIZED hint.
func (c *Conn) SetMaterializeCTE(enabled bool) { c.materializeCTE = enabled }

// MaterializeCTE reports the current materialize-multi-ref-CTE flag.
func (c *Conn) MaterializeCTE() bool { return c.materializeCTE }

// SetAutoIndexMode toggles automatic index creation in the analyzer.
func (c *Conn) SetAutoIndexMode(enabled bool) { c.analyzer.SetAutoIndexMode(enabled) }

// SetExplainMode toggles EXPLAIN QUERY PLAN execution in the analyzer.
func (c *Conn) SetExplainMode(enabled bool) { c.analyzer.SetExplainMode(enabled) }

// MaxNamePath returns the maximum name-path length.
func (c *Conn) MaxNamePath() int { return c.analyzer.MaxNamePath() }

// NamePath returns the name-path prefix.
func (c *Conn) NamePath() []string { return c.analyzer.NamePath() }

// AddNamePath appends a single element to the name-path prefix.
func (c *Conn) AddNamePath(path string) error { return c.analyzer.AddNamePath(path) }

// RegisterProto installs a serialized google.protobuf.FileDescriptorProto
// into the connection's catalog so the analyzer can resolve the
// messages and enums it declares.
func (c *Conn) RegisterProto(fileDescriptorProtoBytes []byte) error {
	return c.catalog.RegisterProto(fileDescriptorProtoBytes)
}

// RegisterProtoMessage promotes a message already present in the
// catalog's DescriptorPool to a fully resolved analyzer ProtoType.
func (c *Conn) RegisterProtoMessage(fullName string) error {
	_, err := c.catalog.RegisterProtoMessage(fullName)
	return err
}
