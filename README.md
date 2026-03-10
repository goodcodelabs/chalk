# Chalk — Slate Admin UI

A Vue 3 admin interface for [Slate](../slate/), the TCP-based LLM agent orchestration server.

## Architecture

```
Slate (TCP :4242)
    ↕  JSON-TCP protocol
bridge/   — Go HTTP server (:8080), translates REST ↔ Slate commands
    ↕  REST/JSON
ui/       — Vue 3 + TypeScript + Tailwind CSS admin UI (dev: :5173)
```

## Quick Start

**1. Start the Slate server** (in a separate terminal):
```bash
cd ../slate && make run
```

**2. Start the bridge**:
```bash
make run-bridge
```

**3. Start the UI dev server**:
```bash
make dev-ui
# Open http://localhost:5173
```

## Development

```bash
make bridge      # build the bridge binary → bin/bridge
make build-ui    # build the Vue UI for production → ui/dist/
make build       # build everything
make clean       # remove build artifacts
```

## Bridge Environment Variables

| Variable      | Default          | Description                |
|---------------|------------------|----------------------------|
| `SLATE_ADDR`  | `localhost:4242` | Slate TCP server address   |
| `LISTEN_ADDR` | `:8080`          | Bridge HTTP listen address |

## REST API (Bridge)

Async commands (`chat`, `agent_chat`, `run_agent`) return `{"job_id":"..."}` immediately.
Poll `GET /api/jobs/:id/result` until `status` is `completed` or `failed`.

| Method | Path                          | Body                    | Response                  |
|--------|-------------------------------|-------------------------|---------------------------|
| GET    | /api/health                   |                         | `{"status":"ok"}`         |
| GET    | /api/stats                    |                         | system stats object       |
| GET    | /api/workspaces               |                         | `{workspaces:[...]}`      |
| POST   | /api/workspaces               | `{name}`                | `{status:"ok"}`           |
| DELETE | /api/workspaces/:name         |                         | `{status:"ok"}`           |
| PUT    | /api/workspaces/:id/catalog   | `{catalog_id}`          | `{status:"ok"}`           |
| PUT    | /api/workspaces/:id/router    | `{agent_id}`            | `{status:"ok"}`           |
| GET    | /api/workspaces/:id/threads   |                         | `{threads:[...]}`         |
| POST   | /api/workspaces/:id/threads   | `{name?}`               | `{id, name}`              |
| GET    | /api/workspaces/:id/pipelines |                         | `{pipelines:[...]}`       |
| POST   | /api/workspaces/:id/pipelines | `{name}`                | `{pipeline_id}`           |
| GET    | /api/workspaces/:id/jobs      |                         | `[...]` (job array)       |
| GET    | /api/catalogs                 |                         | `{catalogs:[...]}`        |
| POST   | /api/catalogs                 | `{name}`                | `{status:"ok"}`           |
| DELETE | /api/catalogs/:name           |                         | `{status:"ok"}`           |
| POST   | /api/catalogs/:id/agents      | `{name}`                | `{id, name}`              |
| PUT    | /api/agents/:id/instructions  | `{instructions}`        | `{status:"ok"}`           |
| PUT    | /api/agents/:id/model         | `{model}`               | `{status:"ok"}`           |
| POST   | /api/agents/:id/run           | `{input}`               | `{job_id}` *(async)*      |
| GET    | /api/agents/:id/threads       |                         | `{threads:[...]}`         |
| POST   | /api/agents/:id/threads       | `{name?}`               | `{id, name}`              |
| GET    | /api/agents/:id/tools         |                         | `{tools:[...]}`           |
| POST   | /api/agents/:id/tools         | `{tool}`                | `{status:"ok"}`           |
| DELETE | /api/agents/:id/tools/:tool   |                         | `{status:"ok"}`           |
| POST   | /api/threads/:id/chat         | `{message}`             | `{job_id}` *(async)*      |
| GET    | /api/threads/:id/history      |                         | `{messages:[...]}`        |
| POST   | /api/agent-threads/:id/chat   | `{message}`             | `{job_id}` *(async)*      |
| GET    | /api/agent-threads/:id/history|                         | `{messages:[...]}`        |
| POST   | /api/pipelines/:id/steps      | `{agent_id, mode}`      | `{status:"ok"}`           |
| POST   | /api/pipelines/:id/run        | `{input}`               | `{job_id}`                |
| GET    | /api/jobs                     |                         | `[...]` (job array)       |
| GET    | /api/jobs/:id/status          |                         | `{status, ...}`           |
| GET    | /api/jobs/:id/result          |                         | `{status, result, error}` |
| POST   | /api/jobs/:id/cancel          |                         | `{status:"ok"}`           |

## Slate Wire Protocol

The bridge communicates with Slate over a persistent TCP connection using newline-delimited JSON:

```
→ {"cmd":"ls_workspaces","params":{}}\n
← {"ok":true,"data":{"workspaces":[...]}}\n

→ {"cmd":"chat","params":{"thread_id":"abc","message":"hello"}}\n
← {"ok":true,"data":{"job_id":"xyz"}}\n

← {"ok":false,"error":"workspace not found"}\n
```

Responses with no data payload return `{"ok":true}` (no `data` field).
