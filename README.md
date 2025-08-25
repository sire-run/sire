# Sire: The Universal MCP Orchestrator

**Orchestrate anything, from anywhere, with code.**

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

---

Sire is a new kind of automation platform for developers. It's a lightweight, high-performance orchestrator that connects and runs tools using the **Model-Context-Protocol (MCP)**, an open, JSON-RPC-based standard. 

Instead of a heavy, UI-focused platform or a complex, SDK-based engine, Sire offers a simple, powerful alternative: **define workflows in YAML, run them from the CLI, and extend the system in any language** by creating a simple MCP server for your tool.

```
(Your Service) <--> MCP <--> [Sire Orchestrator] <--> MCP <--> (Another Service)
```

## The Sire Difference

How is Sire different from...?

| Category | Sire (The Orchestrator) | Zapier / n8n (The UI Platform) | Temporal (The Workflow Engine) |
| :--- | :--- | :--- | :--- |
| **Philosophy** | **Orchestration as Code.** Declarative YAML files that live in your Git repo. | **Automation as a UI.** Visual, drag-and-drop interface for non-developers. | **Workflows as Code.** Imperative code written with a proprietary, language-specific SDK. |
| **Extensibility** | **Language-agnostic via MCP.** Expose your tool via a simple, open protocol. No plugins to write. | **Language-specific plugins.** Must write plugins in their required language (e.g., TypeScript) and conform to their complex SDK. | **Language-specific SDK.** Requires instrumenting your code with their SDK. Tightly coupled. |
| **Architecture** | **Lightweight & Decoupled.** A single, simple binary that orchestrates any MCP-compliant tool. | **Monolithic.** A large, all-in-one platform that includes the UI, engine, and a vast library of built-in nodes. | **Heavy & Distributed.** A complex, cluster-based system that requires separate deployment and management. |
| **Durability** | **Built-in.** Uses an embedded database to guarantee at-least-once execution of every step. | **Managed Service.** Durability is handled by the platform, but you have little control over it. | **Core Feature.** Provides strong guarantees, but requires careful adherence to deterministic coding rules. |

## Core Principles

*   **MCP-First Architecture:** The orchestrator is completely decoupled from the tools it executes. MCP is the universal contract.
*   **Workflows as Code:** Workflows are defined in declarative YAML, designed to be version-controlled, reviewed, and edited as code.
*   **Durable & Reliable:** With an embedded database, Sire guarantees that your workflows will run to completion, even if the orchestrator or the tools it calls restart or fail transiently.
*   **High-Performance Orchestration:** The compiled Go core is designed to be a lean, low-latency dispatcher, ensuring your workflows run efficiently.

## Example: A Federated Workflow

Sire shines at orchestrating tools from different sources. This workflow fetches data from a remote public API, processes it with a local Sire tool, and then calls a separate internal service to store the result.

```yaml
id: federated-data-pipeline
name: "Fetch, Process, and Store Data"
steps:
  - id: fetch_public_data
    # Call a tool on a remote, third-party MCP server
    tool: "mcp:http://api.public-data.com/rpc#data.fetch"
    params:
      source_id: "12345"
    retry:
      max_attempts: 3

  - id: transform_the_data
    # Use a built-in, high-performance local tool
    tool: "sire:local/data.transform"
    params:
      operation: "map"
      expression: "item.value * 2"
      data: "{{ .fetch_public_data.output.records }}"

  - id: store_in_archive
    # Call a tool on our internal microservice
    tool: "mcp:http://archiver.internal.svc/rpc#s3.upload"
    params:
      bucket: "processed-results"
      body: "{{ .transform_the_data.output.result }}"

edges:
  - from: fetch_public_data
    to: transform_the_data
  - from: transform_the_data
    to: store_in_archive
```

## Getting Started

### Prerequisites

*   Go 1.25 or later.

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/sire-run/sire.git
    cd sire
    ```
2.  Build the `sire` CLI:
    ```bash
    go build -o sire ./cmd/sire
    ```

### Running a Workflow

```bash
# The sire binary is now stateful and requires a database file.
```bash
# The sire binary is now stateful and requires a database file.
./sire --db-path sire.db workflow run -f my_workflow.yml
```
```

## High-Level Roadmap

*   **v0.1 (Current):** MCP-first architecture, in-process tool server, remote tool execution, with initial dispatcher implementation.
*   **v0.2 (Planned):** Enhanced CLI for managing executions, improved templating engine, and official documentation for creating MCP tool servers.
*   **Future:** Community tool registry, and optional web UI for monitoring executions.

## Join the Community

*   **GitHub Discussions:** [Link to be added]
*   **Discord Server:** [Link to be added]

## Contributing

We welcome contributions of all kinds! Please see our (upcoming) `CONTRIBUTING.md` for more details.

## License

Sire is licensed under the [Apache License 2.0](LICENSE.md).
