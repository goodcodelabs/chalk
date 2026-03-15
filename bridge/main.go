// Bridge is an HTTP server that translates REST calls into Slate TCP commands.
// It runs on port 8080 and connects to the Slate server at localhost:4242.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ---- Slate connection -------------------------------------------------------

// slateConn holds a single persistent TCP connection to Slate.
// It reconnects automatically on failure.
type slateConn struct {
	mu     sync.Mutex
	conn   net.Conn
	reader *bufio.Reader
	addr   string
}

func newSlateConn(addr string) *slateConn {
	return &slateConn{addr: addr}
}

func (s *slateConn) reconnect() error {
	conn, err := net.DialTimeout("tcp", s.addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("connecting to slate at %s: %w", s.addr, err)
	}
	s.conn = conn
	s.reader = bufio.NewReader(conn)
	return nil
}

// slateResponse is the envelope Slate wraps every reply in.
type slateResponse struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

// doSendCmd sends {"cmd":cmd,"params":params} and returns the .data field.
// Returns (nil, nil) when the response has no data field (ok-only).
// Retries once on connection error — handles the case where Slate's idle
// timeout drops a quiet connection.
func (s *slateConn) doSendCmd(cmd string, params map[string]interface{}) (json.RawMessage, error) {
	req := map[string]interface{}{"cmd": cmd, "params": params}
	line, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	line = append(line, '\n')

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if s.conn == nil {
			if err := s.reconnect(); err != nil {
				return nil, err
			}
		}

		if _, err := s.conn.Write(line); err != nil {
			s.conn.Close()
			s.conn = nil
			lastErr = fmt.Errorf("write: %w", err)
			continue
		}

		respLine, err := s.reader.ReadString('\n')
		if err != nil {
			s.conn.Close()
			s.conn = nil
			lastErr = fmt.Errorf("read: %w", err)
			continue
		}

		var resp slateResponse
		if err := json.Unmarshal([]byte(strings.TrimRight(respLine, "\r\n")), &resp); err != nil {
			return nil, fmt.Errorf("parse response: %w", err)
		}
		if !resp.OK {
			return nil, errors.New(resp.Error)
		}
		return resp.Data, nil
	}

	return nil, lastErr
}

func (s *slateConn) sendCmd(cmd string, params map[string]interface{}) (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.doSendCmd(cmd, params)
}

// ---- HTTP helpers -----------------------------------------------------------

func corsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeOK(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// passthrough sends the command and writes the raw .data JSON to the client.
func passthrough(w http.ResponseWriter, slate *slateConn, cmd string, params map[string]interface{}) {
	data, err := slate.sendCmd(cmd, params)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if data == nil {
		fmt.Fprintln(w, "{}")
		return
	}
	fmt.Fprintln(w, string(data))
}

func decode(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// pathSegments returns the URL segments after stripping /api/.
// e.g. /api/workspaces/abc123/threads  →  [workspaces, abc123, threads]
func pathSegments(r *http.Request) []string {
	p := strings.TrimPrefix(r.URL.Path, "/api/")
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// ---- Router -----------------------------------------------------------------

type server struct {
	slate *slateConn
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	segs := pathSegments(r)
	if len(segs) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch segs[0] {
	case "health":
		s.handleHealth(w, r)
	case "stats":
		s.handleStats(w, r)
	case "metrics":
		s.handleMetrics(w, r)
	case "workspaces":
		s.handleWorkspaces(w, r, segs[1:])
	case "catalogs":
		s.handleCatalogs(w, r, segs[1:])
	case "agents":
		s.handleAgents(w, r, segs[1:])
	case "agent-threads":
		s.handleAgentThreads(w, r, segs[1:])
	case "threads":
		s.handleThreads(w, r, segs[1:])
	case "pipelines":
		s.handlePipelines(w, r, segs[1:])
	case "jobs":
		s.handleJobs(w, r, segs[1:])
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// ---- Management -------------------------------------------------------------

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	_, err := s.slate.sendCmd("health", map[string]interface{}{})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	passthrough(w, s.slate, "system_stats", map[string]interface{}{})
}

func (s *server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	passthrough(w, s.slate, "system_metrics", map[string]interface{}{})
}

// ---- Workspaces  /api/workspaces/... ----------------------------------------

func (s *server) handleWorkspaces(w http.ResponseWriter, r *http.Request, segs []string) {
	// GET /api/workspaces
	// POST /api/workspaces  body: {name}
	if len(segs) == 0 {
		switch r.Method {
		case http.MethodGet:
			passthrough(w, s.slate, "ls_workspaces", map[string]interface{}{})
		case http.MethodPost:
			var body struct {
				Name string `json:"name"`
			}
			if err := decode(r, &body); err != nil || body.Name == "" {
				writeError(w, http.StatusBadRequest, "name required")
				return
			}
			_, err := s.slate.sendCmd("add_workspace", map[string]interface{}{"name": body.Name})
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeOK(w)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	wsID := segs[0]
	if len(segs) == 1 {
		// DELETE /api/workspaces/:name
		if r.Method != http.MethodDelete {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		_, err := s.slate.sendCmd("del_workspace", map[string]interface{}{"name": wsID})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)
		return
	}

	sub := segs[1]
	switch sub {
	case "catalog":
		// PUT /api/workspaces/:id/catalog  body: {catalog_id}
		if r.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			CatalogID string `json:"catalog_id"`
		}
		if err := decode(r, &body); err != nil || body.CatalogID == "" {
			writeError(w, http.StatusBadRequest, "catalog_id required")
			return
		}
		_, err := s.slate.sendCmd("set_workspace_catalog", map[string]interface{}{
			"workspace_id": wsID,
			"catalog_id":   body.CatalogID,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)

	case "router":
		// PUT /api/workspaces/:id/router  body: {agent_id}
		if r.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			AgentID string `json:"agent_id"`
		}
		if err := decode(r, &body); err != nil || body.AgentID == "" {
			writeError(w, http.StatusBadRequest, "agent_id required")
			return
		}
		_, err := s.slate.sendCmd("set_workspace_router", map[string]interface{}{
			"workspace_id": wsID,
			"agent_id":     body.AgentID,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)

	case "threads":
		// GET  /api/workspaces/:id/threads
		// POST /api/workspaces/:id/threads  body: {name?}
		switch r.Method {
		case http.MethodGet:
			passthrough(w, s.slate, "ls_threads", map[string]interface{}{"workspace_id": wsID})
		case http.MethodPost:
			var body struct {
				Name string `json:"name"`
			}
			_ = decode(r, &body) // name is optional
			passthrough(w, s.slate, "new_thread", map[string]interface{}{
				"workspace_id": wsID,
				"name":         body.Name,
			})
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}

	case "pipelines":
		// GET  /api/workspaces/:id/pipelines
		// POST /api/workspaces/:id/pipelines  body: {name}
		switch r.Method {
		case http.MethodGet:
			passthrough(w, s.slate, "ls_pipelines", map[string]interface{}{"workspace_id": wsID})
		case http.MethodPost:
			var body struct {
				Name string `json:"name"`
			}
			if err := decode(r, &body); err != nil || body.Name == "" {
				writeError(w, http.StatusBadRequest, "name required")
				return
			}
			passthrough(w, s.slate, "create_pipeline", map[string]interface{}{
				"workspace_id": wsID,
				"name":         body.Name,
			})
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}

	case "jobs":
		// GET /api/workspaces/:id/jobs
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		passthrough(w, s.slate, "ls_jobs", map[string]interface{}{"workspace_id": wsID})

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// ---- Catalogs  /api/catalogs/... --------------------------------------------

func (s *server) handleCatalogs(w http.ResponseWriter, r *http.Request, segs []string) {
	// GET /api/catalogs
	// POST /api/catalogs  body: {name}
	if len(segs) == 0 {
		switch r.Method {
		case http.MethodGet:
			passthrough(w, s.slate, "ls_catalogs", map[string]interface{}{})
		case http.MethodPost:
			var body struct {
				Name string `json:"name"`
			}
			if err := decode(r, &body); err != nil || body.Name == "" {
				writeError(w, http.StatusBadRequest, "name required")
				return
			}
			_, err := s.slate.sendCmd("add_catalog", map[string]interface{}{"name": body.Name})
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeOK(w)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	catID := segs[0]
	if len(segs) == 1 {
		// DELETE /api/catalogs/:name
		if r.Method != http.MethodDelete {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		_, err := s.slate.sendCmd("del_catalog", map[string]interface{}{"name": catID})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)
		return
	}

	sub := segs[1]
	if sub == "agents" {
		// POST /api/catalogs/:id/agents  body: {name}
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			Name string `json:"name"`
		}
		if err := decode(r, &body); err != nil || body.Name == "" {
			writeError(w, http.StatusBadRequest, "name required")
			return
		}
		passthrough(w, s.slate, "add_agent", map[string]interface{}{
			"catalog_id": catID,
			"name":       body.Name,
		})
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

// ---- Agents  /api/agents/... ------------------------------------------------

func (s *server) handleAgents(w http.ResponseWriter, r *http.Request, segs []string) {
	if len(segs) == 1 {
		if r.Method != http.MethodDelete {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		_, err := s.slate.sendCmd("del_agent", map[string]interface{}{"agent_id": segs[0]})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)
		return
	}
	if len(segs) < 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	agentID := segs[0]
	sub := segs[1]

	switch sub {
	case "instructions":
		// PUT /api/agents/:id/instructions  body: {instructions}
		if r.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			Instructions string `json:"instructions"`
		}
		if err := decode(r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}
		_, err := s.slate.sendCmd("set_agent_instructions", map[string]interface{}{
			"agent_id":     agentID,
			"instructions": body.Instructions,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)

	case "model":
		// PUT /api/agents/:id/model  body: {model}
		if r.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			Model string `json:"model"`
		}
		if err := decode(r, &body); err != nil || body.Model == "" {
			writeError(w, http.StatusBadRequest, "model required")
			return
		}
		_, err := s.slate.sendCmd("set_agent_model", map[string]interface{}{
			"agent_id": agentID,
			"model":    body.Model,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)

	case "run":
		// POST /api/agents/:id/run  body: {input}
		// Returns {"job_id":"..."} immediately — caller polls job_result.
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			Input string `json:"input"`
		}
		if err := decode(r, &body); err != nil || body.Input == "" {
			writeError(w, http.StatusBadRequest, "input required")
			return
		}
		passthrough(w, s.slate, "run_agent", map[string]interface{}{
			"agent_id": agentID,
			"input":    body.Input,
		})

	case "threads":
		// GET  /api/agents/:id/threads
		// POST /api/agents/:id/threads  body: {name?}
		switch r.Method {
		case http.MethodGet:
			passthrough(w, s.slate, "ls_agent_threads", map[string]interface{}{"agent_id": agentID})
		case http.MethodPost:
			var body struct {
				Name string `json:"name"`
			}
			_ = decode(r, &body) // name is optional
			passthrough(w, s.slate, "new_agent_thread", map[string]interface{}{
				"agent_id": agentID,
				"name":     body.Name,
			})
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}

	case "tools":
		// GET    /api/agents/:id/tools
		// POST   /api/agents/:id/tools           body: {tool}
		// DELETE /api/agents/:id/tools/:toolName
		if len(segs) == 3 {
			// DELETE /api/agents/:id/tools/:toolName
			toolName := segs[2]
			if r.Method != http.MethodDelete {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			_, err := s.slate.sendCmd("remove_tool", map[string]interface{}{
				"agent_id": agentID,
				"tool":     toolName,
			})
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeOK(w)
			return
		}
		switch r.Method {
		case http.MethodGet:
			passthrough(w, s.slate, "ls_tools", map[string]interface{}{"agent_id": agentID})
		case http.MethodPost:
			var body struct {
				Tool string `json:"tool"`
			}
			if err := decode(r, &body); err != nil || body.Tool == "" {
				writeError(w, http.StatusBadRequest, "tool required")
				return
			}
			_, err := s.slate.sendCmd("add_tool", map[string]interface{}{
				"agent_id": agentID,
				"tool":     body.Tool,
			})
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeOK(w)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// ---- Agent Threads  /api/agent-threads/... ----------------------------------

func (s *server) handleAgentThreads(w http.ResponseWriter, r *http.Request, segs []string) {
	if len(segs) < 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	threadID := segs[0]
	sub := segs[1]

	switch sub {
	case "chat":
		// POST /api/agent-threads/:id/chat  body: {message}
		// Returns {"job_id":"..."} immediately — caller polls job_result.
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			Message string `json:"message"`
		}
		if err := decode(r, &body); err != nil || body.Message == "" {
			writeError(w, http.StatusBadRequest, "message required")
			return
		}
		passthrough(w, s.slate, "agent_chat", map[string]interface{}{
			"thread_id": threadID,
			"message":   body.Message,
		})

	case "history":
		// GET /api/agent-threads/:id/history
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		passthrough(w, s.slate, "agent_thread_history", map[string]interface{}{"thread_id": threadID})

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// ---- Threads  /api/threads/... ----------------------------------------------

func (s *server) handleThreads(w http.ResponseWriter, r *http.Request, segs []string) {
	if len(segs) < 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	threadID := segs[0]
	sub := segs[1]

	switch sub {
	case "chat":
		// POST /api/threads/:id/chat  body: {message}
		// Returns {"job_id":"..."} immediately — caller polls job_result.
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			Message string `json:"message"`
		}
		if err := decode(r, &body); err != nil || body.Message == "" {
			writeError(w, http.StatusBadRequest, "message required")
			return
		}
		passthrough(w, s.slate, "chat", map[string]interface{}{
			"thread_id": threadID,
			"message":   body.Message,
		})

	case "history":
		// GET /api/threads/:id/history
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		passthrough(w, s.slate, "thread_history", map[string]interface{}{"thread_id": threadID})

	case "trace":
		// GET /api/threads/:id/trace
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		passthrough(w, s.slate, "thread_trace", map[string]interface{}{"thread_id": threadID})

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// ---- Pipelines  /api/pipelines/... ------------------------------------------

func (s *server) handlePipelines(w http.ResponseWriter, r *http.Request, segs []string) {
	if len(segs) < 2 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	pipelineID := segs[0]
	sub := segs[1]

	switch sub {
	case "steps":
		// POST /api/pipelines/:id/steps  body: {agent_id, mode}
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			AgentID string `json:"agent_id"`
			Mode    string `json:"mode"`
		}
		if err := decode(r, &body); err != nil || body.AgentID == "" || body.Mode == "" {
			writeError(w, http.StatusBadRequest, "agent_id and mode required")
			return
		}
		_, err := s.slate.sendCmd("add_pipeline_step", map[string]interface{}{
			"pipeline_id": pipelineID,
			"agent_id":    body.AgentID,
			"mode":        body.Mode,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)

	case "run":
		// POST /api/pipelines/:id/run  body: {input}
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var body struct {
			Input string `json:"input"`
		}
		if err := decode(r, &body); err != nil || body.Input == "" {
			writeError(w, http.StatusBadRequest, "input required")
			return
		}
		passthrough(w, s.slate, "run_pipeline", map[string]interface{}{
			"pipeline_id": pipelineID,
			"input":       body.Input,
		})

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// ---- Jobs  /api/jobs/... ----------------------------------------------------

func (s *server) handleJobs(w http.ResponseWriter, r *http.Request, segs []string) {
	// GET /api/jobs
	if len(segs) == 0 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		passthrough(w, s.slate, "ls_jobs", map[string]interface{}{})
		return
	}

	jobID := segs[0]
	if len(segs) == 1 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	sub := segs[1]

	switch sub {
	case "status":
		// GET /api/jobs/:id/status
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		passthrough(w, s.slate, "job_status", map[string]interface{}{"job_id": jobID})

	case "result":
		// GET /api/jobs/:id/result
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		passthrough(w, s.slate, "job_result", map[string]interface{}{"job_id": jobID})

	case "cancel":
		// POST /api/jobs/:id/cancel
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		_, err := s.slate.sendCmd("cancel_job", map[string]interface{}{"job_id": jobID})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w)

	case "wait":
		// POST /api/jobs/:id/wait
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		passthrough(w, s.slate, "wait_job", map[string]interface{}{"job_id": jobID})

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// ---- Main -------------------------------------------------------------------

func main() {
	slateAddr := os.Getenv("SLATE_ADDR")
	if slateAddr == "" {
		slateAddr = "localhost:4242"
	}
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	sc := newSlateConn(slateAddr)

	// Attempt initial connection (log warning but don't exit — slate may start later).
	if err := sc.reconnect(); err != nil {
		log.Printf("warning: could not connect to slate at %s: %v (will retry on request)", slateAddr, err)
	} else {
		log.Printf("connected to slate at %s", slateAddr)
	}

	srv := &server{slate: sc}
	mux := http.NewServeMux()
	mux.Handle("/api/", srv)

	log.Printf("bridge listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatal(err)
	}
}
