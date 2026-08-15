# Race Detector Limitation on Windows (MinGW32)

## Overview

When running Go tests with the race detector enabled (`go test -race ./...`) on 64-bit Windows environments using legacy 32-bit MinGW (`MinGW.org`), compilation fails during runtime race detection code generation.

## Exact Error Message

```text
cc1.exe: sorry, unimplemented: 64-bit mode not compiled in
```

## Root Cause Analysis

- **GCC Version / Target**: `MinGW.org GCC-6.3.0-1` (`mingw32` target)
- **Limitation**: The Go race detector (`-race`) requires CGO and ThreadSanitizer (TSan) runtime support, which requires a 64-bit C compiler target when building on 64-bit Windows (`windows/amd64`). The 32-bit `mingw32` GCC release lacks 64-bit mode code generation (`-m64`), causing `cc1.exe` to fail when Go passes 64-bit architecture options.

## Resolution Strategy

To execute race detection cleanly without toolchain errors:

1. **Linux CI Pipeline**: Run `-race` tests on `ubuntu-latest` in CI/CD where 64-bit GCC and CGO are natively available.
2. **MSYS2 / MinGW-w64**: Upgrade local Windows environment to 64-bit `MinGW-w64` toolchain which supports 64-bit mode.

## CI Configuration Location

Race detection is enabled for Linux in the GitHub Actions workflow:
- **File**: `.github/workflows/ci.yml`
- **Linux Job**: `race-test-linux` (`ubuntu-latest`) executes `go test -race ./...`
- **Windows Job**: `build-test-windows` (`windows-latest`) executes `go test ./...` without `-race` to avoid MinGW32 toolchain failure.

## Manual Workaround for Local Windows Development

If you require `go test -race ./...` locally on Windows:

1. Install 64-bit **MinGW-w64** from [https://www.mingw-w64.org/](https://www.mingw-w64.org/) or via MSYS2 (`pacman -S mingw-w64-x86_64-gcc`).
2. Update system `%PATH%` so that the 64-bit `gcc.exe` binary precedes any 32-bit MinGW installation.
3. Verify the compiler target by running:
   ```cmd
   gcc -v
   ```
   Confirm that `Target:` displays `x86_64-w64-mingw32` instead of `mingw32`.
4. Run `go test -race ./...`.
