# ob-operator agent instructions

## DIMA workflow

When asked to handle a DIMA work item, use
[ob-operator-dima](.agents/skills/ob-operator-dima/SKILL.md). This includes a
`project.alipay.com` work-item URL, `openWorkItemId`, or an ID identified by the
user as a DIMA item. Respect requests for analysis only, no submission, or other
scope limits. Discussing or editing the workflow itself does not start a bugfix.

The workflow covers reading the item and linked documents, RED/GREEN regression
coverage, three independent subagents that each execute tests and review the fix,
submission within the user's scope, and verified DIMA traceability. After the
three test passes and authorized submission are verified, set the work item to
Fixed and read back its status. This is standing user authorization; do not ask
again. Explicit analysis-only, no-write, or other scope limits still take priority.
If the user requests a direct push without a PR/MR, honor that scope; a missing
PR/MR is not a blocker for the three tests or the Fixed transition.
Use the skill for the precise gates and completion criteria.

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

## Hook configuration

`.codex/hooks.json` and `.claude/settings.json` route DIMA prompts to the same
skill. The hook only supplies context; it neither runs tests nor writes to DIMA.
Keep hook configuration and user trust settings intact unless the task concerns
them. Do not automatically change global trust or bypass hook trust checks.
