# Proposal: CodeGraph Version-Aware Upgrade/Reconcile

## Intent

CodeGraph is currently installed once via `npm install -g @colbymchenry/codegraph@latest` and then never updated. `CodeGraphReconcileSatisfied()` in `internal/components/communitytool/tool.go` early-returns on subsequent `gentle-ai install`/`sync` runs, so users stay pinned to the first version they ever installed. This change makes CodeGraph version-aware: detect the installed CLI version, compare it to the npm registry latest, and auto-upgrade during sync when outdated.

## Scope

### In Scope
- Phase 1 (already complete and uncommitted): npm registry fetcher, CodeGraph registered in `internal/update`, npm-global upgrade strategy, detection routing, tests.
- Phase 2: sync auto-upgrade step that reinstalls `@latest` and re-runs target-aware `codegraph install` when installed < latest; respects `--dry-run`; reports changes per the sync contract.
- Phase 3: `--force-community-tools` flag and install-side bypass of the `CodeGraphReconcileSatisfied` early-return.
- Phase 4: TUI surfacing of installed→latest version in the update registry and Community Tools screen.

### Out of Scope
- Replacing npm with pnpm/yarn as the default installer for CodeGraph.
- Changing the signed-release manifest pipeline used for GitHub-release tools.
- Version-aware handling for community tools other than CodeGraph.

## Capabilities

### New Capabilities
- `codegraph-version-aware-upgrade`: Detect outdated CodeGraph by comparing the installed CLI version to the npm registry latest, auto-upgrade during sync, allow forced reinstall, and surface the status in the TUI.

### Modified Capabilities
- None

## Approach

Extend the existing `internal/update` framework. Use the npm registry fetcher built in Phase 1 to obtain the latest version. Wire a new `codeGraphUpgradeSyncStep` into `syncRuntime.stagePlan` (`internal/cli/sync.go`) that invokes the npm-global upgrade strategy and reuses `detectedCodeGraphTargets` from `codegraph_contract.go`. Add a `--force-community-tools` sync/install flag that short-circuits `CodeGraphReconcileSatisfied`. TUI screens consume the existing update registry hints to show installed and latest versions.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/update/types.go` | Modified | `InstallNpmGlobal` install method. |
| `internal/update/registry.go` | Modified | CodeGraph registered for updates. |
| `internal/update/npm.go` | New | npm registry latest-version fetcher. |
| `internal/update/check.go`, `detect.go`, `instructions.go` | Modified | Detection routing and update hint. |
| `internal/update/upgrade/strategy.go`, `executor.go` | Modified | npm-global strategy and executor short-circuit. |
| `internal/update/...` tests | New | `npm_test.go`, `upgrade/npm_global_test.go`. |
| `internal/cli/sync.go` | Modified | `codeGraphUpgradeSyncStep` in `syncRuntime.stagePlan`. |
| `internal/components/communitytool/tool.go` | Modified | Bypass satisfied early-return when forced. |
| `internal/components/communitytool/codegraph_contract.go` | Read | `detectedCodeGraphTargets` reused for re-wiring. |
| `internal/cli` flags | Modified | `--force-community-tools` added to sync/install flags. |
| `internal/tui` update/community screens | Modified | Display installed→latest version. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| npm registry unreachable or rate-limited | Medium | Cache the latest version per run; on failure keep the existing install, log a warning, and do not fail sync. |
| Reinstall leaves CodeGraph in a broken state | Low | Capture the current version and targets before reinstall; provide a rollback reinstall path. |
| `--force-community-tools` scope confusion | Low | Flag affects CodeGraph only in this change; rename before expanding to other tools. |
| Auto-upgrade feels intrusive | Medium | Only upgrade when installed < latest; `--dry-run` reports without changing anything. |

## Rollback Plan

Before auto-upgrading, capture the currently installed CodeGraph version (`codegraph --version`) and the detected target list. If the upgrade fails, reinstall `@colbymchenry/codegraph@<captured_version>` globally and rerun `codegraph install --target <targets> --location global --yes`. Reuse the snapshot pattern already present in `InstallWithHome` for any home-directory files touched during re-wiring. Users can always perform the same rollback manually with npm.

## Dependencies

- npm CLI available on PATH.
- Outbound HTTPS to `registry.npmjs.org`.

## Success Criteria

- [ ] `go test ./...` passes, including new and existing `internal/update/...` tests.
- [ ] `gentle-ai sync` upgrades CodeGraph when the installed version is older than the npm registry latest.
- [ ] `gentle-ai sync --dry-run` reports the pending CodeGraph upgrade without modifying the installation.
- [ ] `gentle-ai install --force-community-tools` reinstalls CodeGraph even when the installed version already matches latest.
- [ ] The update registry and Community Tools TUI screens show installed and latest CodeGraph versions.
