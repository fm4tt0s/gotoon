# gotoon - JSON-to-TOON Persistent Proxy

![Go CI](https://github.com/fm4tt0s/gotoon/actions/workflows/ci.yml/badge.svg)

A high-performance, go-based translation layer designed to bridge standard JSON payloads with the [**TOON (Token-Efficient Object Notation)**](https://openapi.com/blog/what-the-toon-format-is-token-oriented-object-notation) format. This proxy is built for high-throughput environments where token efficiency and persistent connection management are critical, such as **LLM context window optimization**, resource-constrained **IoT signaling**, or high-scale **SRE logging**.

## Key Architectural Features

* **Generic Tabular Transformation**: Automatically detects uniform JSON arrays and converts them into the TOON tabular format (`key[count]{fields}:`) for significant token savings.
* **Persistent Connection Management**: Reuses a single backend TCP pipe to the TOON target, minimizing the overhead of repeated TCP handshakes.
* **SRE-Focused Observability**: Built-in thread-safe metrics for real-time tracking of transaction throughput and connection stability.
* **Self-Healing Heartbeat**: A background Go routine sends periodic pulses to keep connections "hot" and prevent stateful firewalls from killing idle pipes.
* **Thread-Safe Concurrency**: Employs `sync.Mutex` protection to ensure strict data ordering and prevent payload interleaving during high-concurrency bursts.

---

## How it Works

The proxy acts as a middleman. It listens for JSON over a TCP socket, parses it into a schema-less interface, recursively translates it into the TOON format, and forwards it through a persistent shared connection.

### Example Transformation

**Input (JSON):**

```json
{
  "sensors": [
    {"id": "A1", "value": 22.5},
    {"id": "B2", "value": 19.8}
  ]
}

```

**Output (TOON):**

```text
sensors[2]{id,value}:
  A1,22.5
  B2,19.8

```

---

## Getting Started

### Prerequisites

* Go 1.18+
* Docker & Docker Compose (optional)

### Quick Start (Docker)

The most efficient way to test the proxy with a mock target:

```bash
docker-compose up --build

```

### Manual Installation

```bash
go build -o gotoon
./gotoon -l 8080 -t 127.0.0.1:9999

```

| Flag | Default | Description |
| --- | --- | --- |
| `-l` | `8080` | Local port to listen for incoming JSON |
| `-t` | `127.0.0.1:9999` | Destination TOON server address |

---

## Token Efficiency Benchmarks

TOON’s primary value proposition is the aggressive reduction of character counts by stripping redundant keys and structural overhead. This is essential for lowering LLM API costs and improving response latency.

| Data Scenario | JSON Size (chars) | TOON Size (chars) | **Reduction (%)** |
| --- | --- | --- | --- |
| Single Object (5 fields) | 124 | 78 | **37%** |
| Uniform Array (10 items) | 840 | 312 | **63%** |
| Deeply Nested Structure | 512 | 240 | **53%** |

---

## Observability & Monitoring

The proxy provides a **summary** to the logs every 60 seconds. This visibility allows for quick assessment of system health and connection flapping.

```text
2026/03/04 09:35:00 [METRICS] Req: 154 | HB Success: 12 | HB Fail: 0

```

* **Req**: Total JSON transactions successfully translated and forwarded.
* **HB Success**: Successful keep-alive pulses sent to target.
* **HB Fail**: Connection drops detected by the heartbeat mechanism.

---

## Design Decisions & Challenges

### Mutex vs. Channels

While Go often favors channels, this architecture uses a `sync.Mutex` to guard the persistent `net.Conn`. This ensures **strict ordering** of data on the wire, preventing binary "interleaving" that can occur with high-concurrency channel consumers sharing a single socket.

### Persistent vs. Ephemeral Pipes

To minimize latency for time-sensitive AI agents, a persistent pipe was chosen over opening a new connection for every request. A background heartbeat (sending `\n`) maintains the connection during idle periods, ensuring the proxy is always ready for immediate delivery.

---

## 🏗️ CI/CD & Testing

The repository includes a **Go CI** pipeline via GitHub Actions.

* **Linting**: Ensures code adheres to `gofmt` standards.
* **Unit Testing**: Validates translation logic and tabular detection via `main_test.go`.
* **Build Verification**: Confirms cross-platform compilation on every push.

---

## license

MIT