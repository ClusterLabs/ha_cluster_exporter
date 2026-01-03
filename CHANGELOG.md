# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Added `--collector.<name>` flags to enable/disable specific collectors.
- Added `--collector.timeout` flag to configure the execution timeout for external commands (default: 10s).
- Implemented global timeout context for all external command executions (`crm_mon`, `cibadmin`, `corosync-*`, `sbd`, `drbdsetup`).
- Added graceful failure: exporter now starts even if collector binaries are missing (logs a warning).
- Added `Dockerfile` for containerized environments.
- Added `docker` and `lint` targets to `Makefile`.

### Changed
- Refactored project structure to follow Standard Go Project Layout (`cmd/`, `internal/`).
- Migrated to **Go 1.24** and toolchain **1.24.11**.
- Migrated logging to **`log/slog`** (standard library).
- Replaced `pkg/errors` with standard library **`errors`** and **`fmt.Errorf`**.
- Refactored `main.go` to remove deprecated code.
- Removed deprecated flags: `--address`, `--port`, `--log-level`, `--enable-timestamps`.
- Improved help output by removing clutter.
- Fixed and updated all unit tests to match new structure and signatures.

