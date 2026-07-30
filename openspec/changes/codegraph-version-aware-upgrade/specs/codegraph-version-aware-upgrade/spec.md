# Delta for CodeGraph Version-Aware Upgrade

## ADDED Requirements

### Requirement: Version Detection and Comparison

The system SHALL compare the installed CodeGraph CLI version to the latest npm registry version when the CodeGraph community tool is selected.

#### Scenario: Installed version is older than latest

- GIVEN CodeGraph `1.4.1` is installed and npm latest is `1.5.0`
- WHEN sync plans the CodeGraph step
- THEN the system detects an upgrade is available

#### Scenario: Installed version matches latest

- GIVEN CodeGraph `1.5.0` is installed and npm latest is `1.5.0`
- WHEN sync plans the CodeGraph step
- THEN the system reports no upgrade is needed

### Requirement: Sync Auto-Upgrade

When an upgrade is available and `--dry-run` is false, the system SHALL reinstall `@colbymchenry/codegraph@latest` globally and rerun `codegraph install --target <target> --location global --yes` for every detected native target.

#### Scenario: Outdated CLI with detected targets

- GIVEN the installed version is older than latest and native targets exist
- AND `--dry-run` is false
- WHEN sync executes the CodeGraph step
- THEN npm installs the latest package globally and `codegraph install` runs for each target

#### Scenario: Outdated CLI with no native targets

- GIVEN the installed version is older than latest and no native targets are detected
- AND `--dry-run` is false
- WHEN sync executes the CodeGraph step
- THEN only the global npm reinstall occurs

### Requirement: Dry-Run Reporting

When `--dry-run` is true and an upgrade is available, the system SHALL report the pending upgrade and SHALL NOT modify the CLI or wiring.

#### Scenario: Dry run reports pending upgrade

- GIVEN the installed version is older than latest and `--dry-run` is true
- WHEN sync executes
- THEN the plan reports the pending CodeGraph upgrade and no install commands run

### Requirement: Resilient Sync on Registry Failure

If the npm registry is unreachable or returns an error, the system SHALL warn, preserve the existing installation, and continue sync without failing.

#### Scenario: npm registry is unreachable

- GIVEN the npm registry returns an error
- WHEN sync plans the CodeGraph upgrade
- THEN the system logs a warning, leaves the installed CLI unchanged, and sync succeeds

### Requirement: Forced Reinstall

The `--force-community-tools` flag on `sync` and `install` SHALL bypass the `CodeGraphReconcileSatisfied` early-return and reinstall CodeGraph. In this change the flag SHALL affect CodeGraph only.

#### Scenario: Force reinstall when already satisfied

- GIVEN CodeGraph is satisfied at the latest version and `--force-community-tools` is true
- WHEN `sync` or `install` runs
- THEN the system reinstalls `@latest` globally and rewires detected targets

#### Scenario: No force flag preserves satisfied install

- GIVEN CodeGraph is satisfied and `--force-community-tools` is false
- WHEN `sync` or `install` runs
- THEN the existing installation remains untouched

### Requirement: Upgrade Rollback

Before upgrading, the system SHALL capture the installed CodeGraph version and detected target list. If the upgrade fails, it SHALL restore CodeGraph-managed files and reinstall the captured version for the captured targets.

#### Scenario: Wiring fails after CLI upgrade

- GIVEN an upgrade is in progress and post-install wiring fails
- WHEN rollback runs
- THEN the captured version is reinstalled, targets are rewired, and managed files are restored

### Requirement: TUI Version Surfacing

The update registry, TUI upgrade screen, and welcome banner SHALL display the installed CodeGraph version and the latest npm version. The Community Tools screen SHOULD display the same version pair.

#### Scenario: Upgrade screen shows version pair

- GIVEN the update registry has a CodeGraph hint with installed and latest versions
- WHEN the Upgrade screen renders
- THEN both versions are visible

#### Scenario: Welcome banner shows available upgrade

- GIVEN the welcome banner renders after a sync check
- AND a CodeGraph upgrade is available
- THEN the banner shows the installed and latest versions
