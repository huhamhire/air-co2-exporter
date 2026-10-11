# Maintenance Guidelines

## Scope and language

These instructions apply to the entire repository. Write documentation, code comments,
commit messages, and pull request descriptions in English. Preserve unrelated local
changes and keep maintenance work focused on the requested scope.

## Commit messages

- Follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/).
- Use `<type>[optional scope][!]: <description>` with an English description.
  Use `feat` for new features and `fix` for bug fixes; use `build`, `ci`, `docs`,
  `refactor`, `perf`, `style`, `test`, or `chore` for the corresponding maintenance work.
- Keep commits focused and choose a scope only when it clarifies the affected component.
- Separate the optional body and footers from the subject with blank lines.
- Mark breaking changes with `!` before the colon or a `BREAKING CHANGE:` footer.
  Explain the compatibility change and required migration, including toolchain
  requirements, public APIs, CLI behavior, metrics, or log formats when affected.

## Project structure

```text
air-co2-exporter/
├── .github/
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md       # Bug report template
│   │   └── feature_request.md  # Feature request template
│   └── workflows/
│       ├── ci.yml              # Go checks, tests, and container validation
│       └── release.yml         # Tag validation, native release packages, and release drafts
├── monitor/                    # USB access and sensor protocol
│   ├── config.go               # USB initialization parameters and measurement types
│   ├── decrypt.go              # Packet decryption and checksum validation
│   ├── device.go               # USB vendor and product IDs
│   ├── monitor.go              # Device lifecycle, packet reads, and sensor decoding
│   └── monitor_test.go         # Sensor protocol regression tests
├── prometheus/                 # Metrics and HTTP handlers
│   ├── exporter.go             # Metric definitions, registry, and HTTP handlers
│   └── exporter_test.go        # Metric exposition regression tests
├── .dockerignore               # Docker build context exclusions
├── .gitignore                  # Git working tree exclusions
├── AGENTS.md                   # Repository maintenance guidelines
├── CHANGELOG                   # Release notes, including unreleased changes
├── Dockerfile                  # Container build and runtime images
├── go.mod                      # Go version requirement and module dependencies
├── go.sum                      # Module dependency checksums
├── LICENSE                     # Project license
├── main.go                     # CLI, logging, polling, HTTP serving, and process lifecycle
├── makefile                    # Build, validation, and packaging commands
├── README.md                   # Installation and usage documentation
└── VERSION                     # Next planned release version, currently 0.4.0
```

Keep this tree and its descriptions in sync when maintained files or directories
are added, renamed, removed, or assigned new responsibilities.

## Toolchain and dependencies

- Require Go 1.26 or later. Validate the minimum version and latest stable Go release in CI.
- Use Go modules. Do not require a GOPATH checkout or install tools with `go get`.
- USB access requires CGO, a native C compiler, pkg-config, and libusb development files.
  Linux amd64 and arm64 are supported deployment targets. Cross-compiling requires
  a matching C compiler and libusb libraries; `GOOS`/`GOARCH` alone are insufficient.
- Prefer current stable upstream dependencies. Review API and release changes,
  run `go get -u ./...` and `go mod tidy`, and commit `go.mod` and `go.sum` together.
- Never fabricate checksums or disable checksum verification to resolve an upgrade.
- Keep the module minimum, CI matrix, Docker Go version, and README consistent.

## Go conventions

- Follow [Effective Go](https://go.dev/doc/effective_go) and the
  [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
  Run `gofmt` on changed Go files.
- Group standard library imports separately from local and third-party imports.
- Use `log/slog` and Prometheus
  [promslog](https://pkg.go.dev/github.com/prometheus/common/promslog) for structured logging.
- Prefer simple control flow, meaningful errors, and standard library APIs.
  Use `go fix ./...` on toolchain upgrades and review its changes.
- Use contexts for cancellation and bounded I/O. Stop workers before releasing USB resources.
- Check acquisition errors and release interfaces, devices, and USB contexts in reverse order.
- Validate packet lengths before indexing. Cover decoding and checksum behavior with tests.
- Preserve CLI flags, defaults, `LABEL_TAG`, port `9110`, `/metrics`, metric names
  (`air_temp`, `air_co2`, `air_rh`), units, and the `tag` label unless a requested
  behavior change explicitly includes a migration plan.

## Validation

Install Linux build prerequisites with:

```sh
sudo apt-get update
sudo apt-get install -y build-essential pkg-config libusb-1.0-0-dev
```

Before submitting code or dependency changes, run:

```sh
go mod tidy
go mod verify
go vet ./...
go test -race ./...
make build
./bin/air_co2_exporter --help
./bin/air_co2_exporter --version
docker build -t air-co2-exporter:check .
```

Confirm tidy leaves no further module changes and `gofmt -l` lists no Go files.
Add focused tests for changed behavior. Unit tests must not require USB hardware.
Hardware checks are separate: verify real readings, labels, and clean termination
with a supported sensor. Report unavailable checks honestly; a CGO-disabled build
does not validate USB support.

## Release preparation

- Develop the next release on `feature/next` when requested.
- Keep `VERSION` and `CHANGELOG` aligned. Do not mark a release published prematurely.
- Describe compatibility changes, validation results, and outstanding hardware checks.
- Do not create release tags, publish images, or deploy without explicit authorization.
- After merging the release PR and passing CI, create an annotated `v<VERSION>` tag
  on the intended `master` commit. The release workflow rejects tags that do not match
  `VERSION` or whose commits are not contained in `master`.
- Tag pushes build native Linux amd64 and arm64 archives, generate SHA-256 checksums,
  and create a GitHub Release draft using the matching CHANGELOG section.
  Review the draft and validation results before publishing. Docker image publication
  remains a separate release action.
