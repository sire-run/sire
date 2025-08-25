# Sire Design Philosophy

This document outlines the core design principles that guide the development of Sire. It serves as a compass for all contributors to ensure the project remains cohesive, maintainable, and true to its vision as a universal orchestrator.

## 1. Core Philosophy

- **MCP-First:** The orchestrator is, and must remain, completely decoupled from the tools it executes. The Model-Context-Protocol (MCP) is the universal contract for all tool communication.
- **Developer-First:** The primary user is a developer. Workflows are defined as code (YAML) in Git, and the CLI is a first-class interface for running, testing, and debugging.
- **Simplicity Through Unification:** Sire avoids a complex, language-specific plugin system. Extensibility is achieved by creating new tools that expose an MCP endpoint, which can be written in any language.
- **High-Performance Orchestration:** The core is built in Go to be a fast, lightweight, and low-latency dispatcher.
- **Testability by Design:** The architecture is designed to be testable at every level, from the core engine to individual tools.

## 2. The Hybrid MCP-First Architecture

To achieve both architectural purity and high performance, Sire uses a hybrid dispatch model. The engine is decoupled from the execution details, which are handled by specialized dispatchers.

### 2.1. The Orchestration Engine (`internal/core/engine.go`)

- **Single Responsibility:** The `core.Engine`'s sole job is to execute a workflow graph. It is stateless regarding tool execution.
- **Dispatcher, Not Executor:** The engine does not execute tools directly. It holds a reference to a single `core.Dispatcher` interface and delegates all execution to it.
- **Execution Flow:**
    1.  Performs a topological sort of the workflow's `Steps`.
    2.  Iterates through the steps in order.
    3.  For each `Step`, it prepares the inputs by merging initial workflow inputs with outputs from parent steps.
    4.  It calls `dispatcher.Dispatch(ctx, step.Tool, stepInputs)`.
    5.  It stores the step's output, making it available to subsequent steps.

### 2.2. The Dispatcher Abstraction (`internal/core/dispatcher.go`)

This is the critical abstraction layer.

- **`Dispatcher` Interface:** A simple, universal interface for tool execution.
    ```go
    type Dispatcher interface {
        Dispatch(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error)
    }
    ```
- **`DispatcherMux` (The Router):** The primary `Dispatcher` implementation used by the `Engine`.
    - It maintains a map of URI schemes to specific dispatchers (e.g., `{"sire": localDispatcher, "mcp": remoteDispatcher}`).
    - When called, it parses the `tool` URI (e.g., `"sire:local/file.write"`) and routes the request to the corresponding registered dispatcher based on the scheme.

### 2.3. The `sire:local` Dispatcher (In-Process Optimization)

This dispatcher provides a high-performance path for executing built-in Go tools without network overhead.

- **Mechanism:** It uses an `InProcessServer` that directly registers Go functions as MCP services. This server is **not a network server**; it executes tools via direct function calls within the same process.
- **Flow (`sire:local/file.write`):**
    1.  The `DispatcherMux` routes the call to the `InProcessDispatcher`.
    2.  The `InProcessDispatcher` calls the `InProcessServer`.
    3.  The `InProcessServer` directly invokes the registered `file.write` function.
- **Benefit:** This model provides the performance of a native function call while still enforcing the MCP contract and keeping the engine decoupled.

### 2.4. The `mcp:` Dispatcher (Remote Executor)

This dispatcher handles calls to any standard, network-accessible MCP server.

- **Mechanism:** It acts as a standard JSON-RPC 2.0 client.
- **Flow (`mcp:http://api.example.com/rpc#math.add`):**
    1.  The `DispatcherMux` routes the call to the `RemoteDispatcher`.
    2.  The dispatcher parses the URI to get the server endpoint (`http://api.example.com/rpc`) and the method name (`math.add`).
    3.  It makes a standard HTTP/JSON-RPC network request to the remote server.
    4.  It handles network errors, timeouts, and response parsing.

## 3. Opportunities for Future Improvement

This design is robust, but several areas offer opportunities for enhancement. Here is a detailed roadmap of architectural improvements.

### 3.1. Scalability and Performance

#### 3.1.1. Concurrent and Parallel Execution
The engine currently executes steps sequentially based on a topological sort. To significantly speed up I/O-bound workflows, it should be enhanced to execute independent branches of the graph concurrently. The topological sort already provides the necessary data to identify these parallelizable steps.

#### 3.1.2. Granular State Persistence & Large Data Handling
The current model of persisting the entire `Execution` object after each step can be inefficient. To scale to larger, more complex workflows:
-   **Granular Updates:** The engine should perform atomic, transactional updates to only the specific `StepState` that has changed, reducing database I/O.
-   **Artifact Store:** For large data artifacts, a dedicated `ArtifactStore` (local disk, S3, etc.) should be used. Tools would write large outputs there, and the `StepState` would only store a lightweight URI reference, keeping the core state database lean.

#### 3.1.3. Dispatcher Connection Caching
The `RemoteDispatcher` can be optimized to cache and reuse network clients (e.g., HTTP/JSON-RPC clients) for specific endpoints, avoiding the overhead of establishing new connections for every tool call within a single workflow.

### 3.2. Resilience and High Availability

#### 3.2.1. High-Availability Agent Model
The single `sire agent` is a potential point of failure. To support production workloads, a multi-agent architecture is needed:
-   **Leader Election:** Allow multiple agent instances to run, using a leader election protocol (via the persistence store) to ensure only one is actively scheduling work.
-   **Distributed Locking:** The leader agent must acquire a short-lived, renewable lock on any `Execution` it processes. This prevents race conditions and ensures that if a leader fails, another agent can safely take over.

### 3.3. Extensibility and Integration

#### 3.3.1. Pluggable Dispatchers
The registration of dispatchers is currently hard-coded. A future version should allow custom dispatchers to be compiled in as Go plugins or registered dynamically. This would enable first-class support for new protocols and services (e.g., `grpc:`, `aws-lambda:`, `gcp-run:`) without modifying the core engine.

### 3.4. Developer Experience (DX) and Workflow Definition

#### 3.4.1. Advanced Workflow Definitions (SDKs)
While the current input mapping is functional, moving beyond basic templating is key. Instead of just a more powerful templating language, Sire should provide optional **Workflow SDKs** (e.g., in Go, Python, TypeScript). This would allow developers to define complex workflows programmatically, gaining the benefits of type safety, loops, conditionals, and native tooling. The SDK would compile down to the canonical workflow model the engine executes.

#### 3.4.2. Interactive Debugging and Testing Tools
To fulfill the "Developer-First" promise, the CLI should be enhanced with:
-   **Interactive Debugger:** A `sire workflow run --debug` mode to step through execution, inspect state, and resume.
-   **Mocking and Dry-Runs:** A `--dry-run` flag to validate syntax and a `--mock-file` flag to provide canned outputs for tools, enabling isolated testing of workflow logic.

### 3.5. Security

#### 3.5.1. Comprehensive Security Model
A robust security model is essential:
-   **Credential Management:** A secure credential store should be integrated, allowing workflows to reference secrets (e.g., `{{ .secrets.my_api_key }}`) by name instead of embedding them.
-   **Workflow RBAC:** Implement Role-Based Access Control to govern who can define, execute, and inspect workflows and their results.
-   **Input/Output Schema Validation:** Allow steps to declare schemas for their inputs and outputs. The engine can enforce these schemas, preventing data corruption or injection attacks between potentially untrusted tools.

### 3.6. Production-Grade Observability

#### 3.6.1. Metrics, Tracing, and Structured Logging
To be operable in production, Sire needs deep visibility:
-   **Structured Logging:** All components must emit structured (JSON) logs with consistent context fields (`workflow_id`, `execution_id`, `step_id`).
-   **Metrics:** The engine should expose key performance indicators as Prometheus metrics (e.g., execution latency, step failure rates, queue depth).
-   **Distributed Tracing:** The dispatcher interface should propagate tracing contexts (e.g., OpenTelemetry), allowing a single trace to follow a workflow from the engine to remote MCP tools and back.

### 3.7. Durable Execution and State Persistence (Moved from Section 3)

To guarantee that workflows run to completion, the engine is **stateful and persistent**. If the `sire` process crashes or a tool fails transiently, the workflow can be resumed automatically from the last successfully completed step.

#### 3.7.1. The Persistence Layer (`internal/storage`)

A dedicated storage layer abstracts all database operations, ensuring the engine's core logic remains clean.

-   **Embedded Database:** To maintain simplicity, Sire uses an embedded key-value store (e.g., BoltDB). This keeps Sire as a single, self-contained binary.
-   **`Storage` Service:** A `storage.Store` interface defines all necessary database operations, such as `CreateExecution`, `GetExecution`, and `UpdateExecution`.

#### 3.7.2. Stateful Core Data Structures

The `Execution` and `StepState` structs are designed to capture the complete state of a workflow run, enabling reliable resumption.

```go
// Execution represents a single, durable run of a workflow.
type Execution struct {
    ID           string                 `json:"id"`
    WorkflowID   string                 `json:"workflow_id"`
    Status       ExecutionStatus        `json:"status"` // e.g., running, completed, failed, retrying
    StepStates   map[string]*StepState  `json:"step_states"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
}

// StepState represents the state of a single step in an execution.
type StepState struct {
    Status      StepStatus             `json:"status"` // e.g., pending, running, completed, failed
    Output      map[string]interface{} `json:"output,omitempty"`
    Error       string                 `json:"error,omitempty"`
    Attempts    int                    `json:"attempts"`
    NextAttempt time.Time              `json:"next_attempt,omitempty"` // For exponential backoff
}
```

#### 3.7.3. Stateful Engine and Resumption Logic

The `core.Engine` is refactored to be state-aware.

-   **Dependency:** The engine takes a `storage.Store` instance during initialization.
-   **Execution and Persistence Flow:**
    1.  A new workflow run begins by creating an `Execution` record in the database.
    2.  The engine receives this `Execution` object to process.
    3.  Before executing a step, the engine checks its status in `Execution.StepStates`. If a step is already `completed`, its execution is skipped, and its stored output is used directly. This is the key to resuming.
    4.  After a step executes successfully, the engine **immediately and atomically** updates its `StepState` and saves the entire `Execution` object back to the database before proceeding to the next step.
    5.  If a tool call fails, the engine updates the step's state to `failed` or `retrying` (based on the step's `RetryPolicy`) and persists the state.

#### 3.7.4. The Agent (`sire agent`)

A long-running agent process is responsible for driving the resumption of incomplete workflows.

-   **Scanning:** The agent periodically scans the database for executions that are in a `running` or `retrying` state.
-   **Re-queueing:** When it finds an incomplete execution (and its retry backoff period has passed), it loads the execution state from the database and submits it to the `core.Engine` for processing. The engine, following the logic above, will pick up exactly where it left off.