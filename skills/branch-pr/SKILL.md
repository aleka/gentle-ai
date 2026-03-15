---
name: gentle-ai-branch-pr
description: >
  PR creation workflow for Gentle AI following the issue-first enforcement system.
  Trigger: When creating a pull request, opening a PR, or preparing changes for review.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "2.0"
---

# Gentle AI — Branch & PR Skill

## When to Use

Load this skill whenever you need to:
- Create a branch for a new fix or feature
- Open a pull request on [Gentleman-Programming/gentle-ai](https://github.com/Gentleman-Programming/gentle-ai)
- Prepare changes for review

## Critical Rules

1. **Every PR MUST link an approved issue** — `Closes/Fixes/Resolves #<N>` in the PR body, and that issue MUST have `status:approved`. PRs without this are **automatically rejected** by CI.
2. **Exactly one `type:*` label** — apply exactly ONE type label to the PR. CI will reject PRs with zero or multiple type labels.
3. **5 automated checks must pass** — see the Automated Checks table below.
4. **No `Co-Authored-By` trailers** — never add AI attribution to commits.
5. **No force-push to main/master** — protected branch.

## Workflow

```
1. Confirm the issue has status:approved
   gh issue view <N> --repo Gentleman-Programming/gentle-ai

2. Create a branch from main using the naming convention below

3. Implement changes following specs and design

4. Run tests locally (unit + E2E)

5. Commit using Conventional Commits format

6. Open a PR referencing the issue
   → Add exactly ONE type:* label
   → Fill in the PR body using the template

7. All 5 automated checks must pass before merge
```

---

## Branch Naming

| Prefix | Use For | Example |
|--------|---------|---------|
| `fix/` | Bug fixes | `fix/agent-claude-detection-linux` |
| `feat/` | New features | `feat/tui-keyboard-help-overlay` |
| `docs/` | Documentation only | `docs/update-contributing-guide` |
| `refactor/` | Code refactoring | `refactor/pipeline-step-interface` |
| `chore/` | Build, CI, tooling | `chore/add-unit-tests-ci-job` |

Branch names must be:
- Lowercase, hyphen-separated
- Descriptive but concise
- Prefixed to match the PR type label

```bash
# Create and switch to a new branch
git checkout -b fix/agent-claude-detection-linux
```

---

## PR Body Format

The PR body must follow the template at `.github/PULL_REQUEST_TEMPLATE.md`. All sections are required unless marked optional.

```markdown
## 🔗 Linked Issue

Closes #<N>

## 🏷️ PR Type

- [ ] `type:bug` — Bug fix (non-breaking change that fixes an issue)
- [ ] `type:feature` — New feature (non-breaking change that adds functionality)
- [ ] `type:docs` — Documentation only
- [ ] `type:refactor` — Code refactoring (no functional changes)
- [ ] `type:chore` — Build, CI, or tooling changes
- [ ] `type:breaking-change` — Breaking change

## 📝 Summary

<!-- Clear description of what this PR does and why. -->

## 📂 Changes

| File / Area | What Changed |
|-------------|-------------|
| `path/to/file` | Brief description |

## 🧪 Test Plan

**Unit Tests**
\`\`\`bash
go test ./...
\`\`\`

**E2E Tests** (Docker required)
\`\`\`bash
cd e2e && ./docker-test.sh
\`\`\`

- [ ] Unit tests pass (`go test ./...`)
- [ ] E2E tests pass (`cd e2e && ./docker-test.sh`)
- [ ] Manually tested locally

## ✅ Contributor Checklist

- [ ] PR is linked to an issue with `status:approved`
- [ ] I have added the appropriate `type:*` label to this PR
- [ ] Unit tests pass (`go test ./...`)
- [ ] E2E tests pass (`cd e2e && ./docker-test.sh`)
- [ ] I have updated documentation if necessary
- [ ] My commits follow Conventional Commits format
- [ ] My commits do not include `Co-Authored-By` trailers
```

---

## Automated Checks

All 5 checks run on every PR and **all must pass** before merge:

| Check | What It Verifies | How to Fix |
|-------|-----------------|------------|
| **Check Issue Reference** | PR body contains `Closes/Fixes/Resolves #N` | Add `Closes #<N>` to the PR body |
| **Check Issue Has `status:approved`** | Linked issue has been approved by a maintainer | Wait for maintainer to add `status:approved` to the issue |
| **Check PR Has `type:*` Label** | Exactly one `type:*` label is applied to the PR | Ask a maintainer to add the correct label; remove extras |
| **Unit Tests** | `go test ./...` passes | Fix failing tests before pushing |
| **E2E Tests** | `cd e2e && ./docker-test.sh` passes | Fix failing E2E scenarios before pushing |

---

## Conventional Commits

All commits on this project MUST follow [Conventional Commits](https://www.conventionalcommits.org/).

### Format

```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

### Type → Label Mapping

| Commit Type | PR Label | Description |
|-------------|----------|-------------|
| `fix` | `type:bug` | Bug fix |
| `feat` | `type:feature` | New feature or enhancement |
| `docs` | `type:docs` | Documentation changes only |
| `refactor` | `type:refactor` | Refactoring, no functional changes |
| `chore` | `type:chore` | Build, CI, tooling |
| `feat!` / `fix!` | `type:breaking-change` | Breaking change (add `!` and `BREAKING CHANGE:` footer) |
| `test` | `type:chore` | Adding or updating tests |
| `perf` | `type:feature` or `type:refactor` | Performance improvement |

### Valid Scopes

`tui`, `cli`, `installer`, `catalog`, `system`, `agent`, `e2e`, `ci`, `docs`

### Examples

```bash
feat(tui): add progress bar to installation steps
fix(agent): correct Claude Code detection on macOS
docs: update contributing guide with E2E test instructions
chore(ci): add unit tests job to CI workflow
test(installer): add coverage for catalog step execution
refactor(pipeline): extract step runner interface
```

### Breaking Change

```bash
feat(cli)!: rename --config flag to --config-file

BREAKING CHANGE: the --config flag has been renamed to --config-file.
Update your scripts and aliases accordingly.
```

---

## Commands

### Setup

```bash
# Confirm issue is approved before starting
gh issue view <N> --repo Gentleman-Programming/gentle-ai

# Create branch
git checkout main && git pull
git checkout -b fix/<short-description>
```

### Testing Locally

```bash
# Unit tests
go test ./...

# Unit tests — specific package
go test ./internal/tui/...

# Unit tests — verbose
go test -v ./...

# E2E tests (Docker must be running)
cd e2e && ./docker-test.sh
```

### Open a PR

```bash
gh pr create \
  --repo Gentleman-Programming/gentle-ai \
  --title "fix(agent): correct Claude Code detection on Linux" \
  --body "$(cat <<'EOF'
## 🔗 Linked Issue

Closes #42

## 🏷️ PR Type

- [x] \`type:bug\` — Bug fix (non-breaking change that fixes an issue)

## 📝 Summary

Fixes Claude Code binary detection failing on Linux when HOME is not set.

## 📂 Changes

| File / Area | What Changed |
|-------------|-------------|
| \`internal/agents/claude.go\` | Added HOME env var fallback |

## 🧪 Test Plan

- [x] Unit tests pass (\`go test ./...\`)
- [x] E2E tests pass (\`cd e2e && ./docker-test.sh\`)
- [x] Manually tested locally

## ✅ Contributor Checklist

- [x] PR is linked to an issue with \`status:approved\`
- [x] I have added the appropriate \`type:*\` label to this PR
- [x] Unit tests pass (\`go test ./...\`)
- [x] E2E tests pass (\`cd e2e && ./docker-test.sh\`)
- [x] I have updated documentation if necessary
- [x] My commits follow Conventional Commits format
- [x] My commits do not include \`Co-Authored-By\` trailers
EOF
)"
```

### Check PR Status

```bash
gh pr checks --repo Gentleman-Programming/gentle-ai <PR-number>
gh pr view --repo Gentleman-Programming/gentle-ai <PR-number>
```

### Add a Label

```bash
gh pr edit <PR-number> --repo Gentleman-Programming/gentle-ai --add-label "type:bug"
```
