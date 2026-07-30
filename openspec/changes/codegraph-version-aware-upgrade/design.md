# Design: CodeGraph Version-Aware Upgrade/Reconcile

## Technical Approach

Gate versions in `internal/cli` via Phase 1's `update.CheckFiltered`; execute in `internal/components/communitytool`, which owns the npm/`codegraph install` commands, snapshot/restore, and target detection — avoiding the import cycle `cli → update/upgrade → cli` (`upgrade/strategy.go` imports `internal/cli`). Sync gains `codeGraphUpgradeSyncStep` before the guidance step; install gains a force bypass; TUI adds only the SHOULD-level version line.

## Architecture Decisions

| # | Decision | Options (tradeoff) | Choice / Rationale |
|---|----------|--------------------|--------------------|
| 1 | Check + execution placement | (a) cli calls exported `upgrade.npmGlobalUpgrade` — **import cycle**. (b) Reimplement in sync.go — duplicates command knowledge. (c) Check in cli, execute in communitytool | **(c).** `cli → internal/update` is acyclic. `codeGraphCommands` already builds the spec'd sequence (`npm\|pnpm install -g @latest` + `codegraph install --target … --yes`) with tested rollback |
| 2 | Step ordering (sync.go:541) | Before vs after guidance | **Before** guidance/pi steps (`sync:community-tool:codegraph-upgrade`). Upgrade rewires first; guidance reconciles the final CLI; overlap is idempotent. Version fetched once per run inside the step (per-run cache); `stagePlan` stays network-free so dry-run works |
| 3 | Force + upgrade + rollback seam | `InstallOpts{ForceReinstall}` + `InstallWithHomeOpts` (`InstallWithHome` delegates zero opts); early-return (tool.go:144) gated by `!opts.ForceReinstall`. New `UpgradeCodeGraphWithHome`: capture version (package-var) + targets → snapshot → forced full-install; on failure restore + reinstall `@<captured>` + rewire | One mechanism for auto-upgrade, force flag, install force. Self-captured pin works even when the registry check was skipped |
| 4 | Failure policy | Fail sync vs rollback-then-warn | Matrix below. Failing a sync that changed other files over an npm hiccup is hostile; unrecoverable corruption must be loud |
| 5 | Dry-run + surfacing | `SyncResult.CodeGraphUpgrade *CodeGraphUpgradeOutcome`; dry-run branch (sync.go:1497) checks, sets `Pending`, executes nothing; report prints pending/upgraded/warning; performed upgrade forces `NoOp=false`; step appends `CodeGraphManagedPaths` as candidates (content-compared) | Dry-Run Reporting without weakening NoOp honesty |
| 6 | Flag threading | `SyncFlags.ForceCommunityTools` + `forceCommunityToolsSet` (`...Set` pattern) → `model.Selection.ForceCommunityTools` → step; same for `InstallFlags` → `communityToolInstallStep.force` → `InstallWithHomeOpts`. CodeGraph-only per spec | Mirrors existing flag patterns; TUI install path untouched |
| 7 | TUI scope | Upgrade screen + banner already render installed→latest from `UpdateResults` (Phase 1 entry) — SHALL needs zero TUI diff. PR4 adds SHOULD: `RenderCommunityTools` reads the pair from `m.UpdateResults` | Minimal diff. Sequence diagrams: N/A — no new TUI flow |

### Decision 4 failure matrix

| Failure | Behavior |
|---|---|
| Registry check fails | Step returns nil; `Warning`; install untouched; sync succeeds (spec: Resilient Sync) |
| `DevBuild`/`VersionUnknown`/`NotInstalled`/`UpToDate`, no force | Skip silently |
| npm/pnpm CLI missing at execution | Abort before any command; warning; sync continues |
| Reinstall/wiring fails; rollback succeeds | `RolledBack=true` + Warning; sync continues |
| Rollback fails | Joined error incl. manual command `npm install -g @colbymchenry/codegraph@<v>`; pipeline snapshot backstop; sync fails loudly |

## Data Flow

```
sync ─► codeGraphUpgradeSyncStep
         force ─► UpgradeCodeGraphWithHome
         else CheckFiltered ─► UpdateAvailable ─► capture → snapshot →
           reinstall @latest → rewire (failure → restore + reinstall @<captured>)
         ▼
   SyncResult.CodeGraphUpgrade ─► RenderSyncReport
```

## File Changes

| File | Action | PR | Description |
|------|--------|----|-------------|
| `internal/components/communitytool/tool.go` | Modify | 3a | `InstallOpts`, `InstallWithHomeOpts`, force bypass, `UpgradeCodeGraphWithHome` |
| `internal/components/communitytool/upgrade_test.go` | Create | 3a | RED: force bypass, rollback-reinstall, rollback-failure join |
| `internal/cli/sync.go` | Modify | 3b | Upgrade step + ordering, outcome field, dry-run check, report |
| `internal/cli/sync_test.go` | Modify | 3b | RED: ordering, registry-down warn, dry-run pending, NoOp |
| `internal/cli/sync.go` + `install.go` flags | Modify | 4 | `--force-community-tools` both commands |
| `internal/cli/run.go`, `internal/model/selection.go` | Modify | 4 | Thread force to install step |
| `internal/tui/model.go`, `tui/screens/community_tools.go` | Modify | 4 | SHOULD version pair from `UpdateResults` |

PR sizing (budget 400): 3a ≈ 250, 3b ≈ 220, 4 ≈ 200 lines. Planned "PR3" splits into 3a/3b; auto-chain absorbs the extra link.

## Interfaces / Contracts

```go
type InstallOpts struct{ ForceReinstall bool }
func InstallWithHomeOpts(id model.CommunityToolID, workspaceDir, homeDir string,
    runner Runner, detector Detector, opts InstallOpts) (Result, error)
func UpgradeCodeGraphWithHome(homeDir, workspaceDir string,
    runner Runner, detector Detector) (Result, error)

type CodeGraphUpgradeOutcome struct {
    Pending, Performed, RolledBack bool
    From, To, Warning              string
}
```

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| communitytool unit | Force bypass; capture; rollback pin+targets; rollback-failure join | Fake Runner/Detector; swap capture var |
| cli unit | Ordering; registry-down warn; dry-run Pending; NoOp suppression | httptest registry (`npmRegistryBaseURL` var) |
| Integration | `go test ./...` | Existing sync harness |

## Threat Matrix

| Boundary | Applicability | Reason |
|---|---|---|
| Documentation-like paths | N/A | No executable-file classification |
| Git repository selection | N/A | No git operations |
| Commit state | N/A | No git operations |
| Push state | N/A | No git operations |
| PR commands | N/A | No PR automation |

Subprocess boundary uses existing seams: argv-slice execution (no shell interpolation), `Runner` + package-var doubles; RED tests assert command construction.

## Migration / Rollout

No migration required. Existing installs gain upgrades on next sync; force flag is additive.

## Open Questions

- [ ] `codeGraphUpstreamVersion = "1.4.1"` pin in `codegraph_contract.go` — confirm it stays informational.
