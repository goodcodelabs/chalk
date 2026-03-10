# Chalk

Vue 3 admin UI + Go HTTP bridge for managing the Slate LLM agent orchestration service.

## Architecture

```
Slate TCP service (:4242)
        ↑
bridge/main.go  — Go HTTP server (:8080), translates REST ↔ Slate JSON-TCP
        ↑
ui/  — Vue 3 + TypeScript + Tailwind CSS admin (:5173 in dev)
```

## Running

```sh
# 1. Start Slate
cd ~/d/slate && make run

# 2. Start bridge
make run-bridge          # builds then runs ./bin/bridge on :8080

# 3. Start UI dev server
make dev-ui              # http://localhost:5173
```

## Common make targets

| Target | What it does |
|---|---|
| `make bridge` | Compile bridge binary to `bin/bridge` |
| `make run-bridge` | Build + run the bridge |
| `make dev-ui` | Vite dev server with HMR |
| `make build` | Production build (bridge + UI) |
| `make clean` | Remove `bin/` and `ui/dist/` |

## Slate Wire Protocol

Every command is a JSON line in, JSON line out:

```
→ {"cmd":"ls_workspaces","params":{}}\n
← {"ok":true,"data":{"workspaces":[...]}}\n
← {"ok":false,"error":"some message"}\n
```

No-data success: `{"ok":true}` (data field absent).

Async commands (`chat`, `agent_chat`, `run_agent`) return `{"job_id":"..."}` immediately. Poll `job_result` for the actual response.

## Bridge (`bridge/main.go`)

- `sendCmd(cmd, params)` — JSON encode/send, decode response envelope, return `.data`
- `passthrough(w, slate, cmd, params)` — send cmd and write `.data` directly to HTTP response
- Single persistent TCP connection with mutex + 2-attempt auto-reconnect
- All `/api/*` routes handled by one `ServeHTTP` dispatching to per-resource handlers

## UI (`ui/src/`)

| Path | Purpose |
|---|---|
| `api/client.ts` | All API calls (typed fetch wrappers) |
| `api/types.ts` | TypeScript interfaces |
| `router/index.ts` | Vue Router routes |
| `components/` | AppLayout, NavItem, PageHeader, Modal, StatusBadge |
| `views/` | Dashboard, Workspaces, WorkspaceDetail, Catalogs, CatalogDetail, AgentDetail, ThreadChat, AgentThreadChat, Jobs |

Chat views poll `api.jobResult(job_id)` every 500 ms after sending a message (chat is async).

## Key Slate Commands

```
ls_workspaces               add_workspace {name}         del_workspace {name}
ls_catalogs                 add_catalog {name}            del_catalog {name}
add_agent {catalog_id,name} set_agent_instructions       set_agent_model
ls_agent_threads {agent_id} new_agent_thread             agent_chat → job_id
ls_threads {workspace_id}   new_thread                   chat → job_id
ls_pipelines {workspace_id} create_pipeline              run_pipeline
run_agent {agent_id,input}  → job_id
ls_jobs                     job_status {job_id}          job_result {job_id}
```
