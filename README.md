# Sire

**A modern, high-performance workflow automation platform, built for developers.**

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

---

Sire is a powerful, yet simple workflow automation engine designed for developers who value performance, simplicity, and control. If you've ever felt that existing automation platforms are too heavy, too focused on UIs, or require a complex plugin architecture, Sire is for you.

Our goal is to provide a developer-first experience, treating workflows as code and prioritizing performance, simplicity, and extensibility.

```
 (You) --> (CLI or MCP Host) --> [ Sire Engine ]
```

## The Sire Difference

In a world with established players like [n8n](https://n8n.io/), Sire takes a fundamentally different approach:

| Feature                 | Sire (The Developer Way)                               | Traditional Platforms (e.g., n8n)         |
| ----------------------- | ------------------------------------------------------ | ----------------------------------------- |
| **Core Philosophy**     | Code-first, developer-centric                          | UI-first, visual-centric                  |
| **Primary Interface**   | CLI & API (via MCP)                                    | Visual Drag-and-Drop UI                   |
| **Performance**         | High-performance, concurrent Go backend                | Node.js backend                           |
| **Node Development**    | Simple Go interface; compile into a single binary      | Complex, multi-package plugin system      |
| **Workflow Definition** | Human-readable YAML/JSON, perfect for Git              | Primarily JSON stored in a database       |

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

### Your First Workflow: Fetch a Joke

Let's create a workflow that fetches a random joke from a public API and saves it to a file. Create a file named `joke_workflow.yml`:

```yaml
id: fetch-a-joke
name: Fetch a Joke Workflow
nodes:
  get_joke:
    type: "http.request"
    config:
      method: "GET"
      url: "https://official-joke-api.appspot.com/random_joke"

  save_joke:
    type: "file.write"
    config:
      path: "joke.txt"
      content: "{{ .get_joke.output.body }}"

edges:
  - from: get_joke
    to: save_joke
```

### Running the Workflow

Execute the workflow from your terminal:

```bash
./sire workflow run -f joke_workflow.yml
```

You'll find a new file, `joke.txt`, in the directory containing the joke!

## The MCP Server: AI-Powered Workflow Generation

Sire includes a built-in **Model-Context-Protocol (MCP) server**, a groundbreaking feature that makes workflow creation accessible to everyone.

*   **What it is:** The MCP server exposes Sire's capabilities as a set of tools that any MCP-compliant host (like an AI chatbot or an IDE plugin) can use.
*   **How it works:** The host application connects to an LLM of your choice. The LLM can then use the tools provided by the Sire MCP server to generate complex workflow files based on a simple prompt.
*   **The Future:** This decouples the workflow engine from the UI, allowing for natural language workflow generation and seamless integration into future AI-powered development environments.

To start the server:
```bash
go build -o mcp-server ./cmd/mcp-server
./mcp-server
```

## High-Level Roadmap

*   **v0.1 (Current):** Core Engine, CLI, initial set of built-in nodes, and a functional MCP server.
*   **v0.2 (Planned):** Expanded set of built-in nodes, improved expression and templating engine, and official documentation.
*   **Future:** Community node marketplace, advanced scheduling and trigger options, and potential for a lightweight, optional web UI (as a separate application).

## Join the Community

*   **GitHub Discussions:** [Link to be added]
*   **Discord Server:** [Link to be added]
*   **Twitter:** [Link to be added]

## Contributing

We welcome contributions of all kinds! Whether it's a bug report, a new feature, or a new node, we'd love to have your help. Please see our (upcoming) `CONTRIBUTING.md` for more details.

## License

Sire is licensed under the [Apache License 2.0](LICENSE.md).