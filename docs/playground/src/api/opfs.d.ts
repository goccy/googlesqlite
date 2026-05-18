// opfs.d.ts declares the OPFS synchronous-access-handle types that the
// TypeScript 5.6 DOM library does not yet ship. They are part of the
// File System Access API and are available to Web Workers in a secure
// context; worker.ts uses them for database persistence.

interface FileSystemReadWriteOptions {
  at?: number
}

interface FileSystemSyncAccessHandle {
  read(buffer: ArrayBufferView, options?: FileSystemReadWriteOptions): number
  write(buffer: ArrayBufferView, options?: FileSystemReadWriteOptions): number
  truncate(newSize: number): void
  getSize(): number
  flush(): void
  close(): void
}

// createSyncAccessHandle is an OPFS-only addition to FileSystemFileHandle;
// merge it into the lib's interface.
interface FileSystemFileHandle {
  createSyncAccessHandle(): Promise<FileSystemSyncAccessHandle>
}
