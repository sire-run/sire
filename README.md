# Sire: The Universal MCP Orchestrator

**Build reliable, distributed workflows that orchestrate tools across any service, language, or infrastructure.**

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![Go Reference](https://pkg.go.dev/badge/github.com/sire-run/sire.svg)](https://pkg.go.dev/github.com/sire-run/sire)
[![Go Report Card](https://goreportcard.com/badge/github.com/sire-run/sire)](https://goreportcard.com/report/github.com/sire-run/sire)

---

Sire is a **production-ready orchestration engine** that connects and coordinates tools across distributed systems using the **Model Context Protocol (MCP)**. Born from the need for reliable, language-agnostic automation, Sire eliminates the complexity of traditional workflow engines while providing enterprise-grade durability guarantees.

**Why Sire exists:** Modern development involves dozens of tools and services. Coordinating them shouldn't require writing language-specific plugins, managing complex SDKs, or accepting brittle, fire-and-forget automation. Sire solves this with a universal protocol and bulletproof execution.

```
┌─────────────┐    MCP     ┌─────────────────┐    MCP     ┌─────────────┐
│ Your API    │ ◄────────► │ Sire            │ ◄────────► │ External    │
│ Service     │            │ Orchestrator    │            │ Service     │
└─────────────┘            └─────────────────┘            └─────────────┘
                                    │
                           ┌─────────────────┐
                           │ Durable State   │
                           │ (Embedded DB)   │
                           └─────────────────┘
```

## Why Choose Sire?

**The problem with existing solutions:**
- **Zapier/n8n**: Great for simple automations, but limited by pre-built connectors and UI-first design
- **Temporal**: Powerful but complex, requiring SDK adoption and distributed infrastructure
- **Custom scripts**: Fragile, hard to monitor, and prone to data loss on failures

**Sire's approach:**

| Feature | Sire | Traditional Solutions |
|---------|------|----------------------|
| **🔌 Extensibility** | **Universal MCP protocol** - any language, any service | Platform-specific plugins or SDKs |
| **💾 Reliability** | **Built-in durability** - automatic state persistence and recovery | Manual error handling or complex infrastructure |
| **⚡ Performance** | **Single binary** - lightweight, fast startup, minimal overhead | Heavy runtimes or distributed complexity |
| **🛠️ Developer Experience** | **Workflows as code** - version controlled YAML with CLI tooling | Visual editors or verbose SDK code |
| **🔒 Production Ready** | **Enterprise features** - retries, monitoring, concurrent execution | Basic automation with limited resilience |

## Production-Grade Features

### 🔄 **Durable Execution**
- **Automatic state persistence** after every step
- **Crash recovery** - workflows resume exactly where they left off  
- **At-least-once execution** guarantees with embedded database
- **Background worker** for processing failed/retrying executions

### 🚀 **High Performance**
- **Concurrent execution** of independent workflow branches
- **Connection pooling** for remote MCP services
- **Optimized state management** with granular updates
- **Single binary deployment** - no external dependencies

### 🔧 **Developer Experience**
- **Intuitive CLI** with execution monitoring and debugging
- **Template engine** for dynamic parameter injection
- **Comprehensive retry policies** with exponential backoff
- **Workflow validation** and dry-run capabilities

### 🌐 **Universal Integration**
- **Built-in tools** (file operations, HTTP requests, data transformation)
- **Remote MCP services** via standard JSON-RPC protocol
- **Custom tools** in any language that can serve MCP
- **Service discovery** and connection management

## Real-World Example: Data Processing Pipeline

Here's a production workflow that demonstrates Sire's power - it orchestrates tools across three different systems with built-in resilience:

```yaml
id: production-data-pipeline
name: "Multi-Service Data Processing with Error Recovery"
steps:
  - id: fetch_customer_data
    tool: "mcp:https://api.customer-service.com/rpc#customers.export"
    params:
      date_range: "{{ .workflow.params.date_range }}"
      format: "json"
    retry:
      max_attempts: 5
      backoff: "exponential"
      initial_delay: "2s"

  - id: validate_and_clean
    tool: "sire:local/data.transform"
    params:
      operation: "validate_schema"
      schema_file: "/schemas/customer.json"
      data: "{{ .fetch_customer_data.output.customers }}"

  - id: enrich_with_analytics
    tool: "mcp:http://analytics.internal:8080/rpc#data.enrich"
    params:
      dataset: "{{ .validate_and_clean.output.valid_records }}"
      enrichment_rules: ["demographic", "behavioral"]
    retry:
      max_attempts: 3

  - id: store_to_warehouse
    tool: "mcp:tcp://warehouse.company.com:9090#warehouse.bulk_insert"
    params:
      table: "customers_enriched"
      data: "{{ .enrich_with_analytics.output.enriched_data }}"
      upsert_key: "customer_id"

  - id: trigger_downstream_jobs
    tool: "sire:local/http.post"
    params:
      url: "https://scheduler.company.com/api/jobs/trigger"
      headers:
        Authorization: "Bearer {{ .workflow.secrets.scheduler_token }}"
      body:
        job_ids: ["customer-segmentation", "ml-feature-refresh"]
        trigger_reason: "data-pipeline-{{ .workflow.execution_id }}"

edges:
  - from: fetch_customer_data
    to: validate_and_clean
  - from: validate_and_clean
    to: enrich_with_analytics
  - from: enrich_with_analytics
    to: [store_to_warehouse, trigger_downstream_jobs]  # Parallel execution
```

**What makes this powerful:**
- **Fault tolerance**: If any step fails, the workflow automatically retries with exponential backoff
- **State preservation**: Stop the orchestrator mid-execution, restart it - the workflow continues exactly where it left off
- **Cross-service orchestration**: Seamlessly coordinates HTTP APIs, TCP services, and local tools
- **Parallel execution**: Steps 4 and 5 run concurrently for optimal performance
- **Template engine**: Dynamic parameter injection and secret management

## Getting Started

### Quick Install

**Option 1: Download Binary (Recommended)**
```bash
# Download the latest release for your platform
curl -L https://github.com/sire-run/sire/releases/latest/download/sire-linux-amd64 -o sire
chmod +x sire
sudo mv sire /usr/local/bin/
```

**Option 2: Build from Source**
```bash
git clone https://github.com/sire-run/sire.git
cd sire
go build -o sire ./cmd/sire
```

### Your First Workflow

Create a simple workflow that demonstrates local and remote tool coordination:

**1. Create `hello-world.yml`:**
```yaml
id: hello-world
name: "Hello World Workflow"
steps:
  - id: generate_message
    tool: "sire:local/data.transform"
    params:
      operation: "template"
      template: "Hello from Sire! Current time: {{ .workflow.started_at }}"

  - id: save_locally
    tool: "sire:local/file.write"
    params:
      path: "/tmp/sire-hello.txt"
      content: "{{ .generate_message.output.result }}"

  - id: post_to_webhook
    tool: "sire:local/http.post"
    params:
      url: "https://httpbin.org/post"
      headers:
        Content-Type: "application/json"
      body:
        message: "{{ .generate_message.output.result }}"
        source: "sire-workflow"

edges:
  - from: generate_message
    to: [save_locally, post_to_webhook]  # Run in parallel
```

**2. Run the workflow:**
```bash
sire workflow run hello-world.yml
```

**3. Monitor execution:**
```bash
# List all executions
sire execution list

# Get detailed status
sire execution status <execution-id>

# Watch real-time progress
sire execution watch <execution-id>
```

### Working with Remote Services

Sire's real power comes from orchestrating remote MCP services. Here's how to integrate with external tools:

**Example: GitHub + Slack Integration**
```yaml
id: github-slack-notify
name: "GitHub Issue to Slack Notification"
steps:
  - id: fetch_github_issues
    tool: "mcp:https://github-mcp.example.com/rpc#issues.list"
    params:
      repo: "myorg/myrepo"
      state: "open"
      labels: ["critical", "bug"]
    retry:
      max_attempts: 3

  - id: format_message
    tool: "sire:local/data.transform"
    params:
      operation: "template"
      template: |
        🚨 Critical Issues Alert:
        Found {{ len .fetch_github_issues.output.issues }} critical bugs
        {{- range .fetch_github_issues.output.issues }}
        • #{{ .number }}: {{ .title }}
        {{- end }}

  - id: send_slack_alert
    tool: "mcp:tcp://slack-service.internal:9091#messages.send"
    params:
      channel: "#engineering-alerts"
      message: "{{ .format_message.output.result }}"

edges:
  - from: fetch_github_issues
    to: format_message
  - from: format_message
    to: send_slack_alert
```

### CLI Commands Overview

```bash
# Workflow management
sire workflow run <file>              # Execute a workflow
sire workflow validate <file>         # Validate workflow syntax
sire workflow list                    # List available workflows

# Execution monitoring
sire execution list                   # Show all executions
sire execution status <id>            # Get execution details
sire execution watch <id>             # Real-time status updates
sire execution logs <id>              # View execution logs
sire execution retry <id>             # Retry failed execution

# Tool discovery
sire tools list                       # List available local tools
sire tools discover <mcp-url>         # Discover remote MCP tools
sire tools test <tool-uri>            # Test tool connectivity

# System management
sire daemon start                     # Start background worker
sire daemon stop                      # Stop background worker
sire daemon status                    # Check daemon status
```


## Architecture & Advanced Features

### MCP Protocol Support
Sire implements the complete Model Context Protocol specification:
- **JSON-RPC 2.0** communication with remote services  
- **HTTP, TCP, and WebSocket** transport protocols
- **Tool discovery** and capability negotiation
- **Streaming responses** for real-time data processing
- **Authentication** and secure credential management

### Built-in Tools
Sire comes with production-ready tools out of the box:

| Tool Category | Available Tools | Description |
|--------------|----------------|-------------|
| **File Operations** | `file.read`, `file.write`, `file.copy`, `file.delete` | Cross-platform file system operations |
| **HTTP/REST** | `http.get`, `http.post`, `http.put`, `http.delete` | Full HTTP client with authentication |
| **Data Processing** | `data.transform`, `data.validate`, `data.filter` | JSON manipulation and validation |
| **Templating** | `template.render`, `template.validate` | Go template engine with helper functions |
| **System** | `system.exec`, `system.env`, `system.wait` | System command execution and environment |

### Enterprise Features

**Security & Compliance**
- **Credential vault** integration (HashiCorp Vault, AWS Secrets Manager)
- **RBAC** (Role-Based Access Control) for workflow permissions
- **Audit logging** with tamper-proof execution trails
- **Input/output schema validation** for data integrity

**Observability & Monitoring** 
- **Structured logging** (JSON) with contextual information
- **Prometheus metrics** for performance monitoring
- **Distributed tracing** (OpenTelemetry) across all MCP calls
- **Health checks** and service discovery integration

**High Availability** 
- **Leader election** for multi-instance deployments  
- **Distributed locking** to prevent duplicate execution
- **Load balancing** across multiple orchestrator instances
- **Automatic failover** with state synchronization

## Roadmap

### ✅ **Current (v1.0) - Production Ready**
- Complete MCP orchestration engine with durable execution
- Built-in tools and remote MCP service integration  
- CLI with execution monitoring and debugging
- Concurrent workflow execution with dependency management
- Comprehensive retry policies and error recovery

### 🚧 **Next (v1.1) - Performance & Scale**
- **Granular state persistence** for handling large data payloads
- **Connection pooling** and dispatcher optimizations  
- **Interactive debugger** with step-through capabilities
- **Workflow mocking** and dry-run testing

### 🔮 **Future (v1.2+) - Enterprise & Ecosystem**
- **Web UI** for workflow visualization and monitoring
- **Community tool registry** for sharing MCP services
- **Kubernetes operator** for cloud-native deployments
- **GraphQL API** for programmatic workflow management
- **Workflow templates** and reusable component library

## Creating Custom MCP Tools

Extend Sire with custom tools in any language. Here's a minimal Python MCP server:

```python
#!/usr/bin/env python3
import json
from jsonrpc import JSONRPCResponseManager, dispatcher

@dispatcher.add_method
def text_analyze(text):
    """Analyze text and return statistics"""
    return {
        "word_count": len(text.split()),
        "char_count": len(text),
        "sentiment": "positive" if "good" in text.lower() else "neutral"
    }

if __name__ == "__main__":
    # Simple HTTP MCP server
    from http.server import HTTPServer, BaseHTTPRequestHandler
    
    class MCPHandler(BaseHTTPRequestHandler):
        def do_POST(self):
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            response = JSONRPCResponseManager.handle(post_data.decode(), dispatcher)
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(response.json.encode())
    
    HTTPServer(('localhost', 8080), MCPHandler).serve_forever()
```

**Use in workflows:**
```yaml
steps:
  - id: analyze_feedback
    tool: "mcp:http://localhost:8080#text_analyze"
    params:
      text: "This product is really good and helpful!"
```

## Use Cases

**DevOps & Infrastructure**
- **CI/CD orchestration** across multiple deployment environments
- **Infrastructure provisioning** with Terraform + cloud APIs
- **Monitoring alert workflows** that auto-remediate common issues

**Data Engineering** 
- **ETL pipelines** that coordinate databases, APIs, and data warehouses
- **Real-time data processing** with stream processors and message queues
- **Data quality monitoring** with automated validation and alerting

**Business Process Automation**
- **Customer onboarding** workflows spanning CRM, billing, and support systems  
- **Invoice processing** that integrates accounting, approval, and payment systems
- **Content publishing** pipelines that coordinate CMS, CDN, and social platforms

**Integration & Migration**
- **System migrations** that coordinate data movement and validation
- **API orchestration** for microservice communication patterns
- **Legacy system integration** with modern cloud services

## Community & Support

- **📖 Documentation:** [docs.sire.run](https://docs.sire.run)
- **💬 Discord Community:** [discord.gg/sire](https://discord.gg/sire)  
- **🐛 Issues & Feature Requests:** [GitHub Issues](https://github.com/sire-run/sire/issues)
- **📧 Enterprise Support:** enterprise@sire.run

## Contributing

We welcome contributions! Areas where you can help:

- **🔧 MCP Tools:** Create and share MCP servers for popular services
- **📚 Documentation:** Improve guides, examples, and tutorials  
- **🐛 Bug Reports:** Help us identify and fix issues
- **💡 Feature Ideas:** Suggest enhancements and new capabilities
- **🧪 Testing:** Write tests and help improve reliability

See our [Contributing Guide](CONTRIBUTING.md) for details.

## License

Sire is licensed under the [Apache License 2.0](LICENSE). This means you can use it freely in commercial and open-source projects.
