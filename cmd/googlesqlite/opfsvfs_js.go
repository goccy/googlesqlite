//go:build js && wasm

package main

import (
	"errors"
	"io"
	"syscall/js"

	"github.com/ncruces/go-sqlite3/vfs"
)

// This file gives the Playground a SQLite VFS named "opfs" that stores
// the database in a single Origin Private File System (OPFS) file.
//
// A browser exposes no filesystem an ordinary SQLite VFS could use, so
// the database would otherwise have to live in memory and be copied
// out to IndexedDB to survive a reload. OPFS is a real, persistent,
// origin-scoped filesystem, and its FileSystemSyncAccessHandle exposes
// *synchronous* read/write/getSize/truncate/flush — exactly what a
// SQLite VFS needs.
//
// Acquiring the handle is asynchronous, so worker.ts does it before
// the Go program starts and publishes the handle on globalThis; every
// method here is then a plain synchronous call. The Playground DSN
// pairs this VFS with journal_mode=memory and temp_store=memory, so
// SQLite never opens a rollback journal or a temporary file: this VFS
// only ever serves the one main database.

// opfsVFSName is the VFS name the Playground DSN selects with ?vfs=.
const opfsVFSName = "opfs"

// dbHandleGlobal is the globalThis property where worker.ts publishes
// the OPFS FileSystemSyncAccessHandle for the database file.
const dbHandleGlobal = "__googlesqliteDBHandle"

var uint8Array = js.Global().Get("Uint8Array")

// registerOPFSVFS registers the "opfs" SQLite VFS. It must be called
// before the Playground database is opened.
func registerOPFSVFS() {
	vfs.Register(opfsVFSName, opfsVFS{})
}

type opfsVFS struct{}

func (opfsVFS) Open(name string, flags vfs.OpenFlag) (vfs.File, vfs.OpenFlag, error) {
	if flags&vfs.OPEN_MAIN_DB == 0 {
		// journal_mode=memory and temp_store=memory keep SQLite from
		// opening anything but the main database through this VFS.
		return nil, flags, errors.New("opfs vfs: only the main database is supported")
	}
	handle := js.Global().Get(dbHandleGlobal)
	if !handle.Truthy() {
		return nil, flags, errors.New("opfs vfs: database handle not initialised")
	}
	return &opfsFile{handle: handle}, flags, nil
}

func (opfsVFS) Delete(string, bool) error { return nil }

// Access always reports "not found". SQLite uses it mainly to probe
// for a hot journal; with journal_mode=memory there is never one.
func (opfsVFS) Access(string, vfs.AccessFlag) (bool, error) { return false, nil }

func (opfsVFS) FullPathname(name string) (string, error) { return name, nil }

// opfsFile is a SQLite File backed by an OPFS sync access handle.
type opfsFile struct {
	handle js.Value
}

func (f *opfsFile) ReadAt(p []byte, off int64) (int, error) {
	buf := uint8Array.New(len(p))
	n := f.handle.Call("read", buf, opfsAt(off)).Int()
	js.CopyBytesToGo(p, buf)
	if n < len(p) {
		// SQLite expects io.EOF on a short read; vfsRead then zero-fills
		// the tail and reports SQLITE_IOERR_SHORT_READ.
		return n, io.EOF
	}
	return n, nil
}

func (f *opfsFile) WriteAt(p []byte, off int64) (int, error) {
	buf := uint8Array.New(len(p))
	js.CopyBytesToJS(buf, p)
	n := f.handle.Call("write", buf, opfsAt(off)).Int()
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func (f *opfsFile) Truncate(size int64) error {
	f.handle.Call("truncate", size)
	return nil
}

func (f *opfsFile) Sync(vfs.SyncFlag) error {
	f.handle.Call("flush")
	return nil
}

func (f *opfsFile) Size() (int64, error) {
	// getSize returns a JS number; read it as float64 so a database
	// larger than the 32-bit wasm int is still represented exactly.
	return int64(f.handle.Call("getSize").Float()), nil
}

// Close flushes but deliberately keeps the OPFS handle open: the handle
// is owned by the page for its whole lifetime and reused by any later
// Open (for example the short-lived catalog connection).
func (f *opfsFile) Close() error {
	f.handle.Call("flush")
	return nil
}

// Locking is a no-op: the Playground uses a single database/sql
// connection at a time with serialised calls, and an OPFS sync access
// handle is itself exclusive, so there is nothing to coordinate.
func (f *opfsFile) Lock(vfs.LockLevel) error         { return nil }
func (f *opfsFile) Unlock(vfs.LockLevel) error       { return nil }
func (f *opfsFile) CheckReservedLock() (bool, error) { return false, nil }

func (f *opfsFile) SectorSize() int { return 0 }

func (f *opfsFile) DeviceCharacteristics() vfs.DeviceCharacteristic { return 0 }

// opfsAt builds the { at: offset } options object the OPFS read/write
// methods take.
func opfsAt(off int64) js.Value {
	return js.ValueOf(map[string]any{"at": off})
}
