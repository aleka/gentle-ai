# Tasks: CodeGraph Version-Aware Upgrade/Reconcile

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 3a ~250, 3b ~220, 4 ~200 (total ~670) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR3a → PR3b → PR4 (stacked to main, auto-chain) |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|------|------|-----------|----------------------|-----------------|-------------------|
| 1 (PR3a) | Community-tool upgrade/rollback core | PR3a | `go test ./internal/components/communitytool/...` | Fake `Runner`/`Detector`; `t.TempDir()` | `tool.go`, `upgrade_test.go` only |
| 2 (PR3b) | Sync auto-upgrade step + dry-run/reporting | PR3b | `go test ./internal/cli/... -run CodeGraph` | `httptest` npm registry | `sync.go`, `sync_test.go` only |
| 3 (PR4) | Force flags + TUI version surfacing | PR4 | `go test ./internal/tui/... -run CommunityTools` | Direct `Model.Update()` + `RenderCommunityTools` | `install.go`, `run.go`, `model/selection.go`, TUI files only |

## Phase 1: Foundation (already implemented)

- [x] 1.1 `internal/update/npm.go` + `registry.go` + `check.go` — CodeGraph registered for version checks via `CheckFiltered` (npm registry fetcher + `update.UpgradeNpmGlobal` strategy). **Done.**
- [x] 1.2 `internal/update/upgrade/strategy.go` + `executor.go` + tests — manual CodeGraph npm-global upgrade path. **Done.**
- [ ] 1.3 `internal/components/communitytool/codegraph_contract.go` — confirm `codeGraphUpstreamVersion = "1.4.1"` remains an informational contract marker; add doc comment if absent. **Verify before PR3a.**

## Phase 2: PR3a — Community-Tool Core Upgrade/Rollback

- [ ] 2.1 RED test `TestUpgradeCodeGraphWithHome_CapturesAndReinstallsLatest`: fake `Runner` records `npm install -g @colbymchenry/codegraph@latest` and `codegraph install --target ...` commands; create `internal/components/communitytool/upgrade_test.go`.
- [ ] 2.2 Modify `internal/components/communitytool/tool.go` — add `InstallOpts{ForceReinstall bool}`; add `InstallWithHomeOpts` delegating to `InstallWithHome` with zero opts; keep `InstallWithHome` backward-compatible.
- [ ] 2.3 RED test `TestInstallWithHomeOpts_ForceBypassesReconcileSatisfied`: when `ForceReinstall=true`, `CodeGraphReconcileSatisfied` early-return at `tool.go:144` is skipped.
- [ ] 2.4 Implement force bypass in `tool.go` — gate `CodeGraphReconcileSatisfied` early-return with `!opts.ForceReinstall`; propagate `opts` to the full-install path.
- [ ] 2.5 RED test `TestUpgradeCodeGraphWithHome_RollbackReinstallsCapturedVersion`: inject a failing second `codegraph install` command; assert rollback restores snapshots and runs `npm install -g @colbymchenry/codegraph@<captured>` plus target-aware `codegraph install`.
- [ ] 2.6 RED test `TestUpgradeCodeGraphWithHome_RollbackFailureJoinsManualCommand`: when both upgrade and rollback fail, returned error includes `npm install -g @colbymchenry/codegraph@<captured>` and warns sync continues.
- [ ] 2.7 Implement `UpgradeCodeGraphWithHome` in `tool.go`: capture installed version + detected targets, snapshot `CodeGraphManagedPaths`, run forced full-install `@latest`, on failure restore snapshots and reinstall `@<captured>` + rewire targets.

## Phase 3: PR3b — Sync Auto-Upgrade Step + Reporting

- [ ] 3.1 RED test `TestSyncPlan_IncludesCodeGraphUpgradeBeforeGuidance`: when CodeGraph is selected, `syncRuntime.stagePlan()` includes `sync:community-tool:codegraph-upgrade` before `sync:community-tool:codegraph-guidance`.
- [ ] 3.2 Add `CodeGraphUpgradeOutcome` struct to `internal/cli/sync.go` and extend `SyncResult` with `CodeGraphUpgrade *CodeGraphUpgradeOutcome`.
- [ ] 3.3 Implement `codeGraphUpgradeSyncStep` in `sync.go` — call `update.CheckFiltered` filtered to `"codegraph"`; on `UpdateAvailable` invoke `communitytool.UpgradeCodeGraphWithHome`.
- [ ] 3.4 RED test `TestSyncCodeGraphUpgrade_DryRunReportsPending`: with `--dry-run`, `RunSync` returns `CodeGraphUpgradeOutcome.Pending=true` and no install commands run; `NoOp` remains false if a pending upgrade exists.
- [ ] 3.5 Implement dry-run branch in `RunSync` (`sync.go:1497`) — perform version check, populate `CodeGraphUpgradeOutcome`, append managed paths as candidates, execute nothing.
- [ ] 3.6 RED test `TestSyncCodeGraphUpgrade_RegistryDownWarnsAndContinues`: `httptest` registry returns 500; assert `Warning` is set, CLI unchanged, sync succeeds.
- [ ] 3.7 Implement resilient registry failure in `codeGraphUpgradeSyncStep`: warn, preserve install, do not fail sync (per decision matrix).
- [ ] 3.8 RED test `TestRenderSyncReport_ShowsCodeGraphUpgradeOutcome`: `RenderSyncReport` prints performed/pending/warning lines for CodeGraph.
- [ ] 3.9 Update `RenderSyncReport` in `sync.go` to render `CodeGraphUpgrade` outcome (performed, pending, warning, from/to versions).

## Phase 4: PR4 — Force Flags + TUI Version Surfacing

- [ ] 4.1 RED test `TestParseSyncFlags_ForceCommunityTools`: `gentle-ai sync --force-community-tools` sets `SyncFlags.ForceCommunityTools=true` and `forceCommunityToolsSet=true`.
- [ ] 4.2 RED test `TestParseInstallFlags_ForceCommunityTools`: `gentle-ai install --force-community-tools` sets `InstallFlags.ForceCommunityTools=true`.
- [ ] 4.3 Add `ForceCommunityTools` + `forceCommunityToolsSet` to `SyncFlags`/`InstallFlags` and `model.Selection.ForceCommunityTools` in `internal/model/selection.go`.
- [ ] 4.4 Thread flag: `RunSync` → `BuildSyncSelection`; `RunInstall` → `NormalizeInstallFlags` → `buildStagePlan` → `communityToolInstallStep`.
- [ ] 4.5 RED test `TestCommunityToolInstallStep_PassesForceReinstall`: when `Selection.ForceCommunityTools` is true, `communityToolInstallStep.Run()` calls `InstallWithHomeOpts` with `ForceReinstall=true`.
- [ ] 4.6 Modify `communityToolInstallStep` in `internal/cli/run.go` to use `InstallWithHomeOpts` with `ForceReinstall` when the flag is set.
- [ ] 4.7 RED test `TestRenderCommunityTools_ShowsVersionPair`: when `UpdateResults` contains a CodeGraph update hint, `screens.RenderCommunityTools` renders "CodeGraph 1.4.1 → 1.5.0".
- [ ] 4.8 Modify `internal/tui/screens/community_tools.go` — add version line from `UpdateResults` (SHOULD-level); read `InstalledVersion`/`LatestVersion` for `CommunityToolCodeGraph`.
- [ ] 4.9 RED test `TestWelcomeBannerSummaryLine_IncludesCodeGraphVersionPair`: at `internal/tui/model.go:1098`, `update.UpdateSummaryLine(m.UpdateResults)` includes `codegraph X -> Y` when an update is available.

