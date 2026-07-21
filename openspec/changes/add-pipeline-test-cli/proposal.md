## Why

MDA currently validates Pipeline files structurally but cannot execute a node against a saved screenshot, so recognition changes require manual UI testing and are difficult to reproduce in review or CI. We need a repeatable testing entry point that starts with node recognition tests and can later grow into deterministic Pipeline and Project Interface task tests.

## What Changes

- Introduce an extensible `pipeline-test` CLI with a `node` command as the MVP.
- Allow the node command to load the complete MDA resource bundle, run one named node's recognition against a local screenshot, and support built-in, composed, and Agent-provided Custom Recognition.
- Provide explicit hit/miss assertions, structured recognition details, deterministic process exit codes, and optional machine-readable output for scripts and CI.
- Keep node tests recognition-only: the command must not execute the node action or follow `next` branches.
- Establish fixture and result conventions that leave room for future `pipeline` and `task` commands without including those commands in the MVP implementation scope.
- Add a repository skill describing how agents and contributors prepare screenshots, select assertions, run node tests, interpret artifacts, and decide when live testing is still required.
- Add automated tests and developer documentation for the CLI and skill.

## Capabilities

### New Capabilities

- `pipeline-test-cli`: Defines the node-test CLI contract, resource and Agent integration, assertions, output, exit behavior, and extension boundary for future Pipeline and task testing.
- `pipeline-testing-skill`: Defines the reusable contributor and coding-agent workflow for selecting, executing, and diagnosing MDA Pipeline tests.

### Modified Capabilities

None.

## Impact

- Adds a new Go command and test-support packages under `agent\go-service`.
- Uses the existing MaaFramework runtime, resource bundle, Go binding, and Agent server; no second recognition implementation is introduced.
- Adds repository-owned node fixtures and a testing skill under the existing project skill structure.
- Extends CI with deterministic node-test coverage after the MVP fixtures are available.
- Does not change production Pipeline behavior, Project Interface task definitions, or end-user UI behavior.
