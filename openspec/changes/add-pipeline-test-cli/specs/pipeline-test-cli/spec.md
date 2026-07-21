## ADDED Requirements

### Requirement: Node recognition command
The system SHALL provide a `pipeline-test node` command that accepts a Pipeline node name and a local screenshot and evaluates only that node's recognition configuration.

#### Scenario: Test a matching node
- **WHEN** a contributor invokes the node command with an existing node, a readable screenshot, and an expected result of `hit`
- **THEN** the command evaluates the node recognition and exits successfully when the recognition hits

#### Scenario: Test an intentional non-match
- **WHEN** a contributor invokes the node command with an expected result of `miss`
- **THEN** the command exits successfully only when the recognition does not hit

#### Scenario: Do not execute node behavior
- **WHEN** the target node contains an action, delays, waits, `next`, or `on_error` configuration
- **THEN** the command evaluates its recognition without executing the action, waiting, or traversing another node

### Requirement: Resource-aware recognition
The node command MUST load the complete selected MaaFramework resource bundle before evaluating a node so that templates, OCR models, referenced nodes, resource overlays, and runtime node data use production semantics.

#### Scenario: Resolve a referenced recognition node
- **WHEN** the target uses `And`, `Or`, or another recognition that references a node outside the target's source JSON file
- **THEN** the command resolves the reference from the fully loaded resource bundle

#### Scenario: Select a non-default resource
- **WHEN** the contributor supplies an explicit resource path
- **THEN** the command loads that resource instead of the repository default

#### Scenario: Reject an unknown node
- **WHEN** the requested node does not exist in the loaded resource
- **THEN** the command reports an input error without producing a passing assertion

### Requirement: Custom Recognition support
The node command SHALL support nodes backed by the MDA Go Agent's registered Custom Recognition implementations.

#### Scenario: Evaluate an Agent-backed node
- **WHEN** the target node uses a registered Custom Recognition and the configured Agent is available
- **THEN** the command evaluates the Custom Recognition against the supplied screenshot and reports its result

#### Scenario: Agent is unavailable
- **WHEN** the target requires a Custom Recognition but the Agent executable cannot be started or connected
- **THEN** the command reports a setup error distinct from a recognition miss

### Requirement: Screenshot validation
The node command MUST decode local PNG and JPEG screenshots without modifying their pixels and MUST reject screenshots that do not meet MDA's 1280 by 720 Pipeline coordinate baseline unless an explicit future compatibility mode is introduced.

#### Scenario: Valid screenshot
- **WHEN** the supplied screenshot is a readable 1280 by 720 PNG or JPEG image
- **THEN** the command passes the decoded image to MaaFramework without resizing it

#### Scenario: Invalid screenshot dimensions
- **WHEN** the supplied screenshot dimensions differ from 1280 by 720
- **THEN** the command fails with a diagnostic containing the actual and required dimensions

#### Scenario: Unreadable screenshot
- **WHEN** the image path is absent or cannot be decoded
- **THEN** the command reports an input error and does not run recognition

### Requirement: Deterministic assertions and exit status
The command SHALL distinguish a passing expectation, a failed expectation, and a setup or execution error through stable process exit statuses.

#### Scenario: Assertion passes
- **WHEN** actual hit status equals the requested expectation
- **THEN** the process exits with status `0`

#### Scenario: Assertion fails
- **WHEN** actual hit status differs from the requested expectation
- **THEN** the process exits with status `1` and still reports the actual recognition result

#### Scenario: Test cannot execute
- **WHEN** arguments, resource loading, Agent startup, image loading, or MaaFramework execution fails
- **THEN** the process exits with status `2` and identifies the failing stage

### Requirement: Human and machine-readable results
The node command SHALL provide concise human-readable output by default and a versioned JSON output mode suitable for scripts and CI.

#### Scenario: Human-readable result
- **WHEN** no output format is specified
- **THEN** the result includes the node, image, expected and actual hit status, algorithm, matched box when present, duration, and assertion outcome

#### Scenario: JSON result
- **WHEN** JSON output is requested
- **THEN** stdout contains one valid JSON document with a schema version and the same core result fields

#### Scenario: Combined recognition details
- **WHEN** the recognition returns combined child results
- **THEN** JSON output preserves enough child detail to identify which recognition component hit or missed

#### Scenario: Diagnostic artifact
- **WHEN** MaaFramework provides recognition draw images and the contributor requests an artifact directory
- **THEN** the command writes the artifacts there and reports their paths without changing the assertion result

### Requirement: Extensible command architecture
The CLI implementation SHALL separate runtime setup, recognition execution, result modeling, rendering, and command parsing so future `pipeline` and `task` commands can reuse them without changing the node command contract.

#### Scenario: Node command remains isolated
- **WHEN** the MVP is delivered
- **THEN** only the `node` execution workflow is required and no incomplete `pipeline` or `task` command is presented as functional

#### Scenario: Future command reuse
- **WHEN** a later change adds Pipeline fixture playback or Project Interface task resolution
- **THEN** it can reuse resource, Agent, reporting, and exit-status components defined by this CLI
