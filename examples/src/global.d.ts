// Type definitions for Go WebAssembly runtime (wasm_exec.js)
// Based on Go 1.25.5 wasm_exec.js

declare global {
  /**
   * The Go class provides the runtime environment for running Go-compiled WebAssembly modules.
   * It handles memory management, JavaScript interop, and execution lifecycle.
   */
  class Go {
    /**
     * Command line arguments passed to the Go program.
     * Defaults to ["js"].
     */
    argv: string[];

    /**
     * Environment variables for the Go program.
     */
    env: Record<string, string>;

    /**
     * Exit handler function called when the Go program exits.
     * @param code The exit code from the Go program.
     */
    exit: (code: number) => void;

    /**
     * WebAssembly import object containing all the functions that Go needs
     * from the JavaScript environment.
     */
    importObject: WebAssembly.Imports;

    /**
     * Whether the Go program has exited.
     */
    exited: boolean;

    /**
     * Internal memory view of the WebAssembly instance.
     */
    mem: DataView;

    /**
     * Internal WebAssembly instance reference.
     */
    _inst: WebAssembly.Instance;

    /**
     * Internal values array for JavaScript object references.
     */
    _values: any[];

    /**
     * Internal reference counts for JavaScript objects.
     */
    _goRefCounts: number[];

    /**
     * Internal mapping from JavaScript values to reference IDs.
     */
    _ids: Map<any, number>;

    /**
     * Internal pool of unused reference IDs.
     */
    _idPool: number[];

    /**
     * Internal promise for exit resolution.
     */
    _exitPromise: Promise<void>;

    /**
     * Internal exit promise resolver.
     */
    _resolveExitPromise: (value?: void) => void;

    /**
     * Internal pending event for async operations.
     */
    _pendingEvent: any;

    /**
     * Internal scheduled timeouts map.
     */
    _scheduledTimeouts: Map<number, number>;

    /**
     * Internal next callback timeout ID.
     */
    _nextCallbackTimeoutID: number;

    /**
     * Creates a new Go runtime instance.
     */
    constructor();

    /**
     * Runs the WebAssembly instance.
     * @param instance The WebAssembly instance to run.
     * @returns A promise that resolves when the Go program exits.
     */
    run(instance: WebAssembly.Instance): Promise<void>;

    /**
     * Internal method to resume execution after an async operation.
     */
    _resume(): void;

    /**
     * Internal method to create function wrappers for Go callbacks.
     * @param id The callback ID.
     * @returns A wrapped function.
     */
    _makeFuncWrapper(id: number): Function;
  }

  /**
   * Global fs object providing file system operations for the Go runtime.
   * Most operations are no-ops or minimal implementations for WebAssembly.
   */
  const fs: {
    constants: {
      O_WRONLY: number;
      O_RDWR: number;
      O_CREAT: number;
      O_TRUNC: number;
      O_APPEND: number;
      O_EXCL: number;
      O_DIRECTORY: number;
    };
    writeSync(fd: number, buf: Uint8Array): number;
    write(fd: number, buf: Uint8Array, offset: number, length: number, position: number | null, callback: (err: Error | null, n?: number) => void): void;
    chmod(path: string, mode: number, callback: (err: Error) => void): void;
    chown(path: string, uid: number, gid: number, callback: (err: Error) => void): void;
    close(fd: number, callback: (err: Error | null) => void): void;
    fchmod(fd: number, mode: number, callback: (err: Error) => void): void;
    fchown(fd: number, uid: number, gid: number, callback: (err: Error) => void): void;
    fstat(fd: number, callback: (err: Error) => void): void;
    fsync(fd: number, callback: (err: Error | null) => void): void;
    ftruncate(fd: number, length: number, callback: (err: Error) => void): void;
    lchown(path: string, uid: number, gid: number, callback: (err: Error) => void): void;
    link(path: string, link: string, callback: (err: Error) => void): void;
    lstat(path: string, callback: (err: Error) => void): void;
    mkdir(path: string, perm: number, callback: (err: Error) => void): void;
    open(path: string, flags: number, mode: number, callback: (err: Error) => void): void;
    read(fd: number, buffer: Uint8Array, offset: number, length: number, position: number | null, callback: (err: Error) => void): void;
    readdir(path: string, callback: (err: Error) => void): void;
    readlink(path: string, callback: (err: Error) => void): void;
    rename(from: string, to: string, callback: (err: Error) => void): void;
    rmdir(path: string, callback: (err: Error) => void): void;
    stat(path: string, callback: (err: Error) => void): void;
    symlink(path: string, link: string, callback: (err: Error) => void): void;
    truncate(path: string, length: number, callback: (err: Error) => void): void;
    unlink(path: string, callback: (err: Error) => void): void;
    utimes(path: string, atime: number, mtime: number, callback: (err: Error) => void): void;
  };

  /**
   * Global process object providing process-related operations for the Go runtime.
   */
  const process: {
    getuid(): number;
    getgid(): number;
    geteuid(): number;
    getegid(): number;
    getgroups(): never;
    pid: number;
    ppid: number;
    umask(): never;
    cwd(): never;
    chdir(): never;
  };

  /**
   * Global path object providing path operations for the Go runtime.
   */
  const path: {
    resolve(...pathSegments: string[]): string;
  };

  /**
   * WebAssembly module result type for instantiateStreaming.
   */
  interface WebAssemblyInstantiateResult {
    instance: WebAssembly.Instance;
    module: WebAssembly.Module;
  }

  /**
   * Extended WebAssembly namespace with streaming instantiation.
   */
  namespace WebAssembly {
    function instantiateStreaming(
      source: Promise<Response>,
      importObject?: Imports
    ): Promise<WebAssemblyInstantiateResult>;
  }
}

export {};