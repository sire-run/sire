# Sire

**A modern, high-performance workflow automation platform, written in Go.**

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

---

Sire is a powerful, yet simple workflow automation engine designed for developers. If you've ever felt that existing automation platforms are too complex, have a steep learning curve, or require too much boilerplate to get started, Sire is for you.

Our goal is to provide a developer-first experience, prioritizing performance, simplicity, and extensibility. We believe that workflow automation should be as simple as writing a script, but with the power and structure of a modern platform.

## Why Sire?

In a world with established players like [n8n](https://n8n.io/), why build another workflow automation tool? The answer lies in our core principles:

*   **Performance by Default:** Built in Go, Sire is designed for speed and concurrency. It can handle a high volume of workflows with minimal resource consumption, making it ideal for performance-critical applications.

*   **Developer-Centric Simplicity:** We've stripped away the complexity often found in other platforms. With Sire, there's less boilerplate and more convention. Creating a new node is as simple as implementing a Go interface, without the need for a complex plugin system or a fragmented multi-package architecture.

*   **Transparent and Extensible:** Sire is built on a clean and cohesive architecture. The API for creating custom nodes is straightforward and well-defined, making it easy to extend the platform with your own custom logic.

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

### Your First Workflow

Create a file named `hello_workflow.yml` with the following content:

```yaml
id: hello-world
name: My First Workflow
nodes:
  start:
    type: "file.write"
    config:
      path: "hello.txt"
      content: "Hello, from Sire!"
  log_message:
    type: "data.transform"
    config:
      operation: "map"
      expression: "'Workflow executed successfully and wrote to hello.txt'"
edges:
  - from: start
    to: log_message
```

This workflow will:
1.  Create a file named `hello.txt` with the content "Hello, from Sire!".
2.  Use the `data.transform` node to create a success message.

### Running the Workflow

Execute the workflow using the `sire` CLI:

```bash
./sire workflow run -f hello_workflow.yml
```

You should see a `hello.txt` file in the same directory, and the output of the execution printed to the console.

## Core Concepts

### Workflows

Workflows are defined in YAML or JSON files. They consist of a set of nodes and the edges that connect them.

### Nodes

Nodes are the building blocks of a workflow. Each node has a `type` and a `config` block. The `type` corresponds to a registered node implementation, and the `config` provides the static configuration for that node.

### Expressions

Sire uses the [Expr](https://github.com/expr-lang/expr) language for powerful data transformations. You can use it in nodes like `data.transform` to perform `map`, `filter`, and `reduce` operations on your data.

## Built-in Nodes

Sire comes with a set of built-in nodes to get you started:

*   `http.request`: Make HTTP requests (GET, POST, PUT, DELETE).
*   `file.read`: Read the content of a file.
*   `file.write`: Write content to a file.
*   `data.transform`: Perform `map`, `filter`, and `reduce` operations on data.

## Contributing

We welcome contributions of all kinds! Whether it's a bug report, a new feature, or a new node, we'd love to have your help. Please see our (upcoming) `CONTRIBUTING.md` for more details.

## License

Sire is licensed under the [MIT License](LICENSE.md).
