# Project: Sire - The MCP Orchestration Engine

## 1. Context

### Problem Statement

Modern development involves a multitude of tools and services, each with its own API. Automating workflows across these tools is complex, often requiring developers to write language-specific plugins for a central, monolithic engine. This creates tight coupling and slows down development. Furthermore, these distributed systems are inherently unreliable; network failures and transient service outages can cause workflows to fail, leading to data loss and incomplete processes. There is a need for a lightweight, high-performance orchestrator that can connect and execute tools from any source in a decoupled, language-agnostic, and **reliable** manner.

### Objectives and Non-Goals

**Objectives:**

*   **Pivot Sire:** Refactor the Sire platform to be a pure orchestration engine for any tool that exposes a Model-Context-Protocol (MCP) endpoint.
*   **Unify Tool Execution:** Treat all executable units, whether local Go functions or remote services, as MCP tools. The engine should not know the difference.
*   **Enable Federated Workflows:** Allow users to build workflows that seamlessly orchestrate tools across multiple, independent MCP servers.
*   **Guarantee Durable Execution:** Provide at-least-once execution guarantees for all workflow steps. The system must be able to survive orchestrator restarts and transient tool failures, resuming workflows automatically.
*   **Developer-First Experience:** Define workflows as code (YAML) and maintain a simple, powerful CLI experience.

**Non-Goals:**

*   **Building a large, built-in library of nodes:** The new focus is on enabling the community to provide tools via MCP, not on building an extensive internal library.
*   **A visual UI or workflow editor:** The primary interface remains the CLI and the workflow-as-code definition files.
*   **Complex tool discovery mechanisms (in this phase):** The initial implementation will rely on the user knowing the endpoints and tool names of the services they wish to orchestrate.
*   **High-availability orchestrator cluster:** The initial focus is on durability for a single orchestrator instance. Clustering for high availability is out of scope *for the initial launch*, but is planned for a future version (see Epic E11).

### Constraints and Assumptions

**Constraints:**

*   The core orchestrator must remain in Go for performance.
*   *   The orchestrator requires a persistence layer to store execution state. For simplicity, the initial implementation will use an embedded database (`bbolt`, a actively maintained fork of BoltDB).
*   All communication with tools must happen over the JSON-RPC 2.0-based MCP specification.

**Assumptions:**

*   The MCP standard is a suitable abstraction for a wide variety of tools and services.
*   A "bring your own tools" model is more flexible and powerful than a "batteries-included" one.
*   For most workflows, the performance cost of persisting state after each step is an acceptable trade-off for reliability.

### Success Metrics

*   **Federation:** A user can successfully define and execute a workflow that orchestrates tools from at least two different remote MCP servers in a single run.
*   **Decoupling:** The core engine has zero direct dependencies on any specific tool's implementation (e.g., `file.write`). Its only dependency is on a generic MCP client interface.
*   **Durability:** If the Sire orchestrator is stopped mid-workflow and restarted, it automatically resumes the workflow from the last successfully completed step.
*   **Resilience:** If a remote MCP tool fails due to a transient network error, the orchestrator will automatically retry the step according to its configured policy and eventually complete the workflow.

## 2. Scope and Deliverables

### In Scope

*   A complete refactor of the core engine to be an MCP-native orchestrator.
*   An in-process MCP server that wraps the existing Go-based nodes (`file`, `http`, etc.) to serve them as standard tools.
*   A robust, production-ready `mcp.execute` node (or equivalent) for calling remote MCP servers.
*   A persistence layer for storing and retrieving workflow execution state.
*   Engine logic for resuming workflows and retrying failed steps.
*   A complete rewrite of all user-facing documentation to reflect the new vision and features.

### Out of Scope

*   A public registry or marketplace for discovering third-party MCP tool servers.
*   Advanced authentication or authorization mechanisms for connecting to remote MCP servers (e.g., OAuth2). These are planned for a future version (see Epic E13).
*   A tool for automatically generating an MCP server from an OpenAPI or other API specification.
*   A separate, standalone database server. The persistence layer will be embedded within the Sire binary.

### Deliverables

| ID    | Description                                      | Owner | Acceptance Criteria                                                                                             |
| :---- | :----------------------------------------------- | :---- | :-------------------------------------------------------------------------------------------------------------- |
| D7    | Refactored MCP-First Orchestration Engine        | TBD   | The engine can execute a workflow by dispatching all steps to MCP tool servers (in-process or remote).          |
| D8    | In-Process Tool Server                           | TBD   | The original built-in nodes (`file`, `http`) are served on an in-process MCP server, accessible to the engine.   |
| D9    | Remote Tool Execution                            | TBD   | A workflow step can successfully call a tool on an external, network-accessible MCP server.                     |
| D10   | Rewritten Documentation and Examples             | TBD   | The `README.md` and supporting docs clearly explain the new MCP-first vision and examples.              |
| D11   | Durable Workflow Execution                       | TBD   | Workflows can survive orchestrator restarts and transient tool failures, with state persisted to a local DB.    |

## 3. Checkable Work Breakdown

### Archived

*   **E1: Core Engine:** Initial implementation of a Go-based workflow engine.
*   **E2: Command-Line Interface (CLI):** Initial implementation of the `sire` CLI.
*   **E3: Built-in Nodes:** Creation of the first set of native Go nodes.
*   **E4: MCP Server:** Proof-of-concept MCP server for workflow generation.

### E5: Architectural Pivot: MCP-First Core

*   [x] T5.1 Redesign core data structures  Owner: Gemini  Est: 3h Completion Date: 2025-08-24
    *   [x] S5.1.1 Redefine `Workflow` struct. A `Node` is now a `Step`, which contains a `Tool` identifier and `Params`. Completion Date: 2025-08-24
    *   [x] S5.1.2 Define a `Tool` URI format (e.g., `sire:local/file.write`, `mcp:http://host/rpc#math.add`) to uniquely identify any tool. Completion Date: 2025-08-24
    *   [x] S5.1.3 Add unit tests for new data structures. Completion Date: 2025-08-24
*   [x] T5.2 Refactor the `core.Engine` to be a tool dispatcher  Owner: Gemini  Est: 5h Completion Date: 2025-08-24
    *   [x] S5.2.1 Remove all direct node instantiation logic from the engine. Completion Date: 2025-08-24
    *   [x] S5.2.2 Implement a dispatcher that routes tool execution based on the scheme in the Tool URI (e.g., `sire:`, `mcp:`). Completion Date: 2025-08-24
    *   [x] S5.2.3 The engine's `Execute` method now orchestrates calls via the dispatcher and manages the flow of data between steps. Completion Date: 2025-08-24
    *   [x] S5.2.4 Add unit and integration tests for the new engine and dispatcher logic. Completion Date: 2025-08-24
*   [x] T5.3 Run linter and formatter  Owner: Gemini  Est: 30m Completion Date: 2025-08-24

### E6: Local Tools as an In-Process MCP Server

*   [x] T6.1 Implement an in-process MCP server and client  Owner: Gemini  Est: 4h Completion Date: 2025-08-24
    *   [x] S6.1.1 Create a singleton `InProcessServer` that can host MCP services without a network listener. Completion Date: 2025-08-24
    *   [x] S6.1.2 Create a `sire:local` dispatcher that acts as a client to this in-process server. Completion Date: 2025-08-24
    *   [x] S6.1.3 Add unit tests for the in-process client-server communication. Completion Date: 2025-08-24
*   [x] T6.2 Wrap existing nodes as in-process MCP tools  Owner: Gemini  Est: 4h Completion Date: 2025-08-24
    *   [x] S6.2.1 Refactor `internal/nodes/file` to register its functions as services on the `InProcessServer`. Completion Date: 2025-08-24
    *   [x] S6.2.2 Refactor `internal/nodes/http` to register its functions as services. Completion Date: 2025-08-24
    *   [x] S6.2.3 Refactor `internal/nodes/transform` to register its functions as services. Completion Date: 2025-08-24
    *   [x] S6.2.4 Ensure all original tests for the nodes pass through the new in-process MCP layer. Completion Date: 2025-08-24
*   [x] T6.3 Run linter and formatter  Owner: Gemini  Est: 30m Completion Date: 2025-08-24

### E7: Remote Tool Execution

*   [x] T7.1 Implement the `mcp:` dispatcher  Owner: Gemini  Est: 4h Completion Date: 2025-08-24
    *   [x] S7.1.1 Create the `mcp:` dispatcher which acts as a remote MCP client. Completion Date: 2025-08-24
    *   [x] S7.1.2 The dispatcher should parse the `mcp:` Tool URI to get the server URL and tool name. Completion Date: 2025-08-24
    *   [x] S7.1.3 Add robust error handling for network failures and remote server errors. Completion Date: 2025-08-24
    *   [x] S7.1.4 Implement a configurable timeout for all remote calls. Completion Date: 2025-08-24
*   [x] T7.2 Add integration tests  Owner: Gemini  Est: 3h Completion Date: 2025-08-24
    *   [x] S7.2.1 Create a mock remote MCP server in the test suite. Completion Date: 2025-08-24
    *   [x] S7.2.2 Write integration tests that use the `core.Engine` to execute a workflow that calls the mock remote server. Completion Date: 2025-08-24
*   [x] T7.3 Run linter and formatter  Owner: Gemini  Est: 30m Completion Date: 2025-08-24

### E8: Documentation and Examples

*   [x] T8.1 Rewrite the project `README.md`  Owner: Gemini  Est: 3h Completion Date: 2025-08-24
*   [x] T8.2 Create new example workflows  Owner: Gemini  Est: 4h Completion Date: 2025-08-24
*   [x] T8.3 Update `docs/design.md`  Owner: Gemini  Est: 2h Completion Date: 2025-08-24

### E9: Durable Execution and State Persistence

                *   [x] S9.1.1 Research and select a Go-native embedded database (e.g., `bbolt`, BadgerDB). Decision criteria: simplicity, transactional support, performance. Tentatively select `bbolt`. Completion Date: 2025-08-24
    *   [x] S9.1.2 Implement a `storage` service that abstracts all database operations (e.g., `SaveExecution`, `LoadExecution`, `ListPending`). Completion Date: 2025-08-24
    *   [x] S9.1.3 Add comprehensive unit tests for the storage service, mocking the database interface. Completion Date: 2025-08-24
*   [x] T9.2 Refactor the engine for stateful execution  Owner: Gemini  Est: 6h Completion Date: 2025-08-24
    *   Dependencies: T5.2
    *   [x] S9.2.1 Modify the `sire workflow run` CLI command to create a new execution record in the DB before starting. Completion Date: 2025-08-24
    *   [x] S9.2.2 The `core.Engine` must load the execution state from storage at the beginning of a run. Completion Date: 2025-08-24
    *   [x] S9.2.3 After each step completes, the engine must atomically save the full execution state (including the step's output) before dispatching the next step. Completion Date: 2025-08-24
    *   [x] S9.2.4 Add integration tests to verify that stopping and restarting the orchestrator resumes an in-flight workflow. Completion Date: 2025-08-24
*   [x] T9.3 Implement retry and resumption logic  Owner: Gemini  Est: 5h Completion Date: 2025-08-24
    *   [x] S9.3.1 When a tool call fails with a transient error, the engine marks the step as `retrying` in the database. Completion Date: 2025-08-24
    *   [x] S9.3.2 Implement a background worker process that periodically scans the database for pending or retrying executions. Completion Date: 2025-08-24
    *   [x] S9.3.3 The background worker re-queues these executions for the engine to process. Completion Date: 2025-08-24
    *   [x] S9.3.4 Implement a configurable exponential backoff policy for retries within the workflow step definition. Completion Date: 2025-08-24
    *   [x] S9.3.5 Add unit tests for the retry logic. Completion Date: 2025-08-24
*   [x] T9.4 Update CLI and documentation  Owner: Gemini  Est: 2h Completion Date: 2025-08-24
    *   [x] S9.4.1 Add new CLI commands: `sire execution list` and `sire execution status <id>`. Completion Date: 2025-08-24
    *   [x] S9.4.2 Update documentation to explain the durability guarantees and how to configure retry policies. Completion Date: 2025-08-24
*   [x] T9.5 Run linter and formatter  Owner: Gemini  Est: 30m Completion Date: 2025-08-24


### E10: Performance and Scalability Enhancements

*   [ ] T10.1: Implement Concurrent and Parallel Execution  Owner: TBD  Est: 8h
    *   Dependencies: T5.2, T9.2
    *   Acceptance Criteria: The engine can execute multiple independent branches of a workflow concurrently, significantly reducing overall workflow execution time for parallelizable tasks.
    *   Risk Notes: Potential for race conditions if state management is not handled carefully.
    *   [x] S10.1.1: Identify independent workflow branches using a dependency graph analysis. Completion Date: 2025-08-25
    *   [ ] S10.1.2: Modify the engine's dispatcher to launch concurrent goroutines for independent steps.
    *   [ ] S10.1.3: Ensure concurrent execution respects resource limits and avoids deadlocks.
    *   [ ] S10.1.4: Add unit and integration tests to verify correct concurrent execution and state updates.
*   [ ] **T10.2: Implement Granular State Persistence & Large Data Handling:**
    *   [ ] S10.2.1: Refactor state updates to be atomic and granular, rather than saving the entire execution object.
    *   [ ] S10.2.2: Design and implement an `ArtifactStore` for handling large data blobs outside the primary database.
*   [ ] **T10.3: Implement Dispatcher Connection Caching:** Optimize the `mcp:` dispatcher to reuse network connections.

### E11: High-Availability (HA) Agent

*   [ ] **T11.1: Implement Leader Election:** Allow multiple agent instances to run, with one elected as the leader.
*   [ ] **T11.2: Implement Distributed Locking:** Ensure executions are locked by the leader agent to prevent double-processing.

### E12: Advanced Developer Experience

*   [ ] **T12.1: Implement Interactive CLI Debugger:** Add a `--debug` mode to the CLI for stepping through workflows.

*   [ ] **T12.2: Add Mocking and Dry-Run Capabilities:** Implement `--dry-run` and `--mock-file` flags.

### E13: Comprehensive Security Model

*   [ ] **T13.1: Implement Secure Credential Store:** Integrate a system for securely managing and injecting secrets into workflows.
*   [ ] **T13.2: Implement Workflow RBAC:** Create a Role-Based Access Control system for managing permissions on workflows.
*   [ ] **T13.3: Implement Input/Output Schema Validation:** Allow steps to define and enforce data schemas.

### E14: Production-Grade Observability

*   [ ] **T14.1: Implement Structured Logging:** Mandate structured (JSON) logging across all components.
*   [ ] **T14.2: Expose Prometheus Metrics:** Instrument the engine and agent with key performance metrics.
*   [ ] **T14.3: Integrate Distributed Tracing:** Propagate OpenTelemetry contexts through the dispatcher for end-to-end tracing.

## 4. Timeline and Milestones

| ID | Milestone               | Dependencies | Exit Criteria                                                              |
| :--- | :---------------------- | :----------- | :------------------------------------------------------------------------- |
| M5 | Core Engine Refactor    | E5           | The engine is fully refactored. It no longer has direct knowledge of nodes, only of tool dispatchers. |
| M6 | Unified Tool Execution  | E6, E7       | A single workflow can successfully execute steps using both the `sire:local` and `mcp:` dispatchers. |
| M7 | Durable Execution       | E9, M6       | Workflows are persisted and can survive orchestrator restarts and transient tool failures. |
| M8 | Public Re-Launch        | E8, M7       | The project is ready for a public announcement with its new vision, complete with updated documentation and examples. |

## 5. Operating Procedure

*   **Definition of Done:** A task is done when it is implemented, tested with sufficient coverage, documented, and merged into the main branch.
*   **Review and QA:** All code must be reviewed by at least one other person before being merged. All new features must have corresponding unit and integration tests.
*   **Testing:** Always add tests when adding new implementation code. Aim for 100% test coverage for new code.
*   **Linting and Formatting:** Always run `go fmt` and `golangci-lint` after code changes and before committing.
*   **Commits:** Never commit files from different logical components in the same commit. Make many small, logical commits.

## 6. Progress Log

*   **2025-08-25 (Change Summary):** Updated `docs/design.md` to reflect the decision to use `bbolt` as the default embedded database and to emphasize the swappable nature of the persistence layer (Store interface). This clarifies the architectural approach for database integration.
*   **2025-08-25 (Change Summary):** Completed S10.1.1: Identified independent workflow branches using a dependency graph analysis. This involved implementing `GetExecutableSteps` and resolving related YAML unmarshaling and linter issues.
*   **2025-08-25 (Change Summary):** Added new task T10.1 "Implement Concurrent and Parallel Execution" under Epic E10 to continue work on Performance and Scalability Enhancements.
*   **2025-08-24 (Change Summary):** Aligned the project plan with the updated design document. This introduces a detailed future roadmap of epics (E10-E14) covering scalability, high availability, advanced developer experience, security, and observability. The Non-Goals and Out of Scope sections were clarified to reflect that features like HA and advanced auth are part of this long-term vision.
*   **2025-08-24 (Change Summary):** Enhanced the plan to include durable execution. This is a major feature that introduces a persistence layer (E9) to ensure workflows can survive orchestrator restarts and transient tool failures. This change makes Sire a more reliable and robust platform. The plan was updated to include tasks for database integration, stateful engine logic, and retry mechanisms. The timeline and deliverables were also adjusted accordingly (D11, M7).
*   **2025-08-24 (Change Summary):** Pivoted the project plan to focus on making Sire a universal MCP orchestration engine. This is a major strategic shift from the original plan of being a Go-based n8n alternative. The new plan archives the completed initial work (E1-E4) and introduces new epics (E5-E8) to refactor the core engine, unify local and remote tool execution via MCP, and update all documentation to reflect the new vision. This change was prompted by the realization that a decoupled, MCP-first architecture is more powerful and extensible.
*   **2025-08-24:** Added Epic E4 for the MCP Server implementation.
*   **2025-08-24:** Initial version of the plan created.
*   **2025-08-24:** Added more detail to the work breakdown.
*   **2025-08-24:** Completed T1.1, S1.2.1, S1.2.2, S1.2.3, S1.2.4.

## 7. Hand off Notes

This document outlines the plan to pivot Sire into an MCP-native orchestration engine with durable execution. The core concepts are:
1.  **Decoupling:** The engine is decoupled from the tools it executes via the MCP standard.
2.  **Persistence:** The engine is stateful. It uses an embedded database (e.g., BoltDB) to persist the state of every execution after each successfully completed step.
3.  **Recovery:** A background worker process will periodically scan for incomplete workflows and re-queue them for execution, making the system resilient to crashes and transient failures.

A new developer should focus on understanding the dispatcher model (`E5.2`) and the storage service (`E9.1`) as they are the foundation of the new architecture.

## 8. Appendix

### Example Federated Workflow (`federated.yml`)

```yaml
id: federated-data-pipeline
name: "Fetch data, process it locally, and upload"
steps:
  - id: fetch_data
    tool: "mcp:http://api.third-party.com/rpc#data.fetch"
    params:
      source_id: "12345"
    retry:
      max_attempts: 5
      backoff: "exponential"

  - id: transform_data
    tool: "sire:local/data.transform"
    params:
      operation: "map"
      expression: "item.value * 2"
      data: "{{ .fetch_data.output.records }}"

  - id: upload_result
    tool: "mcp:http://storage-service.internal/rpc#s3.upload"
    params:
      bucket: "processed-results"
      key: "result-{{.workflow.id}}.json"
      body: "{{ .transform_data.output.result }}"

edges:
  - from: fetch_data
    to: transform_data
  - from: transform_data
    to: upload_result
```
