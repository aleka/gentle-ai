# Archive Report: codegraph-version-aware-upgrade

## Change Identity

- **Change**: `codegraph-version-aware-upgrade`
- **Archived**: 2026-07-30
- **Archive path**: `openspec/changes/archive/2026-07-30-codegraph-version-aware-upgrade/`
- **Main spec**: `openspec/specs/codegraph-version-aware-upgrade/spec.md` (created — full spec copied from delta)

## Final State (authoritative)

Per the orchestrator's final-state facts (highest-ranked source after native review authority):

- **Implementation**: COMPLETE — all 26 tasks across Phases 1–4 are checked in `tasks.md`.
- **Verification**: PASS WITH WARNINGS — 7/7 requirements, 11/11 spec scenarios PROVEN, CRITICAL 0, WARNING 2.
- **Runtime ledger**: 4/4 attempts passed, change marked complete.

### Branch Chain (local, not pushed)

| Branch | Commit | Description |
|--------|--------|-------------|
| `feat/984-codegraph-version-aware` (tracker) | — | Tracker branch |
| `feat/984-pr1-codegraph-detection` | `659f3728` | npm registry version detection |
| `feat/984-pr2-npm-global-strategy` | `1cbb294b` | npm-global upgrade strategy |
| `feat/984-pr3a-communitytool-upgrade` | `f731451f` | community-tool upgrade/rollback core |
| `feat/984-pr3b1-sync-upgrade-step` | `40dde85e` | sync auto-upgrade step |
| `feat/984-pr3b2-sync-report-tests` | `bb6ff8fe` | sync report + tests |
| `feat/984-pr4a-install-force-flags` | `8079e43d` | install force flag threading |
| `feat/984-pr4b-sync-force-gate` | `71800ca0` | sync force gate bypass |
| `feat/984-pr4-force-flags-tui` (tip) | `bb177bbb` | TUI version surfacing + forced reinstall lines |

PR3b was split into pr3b1 (+479, maintainer-accepted `size:exception`) + pr3b2 (+247). PR4 was split into 3a/3b/tip to respect the 400-line review budget.

## Verification Summary

Source: `verify-report.md` (intermediate snapshot at verification time 2026-07-30). Final-state facts from the orchestrator confirm no changes after verification.

| Metric | Value |
|--------|-------|
| Requirements | 7/7 |
| Scenarios | 11/11 |
| CRITICAL | 0 |
| WARNING | 2 |
| SUGGESTION | 1 |
| Verdict | PASS WITH WARNINGS |
| Build | `go build ./...` — OK |
| Tests | `go test ./internal/update/... ./internal/components/communitytool/... ./internal/cli/... ./internal/tui/... -count=1` — OK |
| Vet | `go vet ./...` — OK |
| Format | `go run ./internal/gofmtcheck` — OK |

### Warnings (accepted as final state — not fixed later)

1. **Pre-existing flaky test**: `TestConcurrentFinalizeElectsExactlyOneWriter` in `internal/cli` — unrelated to this change (review-integration concurrency test). Failed once during verification, passed on rerun.
2. **Low file-level coverage on large TUI files**: `internal/tui/model.go` (61.2%) and `internal/tui/screens/community_tools.go` (62.7%) below 80% threshold, but modified code paths are covered. Bulk is pre-existing TUI logic outside this change's scope.

### Suggestion

1. Stabilize `TestConcurrentFinalizeElectsExactlyOneWriter` for deterministic full cli suite.

## Spec Compliance

All 7 requirements and 11 scenarios proven by passing tests:

| Requirement | Scenarios | Status |
|-------------|-----------|--------|
| Version Detection and Comparison | 2/2 | PROVEN |
| Sync Auto-Upgrade | 2/2 | PROVEN |
| Dry-Run Reporting | 1/1 | PROVEN |
| Resilient Sync on Registry Failure | 1/1 | PROVEN |
| Forced Reinstall | 2/2 | PROVEN |
| Upgrade Rollback | 2/2 | PROVEN |
| TUI Version Surfacing | 2/2 | PROVEN (3 scenarios including Community Tools screen) |

## Design Decisions Implemented

All 7 architecture decisions from `design.md` were implemented as specified:

1. ✅ Check in cli, execute in communitytool (avoids import cycle)
2. ✅ Upgrade step before guidance in `stagePlan`
3. ✅ `InstallOpts{ForceReinstall}` seam with `InstallWithHomeOpts`
4. ✅ Decision-4 failure matrix implemented in step and executor
5. ✅ Dry-run + `CodeGraphUpgradeOutcome` with NoOp honesty
6. ✅ Flag threading via `SyncFlags`/`InstallFlags` → `model.Selection`
7. ✅ TUI scope: `RenderCommunityTools` version pair + banner pair

## Known Follow-ups (NOT in this change — future work)

1. **opencode.json rewrite on every sync**: Pre-existing non-idempotency in `codeGraphGuidanceSyncStep` — file idempotency issue.
2. **"rollback failed" string discriminator**: Currently discriminates on string marker in error; future refactor to sentinel error.
3. **pnpm-only rollback detection**: Current rollback assumes npm; pnpm-only environments need detection.

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| `codegraph-version-aware-upgrade` | Created | Full spec copied (7 ADDED requirements, 13 scenarios) — no prior main spec existed |

## Archive Contents

- `proposal.md` ✅
- `specs/codegraph-version-aware-upgrade/spec.md` ✅
- `design.md` ✅
- `tasks.md` ✅ (26/26 tasks complete)
- `apply-progress.md` ✅ (3 batches)
- `verify-report.md` ✅
- `archive-report.md` ✅ (this file)

## Source of Truth Updated

- `openspec/specs/codegraph-version-aware-upgrade/spec.md` — created with the full spec

## SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived. Ready for the next change.
