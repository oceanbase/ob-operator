# ob-operator agent instructions

## Repository facts

- Go version: read `go.mod`; build and test recipes: `Makefile`, `make/`.
- Operator entry point: `cmd/operator/`; reconciliation: `internal/controller/`;
  resource lifecycle/tasks: `internal/resource/`; CRDs/webhooks: `api/`.
- Shared helpers, Kubernetes clients, and task framework: `pkg/`.
- Dashboard backend: `internal/dashboard/`; frontend: `ui/`; CLI: `internal/cli/`.
- Go tests live alongside the code; Kubernetes scenarios live in `tests/`.
  Code and regression tests normally belong in the same ob-operator PR.
- Inspect `git remote -v` and remote branches before submission. The current
  upstream is GitHub `oceanbase/ob-operator`, with `master`, `X.Y.Z_release`,
  `ds_X.Y.Z_release`, and `cli_X.Y.Z_release` branches. Derive the target from
  the affected component/version and actual remote refs, not OBD's `dev_X.Y.Z`.

## Build and test details

- CI generates dashboard assets with `make dashboard-bindata-gen dashboard-doc-gen`
  before vendoring, building, and testing. Check the target branch's workflows.
- `make test` runs webhook, helper/task, and commands/config tests. It does not
  cover all controllers, resource managers, dashboard, CLI, or telemetry tests.
- Read a package's suite setup before running it. `internal/resource/` uses
  `UseExistingCluster`; some SDK/client tests also need external services.
  Run those only against a known, authorized test environment.
- Run Shell scenarios from `tests/` because they source files relative to the
  working directory. Some share `env_vars.sh`; preserve ordering and avoid
  concurrent runs that overwrite their state. Check assertions/logs as well as
  exit codes: `tests/run.sh` does not aggregate failures reliably.
- Generated manifests/code and `make fmt` can change tracked files. Review the
  resulting diff and include only changes needed by the task.
