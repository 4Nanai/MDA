## ADDED Requirements

### Requirement: Repository testing skill
The repository SHALL provide a discoverable Pipeline testing skill that instructs coding agents and contributors to use the project test CLI consistently.

#### Scenario: Skill is selected for a testing request
- **WHEN** a user asks to test, debug, or verify a Pipeline node against a screenshot
- **THEN** the skill description and instructions direct the agent to the node-test workflow

### Requirement: Node-test workflow guidance
The skill MUST define the complete workflow for preparing inputs, invoking the CLI, evaluating results, and reporting evidence.

#### Scenario: Positive fixture workflow
- **WHEN** a node is expected to recognize a screenshot
- **THEN** the skill instructs the contributor to run a `hit` assertion and report the algorithm and matched box

#### Scenario: Negative fixture workflow
- **WHEN** false-positive resistance matters
- **THEN** the skill instructs the contributor to add or run at least one representative `miss` assertion

#### Scenario: Custom Recognition workflow
- **WHEN** a node uses Custom Recognition
- **THEN** the skill explains the required Agent build or executable, connection diagnostics, and how setup failure differs from a miss

### Requirement: Test selection guidance
The skill SHALL explain the boundary between static validation, screenshot-based node recognition testing, future deterministic Pipeline tests, and live controller testing.

#### Scenario: Static screenshot is sufficient
- **WHEN** a change only affects recognition parameters or templates
- **THEN** the skill recommends positive and negative node screenshot tests before live execution

#### Scenario: Static screenshot is insufficient
- **WHEN** correctness depends on an action changing the screen, timing, controller behavior, `next` routing, or task option merging
- **THEN** the skill states that a node screenshot test is insufficient and identifies the additional integration or live testing needed

### Requirement: Evidence and troubleshooting guidance
The skill MUST define what evidence to retain and provide targeted troubleshooting for common failures.

#### Scenario: Report a completed node test
- **WHEN** an agent completes a node test
- **THEN** its report includes the command, fixture path, expectation, actual result, matched box when available, and any saved artifact path

#### Scenario: Diagnose common failures
- **WHEN** a test fails because of dimensions, missing resources, unknown nodes, unavailable Agent registration, OCR mismatch, or template mismatch
- **THEN** the skill provides stage-specific checks rather than recommending arbitrary threshold reductions

### Requirement: Skill and CLI consistency
The skill SHALL use commands and options supported by the implemented CLI and SHALL be verified whenever the CLI interface changes.

#### Scenario: CLI contract changes
- **WHEN** command names, flags, defaults, output fields, or exit statuses change
- **THEN** the implementation change also updates and validates the testing skill examples
