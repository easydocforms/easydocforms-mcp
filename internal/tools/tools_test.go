package tools_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/easydocforms/easydocforms-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakePartnerAPI serves canned Partner API responses. importPosts counts how
// many import requests actually reached the API.
func fakePartnerAPI(t *testing.T, importPosts *atomic.Int32) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /imports", func(w http.ResponseWriter, r *http.Request) {
		importPosts.Add(1)
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]any{"import_id": "imp_1", "status": "queued"})
	})
	mux.HandleFunc("GET /imports/imp_1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"import_id": "imp_1", "status": "succeeded", "filename": "intake.pdf",
			"created_at": "2026-08-12T10:00:00Z", "updated_at": "2026-08-12T10:03:00Z",
			"template_id": "tpl_1", "page_count": 1, "field_count": 4,
			"review_required": true, "review_reasons": []string{"check page 1"}, "warnings": []string{},
		})
	})
	mux.HandleFunc("GET /templates", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"templates": []map[string]any{{
			"template_id": "tpl_1", "title": "Intake", "source_filename": "intake.pdf",
			"page_count": 1, "field_count": 4, "detector": "azure", "version": 1,
			"created_at": "2026-08-12T10:00:00Z", "updated_at": "2026-08-12T10:00:00Z",
		}}})
	})
	mux.HandleFunc("POST /fill-links", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"fill_link_id": "fl_1", "template_id": "tpl_1", "short_code": "Ab12Cd",
			"url": "https://form.easydocforms.com/f/Ab12Cd", "created_at": "2026-08-12T10:05:00Z",
			"external_ref": "visit-42",
		})
	})
	// The submission carries PHI-shaped answers on purpose: the tests assert
	// none of it survives the MCP boundary.
	mux.HandleFunc("GET /submissions/sub_1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"submission_id": "sub_1", "submitted_at": "2026-08-12T11:00:00Z",
			"answers": map[string]any{
				"patient_name": "John Smith",
				"dob":          "1980-01-02",
				"ssn":          "123-45-6789",
			},
			"template_id": "tpl_1", "completed_pdf_status": "ready",
			"fill_link_id": "fl_1", "external_ref": "visit-42",
		})
	})
	mux.HandleFunc("GET /submissions/sub_1/pdf-link", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"url": "https://storage.googleapis.com/signed?sig=abc", "expires_at": "2026-08-12T11:10:00Z",
		})
	})
	mux.HandleFunc("GET /submissions/pending_1/pdf-link", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "completed PDF not frozen yet", "completed_pdf_status": "pending",
			"hint": "stream GET /submissions/{id}/pdf instead",
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// newSession wires the tools server to an in-memory MCP client session.
func newSession(t *testing.T, importPosts *atomic.Int32) *mcp.ClientSession {
	t.Helper()
	api := fakePartnerAPI(t, importPosts)
	client := easydocforms.NewClient("edfk_live_test", easydocforms.WithBaseURL(api.URL))
	server := tools.NewServer(client, "test")

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	return result
}

func resultJSON(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestToolRoster(t *testing.T) {
	var importPosts atomic.Int32
	session := newSession(t, &importPosts)
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"import_pdf_from_url":    false,
		"get_import_status":      false,
		"list_templates":         false,
		"create_fill_link":       false,
		"get_submission":         false,
		"get_completed_pdf_link": false,
	}
	for _, tool := range listed.Tools {
		if _, known := want[tool.Name]; !known {
			t.Errorf("unexpected tool %q", tool.Name)
		}
		want[tool.Name] = true
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("tool %q not registered", name)
		}
	}
}

// TestToolTitlesAndAnnotations is the directory-submission tripwire: every
// registered tool must carry a Title and explicit annotations (the Anthropic
// connectors portal auto-flags tools without them), write tools must opt out
// of the spec's destructive-by-default (*bool defaulting TRUE in go-sdk),
// and OpenWorldHint (same default-TRUE trap) must always be explicit.
func TestToolTitlesAndAnnotations(t *testing.T) {
	var importPosts atomic.Int32
	session := newSession(t, &importPosts)
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	readOnly := map[string]bool{
		"list_templates":         true,
		"get_import_status":      true,
		"get_submission":         true,
		"get_completed_pdf_link": true,
		"import_pdf_from_url":    false,
		"create_fill_link":       false,
	}
	for _, tool := range listed.Tools {
		if tool.Title == "" {
			t.Errorf("%s: missing Title", tool.Name)
		}
		a := tool.Annotations
		if a == nil {
			t.Errorf("%s: missing Annotations", tool.Name)
			continue
		}
		if a.OpenWorldHint == nil {
			t.Errorf("%s: OpenWorldHint must be set explicitly (spec default is true)", tool.Name)
		}
		wantRO, known := readOnly[tool.Name]
		if !known {
			continue // TestToolRoster reports unexpected tools
		}
		if a.ReadOnlyHint != wantRO {
			t.Errorf("%s: ReadOnlyHint = %v, want %v", tool.Name, a.ReadOnlyHint, wantRO)
		}
		if !wantRO && (a.DestructiveHint == nil || *a.DestructiveHint) {
			t.Errorf("%s: write tool must set DestructiveHint explicitly false (spec default is true)", tool.Name)
		}
	}
}

// TestGetSubmissionNeverLeaksAnswers is the PHI boundary test: the fake API
// returns a submission full of PHI-shaped answers, and NOTHING of it may
// appear anywhere in the serialized MCP result.
func TestGetSubmissionNeverLeaksAnswers(t *testing.T) {
	var importPosts atomic.Int32
	session := newSession(t, &importPosts)
	result := callTool(t, session, "get_submission", map[string]any{"submission_id": "sub_1"})
	if result.IsError {
		t.Fatalf("get_submission errored: %s", resultJSON(t, result))
	}
	raw := resultJSON(t, result)
	for _, leaked := range []string{"John Smith", "1980-01-02", "123-45-6789", "patient_name", `"answers"`} {
		if strings.Contains(raw, leaked) {
			t.Errorf("PHI boundary violated: %q appears in tool result:\n%s", leaked, raw)
		}
	}
	structured, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content = %T", result.StructuredContent)
	}
	if structured["answer_count"] != float64(3) {
		t.Errorf("answer_count = %v, want 3", structured["answer_count"])
	}
	if structured["external_ref"] != "visit-42" || structured["completed_pdf_status"] != "ready" {
		t.Errorf("metadata = %v", structured)
	}
}

func TestImportRequiresAttestation(t *testing.T) {
	var importPosts atomic.Int32
	session := newSession(t, &importPosts)
	result := callTool(t, session, "import_pdf_from_url", map[string]any{
		"pdf_url":                "https://example.com/intake.pdf",
		"filename":               "intake.pdf",
		"blank_form_attestation": false,
	})
	if !result.IsError {
		t.Fatal("expected a tool error without attestation")
	}
	if raw := resultJSON(t, result); !strings.Contains(raw, "blank_form_attestation") {
		t.Errorf("error should explain the attestation: %s", raw)
	}
	if importPosts.Load() != 0 {
		t.Error("the API was called despite missing attestation")
	}
}

func TestImportAndStatusFlow(t *testing.T) {
	var importPosts atomic.Int32
	session := newSession(t, &importPosts)

	created := callTool(t, session, "import_pdf_from_url", map[string]any{
		"pdf_url":                "https://example.com/intake.pdf",
		"filename":               "intake.pdf",
		"blank_form_attestation": true,
	})
	if created.IsError {
		t.Fatalf("import errored: %s", resultJSON(t, created))
	}
	if importPosts.Load() != 1 {
		t.Fatalf("import posts = %d, want 1", importPosts.Load())
	}

	status := callTool(t, session, "get_import_status", map[string]any{"import_id": "imp_1"})
	structured := status.StructuredContent.(map[string]any)
	if structured["status"] != "succeeded" || structured["template_id"] != "tpl_1" {
		t.Errorf("status = %v", structured)
	}
	if structured["review_required"] != true {
		t.Errorf("review_required = %v", structured["review_required"])
	}
}

func TestCreateFillLink(t *testing.T) {
	var importPosts atomic.Int32
	session := newSession(t, &importPosts)
	result := callTool(t, session, "create_fill_link", map[string]any{
		"template_id":  "tpl_1",
		"external_ref": "visit-42",
	})
	structured := result.StructuredContent.(map[string]any)
	if structured["url"] != "https://form.easydocforms.com/f/Ab12Cd" {
		t.Errorf("url = %v", structured["url"])
	}
}

func TestCompletedPDFLinkReadyAndPending(t *testing.T) {
	var importPosts atomic.Int32
	session := newSession(t, &importPosts)

	ready := callTool(t, session, "get_completed_pdf_link", map[string]any{"submission_id": "sub_1"})
	if ready.IsError {
		t.Fatalf("ready case errored: %s", resultJSON(t, ready))
	}
	if structured := ready.StructuredContent.(map[string]any); structured["url"] == "" {
		t.Errorf("ready = %v", structured)
	}

	pending := callTool(t, session, "get_completed_pdf_link", map[string]any{"submission_id": "pending_1"})
	if pending.IsError {
		t.Fatalf("pending must be a graceful result, not a tool error: %s", resultJSON(t, pending))
	}
	structured := pending.StructuredContent.(map[string]any)
	if structured["pending"] != true {
		t.Errorf("pending = %v", structured)
	}
	if _, hasURL := structured["url"]; hasURL {
		t.Errorf("pending result should omit url: %v", structured)
	}
}

func TestListTemplates(t *testing.T) {
	var importPosts atomic.Int32
	session := newSession(t, &importPosts)
	result := callTool(t, session, "list_templates", map[string]any{})
	structured := result.StructuredContent.(map[string]any)
	templates := structured["templates"].([]any)
	if len(templates) != 1 {
		t.Fatalf("templates = %v", templates)
	}
	if templates[0].(map[string]any)["template_id"] != "tpl_1" {
		t.Errorf("template = %v", templates[0])
	}
}
