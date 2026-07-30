# Apply Progress: CodeGraph Version-Aware Upgrade/Reconcile

> Cumulative artifact — each apply batch MERGES its results here. Do not delete
> entries from prior batches.

## Batch 1 — PR3a: Community-Tool Core Upgrade/Rollback (2026-07-29)

**Branch**: `feat/984-pr3a-communitytool-upgrade` (from `feat/984-pr2-npm-global-strategy`)
**Mode**: Strict TDD (`go test ./...`)
**Delivery**: auto-chain / feature-branch-chain; PR3a targets PR2 branch
**Tasks completed**: 1.3, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7 (8/8 assigned)

### Completed Tasks

- [x] 1.3 `codeGraphUpstreamVersion = "1.4.1"` verified informational (only referenced by its own contract test); doc comment added stating it is not a pin.
- [x] 2.1 RED `TestUpgradeCodeGraphWithHome_CapturesAndReinstallsLatest` → GREEN
- [x] 2.2 `InstallOpts{ForceReinstall}` + `InstallWithHomeOpts`; `InstallWithHome` delegates zero opts (backward-compatible — full package suite green)
- [x] 2.3 RED `TestInstallWithHomeOpts_ForceBypassesReconcileSatisfied` → GREEN (triangulated: zero-opts satisfied install stays untouched)
- [x] 2.4 Early-return gated with `!opts.ForceReinstall`; force rebuilds `npm install -g @latest` + target rewire even when CLI present
- [x] 2.5 RED `TestUpgradeCodeGraphWithHome_RollbackReinstallsCapturedVersion` → GREEN (snapshot restore + `@1.4.1` pin reinstall + target rewire asserted)
- [x] 2.6 RED `TestUpgradeCodeGraphWithHome_RollbackFailureJoinsManualCommand` → GREEN (joined error carries manual `npm install -g @colbymchenry/codegraph@1.4.1` + "sync continues")
- [x] 2.7 `UpgradeCodeGraphWithHome`: capture version (`codeGraphInstalledVersion` package var) + targets → snapshot `CodeGraphManagedPaths` → forced full-install `@latest` → rollback on failure

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.3 | `codegraph_contract_test.go` (existing) | Unit | ✅ package green | N/A (doc comment, structural) | ✅ package green | ➖ Skipped: constant doc, one possible output | ✅ Clean |
| 2.1/2.2/2.7 | `upgrade_test.go` | Unit (fake Runner/Detector, `t.TempDir()`) | ✅ package green | ✅ Written (compile-undefined) | ✅ Passed | ✅ npm+target command pair asserts non-trivial sequence | ✅ Clean |
| 2.3/2.4 | `upgrade_test.go` | Unit | ✅ package green | ✅ Written | ✅ Passed | ✅ 2 cases: force bypasses / no-force preserves | ✅ Clean |
| 2.5/2.7 | `upgrade_test.go` | Unit | ✅ package green | ✅ Written | ✅ Passed | ✅ rollback command sequence + restored file content asserted | ✅ `rollbackCodeGraphUpgrade` extracted |
| 2.6 | `upgrade_test.go` | Unit | ✅ package green | ✅ Written | ✅ Passed | ✅ rollback-failure path (distinct from 2.5 success path) | ✅ `rollbackPin` extracted |

### Test Summary

- **Total tests written**: 4 (all passing)
- **Layers used**: Unit (4), Integration (0), E2E (0)
- **Approval tests** (refactoring): None — `InstallWithHome` body moved intact behind `InstallWithHomeOpts`; existing suite (safety net) proves behavior preserved
- **Pure functions created**: 2 (`rollbackPin`, `detectCodeGraphCLIVersion`)

### Work Unit Evidence

| Evidence | Value |
|---|---|
| Focused test command | `go test ./internal/components/communitytool/ -run 'TestUpgradeCodeGraphWithHome\|TestInstallWithHomeOpts' -count=1` → `ok ... 0.025s` (4/4 PASS) |
| Full package gate | `go test ./internal/components/communitytool/... -count=1` → `ok 0.343s`; `go test ./internal/update/... -count=1` → `ok` both packages |
| Runtime harness | Fake `Runner`/`Detector` + `t.TempDir()` fixtures per tasks.md Unit 1 row; no live npm/codegraph subprocess (unit boundary) |
| Static gates | `go build ./...` ✅, `go vet ./internal/components/communitytool/...` ✅, `go run ./internal/gofmtcheck` ✅ (no output) |
| Rollback boundary | `internal/components/communitytool/tool.go`, `upgrade_test.go`, `codegraph_contract.go` only — revert these three files to remove this unit without touching PR1/PR2 work |

### Files Changed

| File | Action | What |
|------|--------|------|
| `internal/components/communitytool/tool.go` | Modified | `InstallOpts`, `InstallWithHomeOpts` (+ delegation), force gate + force full-install commands, `UpgradeCodeGraphWithHome`, `rollbackCodeGraphUpgrade`, `rollbackPin`, `codeGraphInstalledVersion` capture var |
| `internal/components/communitytool/upgrade_test.go` | Created | 4 RED-first tests for upgrade/rollback/force |
| `internal/components/communitytool/codegraph_contract.go` | Modified | Doc comment on `codeGraphUpstreamVersion` (task 1.3) |
| `openspec/changes/codegraph-version-aware-upgrade/tasks.md` | Modified | Checked 1.3, 2.1–2.7 |

### Deviations from Design

None — implementation matches design.md Decision 3 seam exactly. Note: tasks.md 2.2 wording says "`InstallWithHomeOpts` delegating to `InstallWithHome`"; design.md (authoritative) says `InstallWithHome` delegates zero opts to `InstallWithHomeOpts` — implemented per design (opts would be lost otherwise).

### Issues Found

None. Pre-existing: none observed in communitytool/update packages.

### Remaining Tasks (next batches)

- [x] 3.1–3.9 PR3b — sync auto-upgrade step + dry-run/reporting (`internal/cli/sync.go`) — delivered as chained PR3b-1 + PR3b-2 (see Batch 2)
- [ ] 4.1–4.9 PR4 — `--force-community-tools` flags + TUI version surfacing

### Chain State

| Branch | Commit | Targets |
|---|---|---|
| `feat/984-codegraph-version-aware` (tracker) | `008b77a6` docs(sdd): add codegraph-version-aware-upgrade change artifacts | upstream/main |
| `feat/984-pr1-codegraph-detection` | `659f3728` feat(update): register codegraph with npm registry version detection | tracker |
| `feat/984-pr2-npm-global-strategy` | `1cbb294b` feat(update): add npm-global upgrade strategy for codegraph | PR1 branch |
| `feat/984-pr3a-communitytool-upgrade` | `f731451f` feat(communitytool): add version-aware codegraph upgrade with rollback | PR2 branch |
| `feat/984-pr3b-sync-autoupgrade` | `5d567d64` feat(cli): auto-upgrade outdated codegraph during sync | PR3a branch — **superseded: kept as pre-split backup, not for review** |
| `feat/984-pr3b1-sync-upgrade-step` | `40dde85e` feat(cli): auto-upgrade outdated codegraph during sync | PR3a branch |
| `feat/984-pr3b2-sync-report-tests` | `47cbaaf6` feat(cli): render codegraph upgrade outcome in sync report | PR3b-1 branch |
| `feat/984-pr4a-install-force-flags` | `8079e43d` feat(cli): add --force-community-tools install flag threading | PR3b-2 branch |
| `feat/984-pr4b-sync-force-gate` | `71800ca0` feat(cli): force community tools reinstall during sync | PR4a branch |
| `feat/984-pr4-force-flags-tui` | (Batch 3 tip) feat(tui): surface codegraph version pair and forced reinstall lines | PR4b branch |

## Batch 2 — PR3b: Sync Auto-Upgrade Step + Dry-Run/Reporting (2026-07-29)

> **SPLIT NOTE (2026-07-30)**: PR3b (`feat/984-pr3b-sync-autoupgrade`, +706 lines) exceeded the 400-line review budget and was split into two chained PRs without losing any code, test, or doc content:
> - **PR3b-1** `feat/984-pr3b1-sync-upgrade-step` (from PR3a): outcome struct, `codeGraphUpgradeSyncStep`, dry-run branch, resilience logic, `SetNpmRegistryBaseURL` seam, and tests for 3.1/3.3/3.4/3.6 (+479 lines — see Deviation 5).
> - **PR3b-2** `feat/984-pr3b2-sync-report-tests` (from PR3b-1): `renderCodeGraphUpgradeOutcome` + `RenderSyncReport` wiring (3.8/3.9), the NoOp-honesty full-path test (D5), and these openspec docs.
> The tip of PR3b-2 is content-identical to the original PR3b branch for all changed files (verified: empty `git diff` on `internal/`; openspec differs only by this split note). The original branch is retained as backup only.

**Branch**: `feat/984-pr3b-sync-autoupgrade` (from `feat/984-pr3a-communitytool-upgrade`)
**Mode**: Strict TDD (`go test ./...`)
**Delivery**: auto-chain / feature-branch-chain; PR3b targets PR3a branch
**Tasks completed**: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9 (9/9 assigned)

### Completed Tasks

- [x] 3.1 RED `TestSyncPlan_IncludesCodeGraphUpgradeBeforeGuidance` → GREEN (step inserted before `codegraph-guidance` in `stagePlan`; triangulated with not-selected omission test)
- [x] 3.2 `CodeGraphUpgradeOutcome{Pending, Performed, RolledBack, From, To, Warning}` + `SyncResult.CodeGraphUpgrade *CodeGraphUpgradeOutcome` (exact design contract)
- [x] 3.3 `codeGraphUpgradeSyncStep` — `codeGraphVersionCheck` seam wraps `update.CheckFiltered(..., []string{"codegraph"})`; on `UpdateAvailable` invokes `upgradeCodeGraphWithHome` seam (`communitytool.UpgradeCodeGraphWithHome`); table test covers full Decision-4 matrix
- [x] 3.4 RED `TestSyncCodeGraphUpgrade_DryRunReportsPending` → GREEN via full `RunSync --dry-run` path (Pending=true, From/To set, zero executions, NoOp=false)
- [x] 3.5 `checkCodeGraphUpgradeDryRun` in the `RunSync` dry-run branch — checks, populates outcome, executes nothing (nil when CodeGraph not selected: no network on unrelated syncs)
- [x] 3.6 RED `TestSyncCodeGraphUpgrade_RegistryDownWarnsAndContinues` → GREEN — real `update.CheckFiltered` against an `httptest` registry returning 500 via new exported seam `update.SetNpmRegistryBaseURL`; Warning set, upgrade never called, sync succeeds
- [x] 3.7 Resilient registry failure in the step: `CheckFailed` → `Warning`, nil error (spec SHALL)
- [x] 3.8 RED `TestRenderSyncReport_ShowsCodeGraphUpgradeOutcome` → GREEN (6-case table incl. nil/zero no-render cases)
- [x] 3.9 `renderCodeGraphUpgradeOutcome` wired into all three `RenderSyncReport` branches (no-op, dry-run, executed) with from/to versions

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 3.1 | `sync_test.go` | Unit (`stagePlan`) | ✅ focused sync/codegraph green | ✅ Failed (step missing) | ✅ Passed | ✅ 2nd case: step omitted when CodeGraph not selected | ✅ Clean |
| 3.2 | (structural) | Unit | ✅ | N/A (struct: one possible shape, covered by 3.3+ usage) | ✅ compile | ➖ Skipped: type/field definition, no logic | ✅ Clean |
| 3.3/3.7 | `sync_test.go` | Unit (step, seam-swapped) | ✅ | ✅ Failed (undefined seams) | ✅ Passed | ✅ 7-case table: performed / registry-fail / up-to-date / not-installed / dev-build / rollback-success / rollback-failure | ✅ candidates fixture fixed (discoverable agent required) |
| 3.4/3.5 | `sync_test.go` | Integration (`RunSync --dry-run`) | ✅ | ✅ Failed (outcome nil) | ✅ Passed | ✅ Pending + no-execute + empty Execution + NoOp=false asserted together | ✅ fixture extracted |
| NoOp (D5) | `sync_test.go` | Integration (2× `RunSync`) | ✅ | ✅ Failed (outcome nil) | ✅ Passed | ✅ mutation-checked: guard disabled → FAIL; restored → PASS | ✅ fixture extracted |
| 3.6 | `sync_test.go` | Integration (`RunSync` + httptest registry) | ✅ | ✅ Failed (`update.SetNpmRegistryBaseURL` undefined) | ✅ Passed | ✅ Warning set AND install untouched (0 upgrade calls) AND sync succeeds | ✅ fixture extracted |
| 3.8/3.9 | `sync_test.go` | Unit (`RenderSyncReport`) | ✅ | ✅ Failed (no lines rendered) | ✅ Passed | ✅ 6 cases: performed/pending/warning-noop/rolledback + nil + zero | ✅ Clean |

### Test Summary

- **Total tests written**: 6 functions (13 scenarios incl. table cases)
- **Layers used**: Unit (2 functions + 1 structural), Integration (3 functions), E2E (0)
- **Approval tests** (refactoring): None — `RenderSyncReport` branches extended additively; existing report tests are the safety net and stay green
- **Pure functions created**: 2 (`codeGraphUpgradeCheckResult`, `renderCodeGraphUpgradeOutcome`)

### Work Unit Evidence

| Evidence | Value |
|---|---|
| Focused test command | `go test ./internal/cli/ -run 'CodeGraph' -count=1` → `ok 1.116s` (all 13 scenarios PASS) |
| Full package gate | `go test ./internal/cli/... -count=1` → `ok 247.652s`; regression: `go test ./internal/components/communitytool/... ./internal/update/...` → `ok` all three packages |
| Runtime harness | `httptest` npm registry (500) via `update.SetNpmRegistryBaseURL` driving the REAL `update.CheckFiltered` path; `RunSync` full pipeline with seam-swapped executor (tasks.md Unit 2 row) |
| Static gates | `go build ./...` ✅, `go vet ./internal/cli/... ./internal/update/...` ✅, `go run ./internal/gofmtcheck` ✅ |
| Rollback boundary | `internal/cli/sync.go`, `internal/cli/sync_test.go`, `internal/update/npm.go` (10-line exported seam) only — revert these to remove this unit without touching PR3a or PR1/PR2 work |

### Files Changed

| File | Action | What |
|------|--------|------|
| `internal/cli/sync.go` | Modified | `CodeGraphUpgradeOutcome` + `SyncResult.CodeGraphUpgrade`; `syncRuntime.codeGraphUpgrade`; `codeGraphUpgradeSyncStep` (+ `Run` with Decision-4 matrix); `codeGraphVersionCheck`/`upgradeCodeGraphWithHome` seams; `codeGraphUpgradeCheckResult`; `checkCodeGraphUpgradeDryRun`; NoOp=false guard; `renderCodeGraphUpgradeOutcome` wired into 3 report branches; stagePlan ordering |
| `internal/cli/sync_test.go` | Modified | 6 test functions (13 scenarios) + `setupCodeGraphSyncHome`/`stubCodeGraphVersionCheck` fixture helpers |
| `internal/update/npm.go` | Modified | Exported `SetNpmRegistryBaseURL` test seam (10 lines, additive only) |
| `openspec/changes/codegraph-version-aware-upgrade/tasks.md` | Modified | Checked 3.1–3.9 |

### Deviations from Design

1. **Decision-4 ambiguity resolved in cli, communitytool untouched** (per orchestrator instruction): PR3a's `UpgradeCodeGraphWithHome` error contract carries "sync continues" in BOTH the rollback-success and rollback-failure messages. The step discriminates on the "rollback failed" marker: rollback-failure → loud (step returns error, pipeline snapshot backstop); anything else → warn-and-continue with `RolledBack=true`. Documented in code at the step.
2. **`SetNpmRegistryBaseURL` exported seam added to `internal/update`** — design's testing strategy ("httptest registry (`npmRegistryBaseURL` var)") requires the var to be reachable from `internal/cli` tests; the var itself stays unexported, a 10-line setter/restore pair is the minimal exported surface.
3. **Line count over forecast**: PR3b actual ≈ 631 insertions vs ~220 forecast. Production diff is 187 lines (within forecast); the remainder is tests — the mandated full-path coverage (RunSync dry-run, double-sync NoOp, httptest registry) plus the 7-case step matrix outweighed the estimate. PR3a showed the same pattern (365 actual vs 250 forecast). Mitigation applied: shared fixture helpers extracted; tests stay with their unit per work-unit-commits. **Post-review update**: the 400-line budget was enforced anyway via the PR3b-1/PR3b-2 split (see Split Note above).
4. **Dry-run "append managed paths as candidates" (3.5)** implemented on the execution step (content-compared via `changedSyncFiles`); in the dry-run branch there is no `changedFiles` consumer, so appending would be a dead write — the check + outcome is the observable dry-run behavior and is what the test asserts.
5. **PR3b-1 lands at +479 lines, over the 400 budget** (2026-07-30 split): the mandated PR3b-1 content (production 152 + seam 10 + tests 317, incl. the 134-line `TestCodeGraphUpgradeSyncStep` Decision-4 matrix assigned to 3.3) arithmetically exceeds 400. Options if the budget is hard: (a) accept `size:exception` — 66% of the diff is tests; (b) move `TestCodeGraphUpgradeSyncStep` to PR3b-2 → 345/361, at the cost of separating the step's unit test from its unit.

### Issues Found

1. **Pre-existing (NOT introduced here)**: the OpenCode CodeGraph reconcile in `codeGraphGuidanceSyncStep` rewrites `~/.config/opencode/opencode.json` on every sync, so a sync with `--agents opencode` + CodeGraph selected is not file-idempotent (`FilesChanged ≥ 1` on rerun). Verified independent of this change (reproduces with the upgrade step reporting UpToDate). Out of PR3b scope — communitytool is frozen; flagged for a future change.
2. **PR3a message ambiguity** (resolved, see Deviation 1): rollback-failure message says "sync continues" although Decision 4 fails sync loudly. Message text left untouched (communitytool frozen, PR3a test 2.6 locks it); the behavioral resolution lives in the cli step.

### Remaining Tasks (next batches)

- [x] 4.1–4.9 PR4 — `--force-community-tools` flags + TUI version surfacing — delivered as chained PR4a + PR4b + PR4 tip (see Batch 3)

## Batch 3 — PR4: Force Flags + TUI Version Surfacing (2026-07-30)

> **SPLIT NOTE (2026-07-30)**: PR4 (forecast ~200 lines) landed at ~725 changed lines and was split proactively into three chained PRs to respect the 400-line budget (same pattern as the PR3b split; tests again outweighed the forecast):
> - **PR4a** `feat/984-pr4a-install-force-flags` (from PR3b-2): install-side flag — `InstallFlags.ForceCommunityTools`, `NormalizeInstallFlags` threading, `model.Selection.ForceCommunityTools`, `communityToolInstallStep.force` + `installCommunityToolWithHomeOpts` seam (tasks 4.2, 4.5, 4.6, install side of 4.3/4.4). **198 lines** (174+24).
> - **PR4b** `feat/984-pr4b-sync-force-gate` (from PR4a): sync-side flag — `SyncFlags.ForceCommunityTools` + `forceCommunityToolsSet`, `BuildSyncSelection` threading, step gate bypass via extracted `executeUpgrade`, dry-run force `Pending` (tasks 4.1, sync side of 4.3/4.4, design data flow "force ─► UpgradeCodeGraphWithHome"). **345 lines** (305+40).
> - **PR4 tip** `feat/984-pr4-force-flags-tui` (from PR4b; requested branch name kept at the tip): user-facing surfacing — forced-reinstall sync-report lines (performed/pending/rolled-back with no from/to pair), Community Tools screen version pair, welcome-banner pair test (tasks 4.7, 4.8, 4.9). **182 lines** (177+5).
> The forced-report renderer lines were sliced from PR4b into the tip to keep PR4b at 345 ≤ 400; within the chain, a forced sync on PR4b alone renders the pre-existing version-pair line with empty versions — transient, fixed at the tip. The tip tree is content-identical to the completed implementation (verified: only tasks.md differs before the docs commit).

**Branch chain**: `feat/984-pr4a-install-force-flags` (`8079e43d`) → `feat/984-pr4b-sync-force-gate` (`71800ca0`) → `feat/984-pr4-force-flags-tui` (`c6fb6c87`)
**Mode**: Strict TDD (`go test ./...`)
**Delivery**: auto-chain / feature-branch-chain; PR4a targets PR3b-2, each child targets its parent
**Tasks completed**: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 4.9 (9/9 assigned)

### Completed Tasks

- [x] 4.1 RED `TestParseSyncFlagsForceCommunityTools` → GREEN (3 cases: absent/explicit/explicit-false; `forceCommunityToolsSet` marker asserted)
- [x] 4.2 RED `TestParseInstallFlagsForceCommunityTools` → GREEN (absent/explicit)
- [x] 4.3 `ForceCommunityTools` + `forceCommunityToolsSet` on `SyncFlags`; `ForceCommunityTools` on `InstallFlags` and `model.Selection` (one-shot flag, never persisted — `RestoreSelection` does not touch it, so no restore guard needed)
- [x] 4.4 Threading both sides with tests: `TestBuildSyncSelectionForceCommunityTools`, `TestNormalizeInstallFlagsForceCommunityTools`, `TestInstallStagePlanThreadsForceCommunityTools`; sync stage plan passes `force` to the upgrade step
- [x] 4.5 RED `TestCommunityToolInstallStepPassesForceReinstall` → GREEN (force true/false triangulated against the opts seam)
- [x] 4.6 Seam renamed to `installCommunityToolWithHomeOpts = communitytool.InstallWithHomeOpts`; step passes `InstallOpts{ForceReinstall: s.force}`; 4 existing stubs updated to the new signature with assertions preserved (approval-test signature update)
- [x] 4.7 RED `TestRenderCommunityToolsShowsVersionPair` → GREEN (4 cases: update available / up-to-date / other tool / nil)
- [x] 4.8 `RenderCommunityTools` gains `updates []update.UpdateResult`; pure helper `codeGraphUpdatePair`; call site passes `m.UpdateResults`
- [x] 4.9 `TestWelcomeBannerSummaryLineIncludesCodeGraphVersionPair` — behavior already existed (Phase 1 registry entry → generic `UpdateSummaryLine`); test locks it at the `model.View()` integration point, plus up-to-date triangulation. No production change needed (as the task allowed)
- [x] "Also" (orchestrator): sync-side force bypasses the satisfied gate — `TestCodeGraphUpgradeSyncStep_ForceBypassesSatisfiedGate` (4 cases: forced perform with 0 registry calls / rollback-success warn / rollback-failure loud / no-force control), full-path `TestSyncCodeGraphUpgrade_ForceRunsUpgradeWhenSatisfied` (upgrade runs under UpToDate, NoOp=false), `TestSyncCodeGraphUpgrade_DryRunForceReportsPending`

### TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 4.1 | `sync_test.go` | Unit (`ParseSyncFlags`) | ✅ focused cli green | ✅ Compile-undefined field | ✅ Passed | ✅ 3 cases incl. `--force-community-tools=false` still sets marker | ✅ Clean |
| 4.2 | `install_test.go` | Unit (`ParseInstallFlags`) | ✅ | ✅ Compile-undefined field | ✅ Passed | ✅ 2 cases | ✅ Clean |
| 4.3/4.4 | `sync_test.go`, `install_test.go`, `run_community_tool_test.go` | Unit (selection/threading) | ✅ | ✅ Compile-undefined field | ✅ Passed | ✅ true/false both directions + stage-plan step inspection | ✅ Clean |
| 4.5/4.6 | `run_community_tool_test.go` | Unit (step, seam-swapped) | ✅ 4 existing step tests | ✅ Compile-undefined seam/field | ✅ Passed | ✅ force=true/false pair; existing stubs assert zero-opts | ✅ Seam renamed to honest `...WithHomeOpts` |
| 4.7/4.8 | `community_tools_test.go` | Unit (`RenderCommunityTools`) | ✅ 2 existing render tests | ✅ Arity mismatch (6th arg) | ✅ Passed | ✅ 4 cases: pair / up-to-date / other tool / nil | ✅ `codeGraphUpdatePair` extracted pure |
| 4.9 | `model_test.go` | Integration (`model.View`) | ✅ tui green | N/A (behavior pre-existed Phase 1; approval test per task note) | ✅ Passed first run | ✅ Up-to-date case asserts no banner pair | ➖ No production change |
| sync force (Also) | `sync_test.go` | Unit (step) + Integration (`RunSync` ×2) | ✅ PR3b matrix green | ✅ Compile-undefined `force` field | ✅ Passed | ✅ 4-case step matrix + full-path + dry-run; `checkCalls==0` proves gate bypass | ✅ `executeUpgrade` extracted (duplication removed), tests green after |

### Test Summary

- **Total tests written**: 10 functions (20 scenarios incl. table cases)
- **Layers used**: Unit (8 functions), Integration (2 functions: full-path `RunSync` force + banner `View`; plus dry-run `RunSync` force), E2E (0)
- **Approval tests** (refactoring): 4 existing `communityToolInstallStep` stubs updated to the renamed opts seam with assertions preserved; 4.9 locks pre-existing banner behavior
- **Pure functions created**: 1 (`codeGraphUpdatePair`); 1 extraction (`executeUpgrade` method)

### Work Unit Evidence

| Evidence | Value |
|---|---|
| Focused test command | `go test ./internal/cli/ -run 'ForceCommunityTools\|TestCommunityToolInstallStep\|TestCodeGraphUpgradeSyncStep\|TestSyncCodeGraphUpgrade\|TestRenderSyncReport' -count=1` → `ok 0.466s`; `go test ./internal/tui/screens/ -run CommunityTools -count=1` → `ok`; `go test ./internal/tui/ -run TestWelcomeBannerSummaryLineIncludesCodeGraphVersionPair -count=1` → `ok` |
| Mandated verification | `go build ./...` ✅; `go test ./internal/cli/... -run 'ForceCommunityTools\|CodeGraph' -count=1` → `ok 2.810s`; `go test ./internal/tui/... -count=1` → `ok` (tui + screens); `go vet ./internal/cli/... ./internal/tui/...` ✅; `go run ./internal/gofmtcheck` ✅ |
| Regression gate | `go test ./internal/cli/... ./internal/model/... -count=1` → `ok 220.102s` / `ok` |
| Runtime harness | Seam-swapped executor + version-check doubles driving the REAL `RunSync` pipeline (full-path force run + dry-run force); direct `Model.View()` for the banner (tasks.md Unit 3 row: `Model.Update()`/render-level) |
| Per-link verification | Each chain link verified standalone via stash-isolated checkout: 4a `ok 0.022s`, 4b pre-check `ok 0.406s`, tip full suite green |
| Rollback boundary | PR4a: `install.go`, `install_test.go`, `validate.go`, `run.go`, `run_community_tool_test.go`, `model/selection.go`. PR4b: `sync.go`, `sync_test.go` (flag/thread/force-exec hunks). Tip: `tui/model.go`, `tui/model_test.go`, `tui/screens/community_tools*.go`, renderer hunks in `sync.go`/`sync_test.go`. communitytool/update/openspec-spec/design untouched |

### Files Changed

| File | Action | What |
|------|--------|------|
| `internal/cli/sync.go` | Modified | `SyncFlags.ForceCommunityTools` + `forceCommunityToolsSet` + registration; `BuildSyncSelection` threading; step `force` field + gate bypass; `executeUpgrade` extraction; dry-run force `Pending`; forced-report renderer lines |
| `internal/cli/sync_test.go` | Modified | 5 new test functions (parse, selection, step force matrix, full-path force, dry-run force, forced report lines) |
| `internal/cli/install.go` | Modified | `InstallFlags.ForceCommunityTools` + registration + help line |
| `internal/cli/install_test.go` | Modified | Parse + normalize force tests |
| `internal/cli/validate.go` | Modified | `NormalizeInstallFlags` → `Selection.ForceCommunityTools` |
| `internal/cli/run.go` | Modified | Seam rename to `installCommunityToolWithHomeOpts`; step `force` field; stage plan threading |
| `internal/cli/run_community_tool_test.go` | Modified | 4 stubs updated to opts seam; force-pair test; stage-plan threading test |
| `internal/model/selection.go` | Modified | `Selection.ForceCommunityTools` field |
| `internal/tui/model.go` | Modified | `RenderCommunityTools` call site passes `m.UpdateResults` |
| `internal/tui/model_test.go` | Modified | Banner version-pair test (4.9) |
| `internal/tui/screens/community_tools.go` | Modified | `updates` param + version-pair line + `codeGraphUpdatePair` |
| `internal/tui/screens/community_tools_test.go` | Modified | Version-pair table test; existing calls updated |
| `openspec/changes/codegraph-version-aware-upgrade/tasks.md` | Modified | Checked 4.1–4.9 |

### Deviations from Design

1. **PR4 split into 3 links** (see Split Note): forecast ~200 vs actual ~725 (tests); budget enforced proactively instead of post-review. The requested branch name `feat/984-pr4-force-flags-tui` is kept as the chain tip; PR4a is the link that targets PR3b-2.
2. **Forced-reinstall report lines** (not literal in design): the design data flow skips the registry under force, leaving no from/to pair; rendering the versioned line would print `CodeGraph upgraded:  → `. Added performed/pending/rolled-back forced lines + tests, sliced into the tip PR for budget.
3. **Dry-run + force sets `Pending=true`** (extension of Decision 5 honesty): a forced dry-run would otherwise report "nothing actionable" while a real run reinstalls. Check still runs best-effort for from/to when the registry answers.
4. **`forceCommunityToolsSet` recorded but not consumed by `RestorePersistedSelection`**: `state.RestoreSelection` only touches Components/Skills/Preset/SDDMode/StrictTDD, so the one-shot flag survives restore without a guard. Field kept for `...Set` pattern parity (task 4.1 requires the marker).
5. **Task 4.9 closed with test only** (explicitly allowed by the task): Phase 1's registry entry already flows through the generic `UpdateSummaryLine`; the test locks it at the `model.View()` integration point.

### Issues Found

None new. The pre-existing non-idempotence of `codeGraphGuidanceSyncStep` (Batch 2, Issue 1) remains out of scope; the forced reinstall appends the same managed paths as content-compared candidates, so honest reporting is preserved.

### Remaining Tasks

None — 4.1–4.9 complete. Change is ready for sdd-verify.
