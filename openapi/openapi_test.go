package openapi

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSpec(t *testing.T) {
	s := NewSpec("TestAPI", "1.0.0")
	if s.OpenAPI != "3.0.3" {
		t.Errorf("expected openapi 3.0.3, got %q", s.OpenAPI)
	}
	if s.Info.Title != "TestAPI" {
		t.Errorf("expected title TestAPI, got %q", s.Info.Title)
	}
	if s.Info.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", s.Info.Version)
	}
	if s.Paths == nil {
		t.Error("expected non-nil paths")
	}
}

func TestSpec_WithDescription(t *testing.T) {
	s := NewSpec("API", "1.0.0").WithDescription("A test API")
	if s.Info.Description != "A test API" {
		t.Errorf("expected description, got %q", s.Info.Description)
	}
}

func TestSpec_AddServer(t *testing.T) {
	s := NewSpec("API", "1.0.0").AddServer("https://api.example.com", "Production")
	if len(s.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(s.Servers))
	}
	if s.Servers[0].URL != "https://api.example.com" {
		t.Errorf("expected url, got %q", s.Servers[0].URL)
	}
}

func TestSpec_AddPath(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	op := NewOperation("List users")
	s.AddPath("/users", "get", op)

	if _, ok := s.Paths["/users"]; !ok {
		t.Fatal("expected /users path")
	}
	if s.Paths["/users"]["get"].Summary != "List users" {
		t.Error("expected operation summary")
	}
}

func TestSpec_AddPath_MultipleMethods(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	s.AddPath("/users", "get", NewOperation("List"))
	s.AddPath("/users", "post", NewOperation("Create"))

	if len(s.Paths["/users"]) != 2 {
		t.Errorf("expected 2 methods, got %d", len(s.Paths["/users"]))
	}
}

func TestSpec_JSON(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	data, err := s.JSON()
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["openapi"] != "3.0.3" {
		t.Error("expected openapi field")
	}
}

func TestSpec_JSONIndent(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	data, err := s.JSONIndent()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty output")
	}
	// Should contain indentation.
	if string(data[0:1]) != "{" {
		t.Error("expected JSON object")
	}
}

func TestSpec_Handler(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	s.AddPath("/health", "get", NewOperation("Health check").AddResponse("200", "OK", nil))

	handler := s.Handler()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	var result Spec
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
}

func TestSpec_Handler_MarshalError(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	// math.Inf cannot be marshaled to JSON, triggering the error path.
	schema := &Schema{Example: math.Inf(1)}
	s.AddPath("/bad", "get", NewOperation("bad").
		AddResponse("200", "ok", schema))

	handler := s.Handler()
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for marshal error, got %d", rr.Code)
	}
}

func TestNewOperation(t *testing.T) {
	op := NewOperation("Test op")
	if op.Summary != "Test op" {
		t.Errorf("expected summary, got %q", op.Summary)
	}
	if op.Responses == nil {
		t.Error("expected non-nil responses")
	}
}

func TestOperation_WithDescription(t *testing.T) {
	op := NewOperation("Test").WithDescription("Detailed desc")
	if op.Description != "Detailed desc" {
		t.Error("expected description")
	}
}

func TestOperation_WithOperationID(t *testing.T) {
	op := NewOperation("Test").WithOperationID("listUsers")
	if op.OperationID != "listUsers" {
		t.Error("expected operationId")
	}
}

func TestOperation_WithTags(t *testing.T) {
	op := NewOperation("Test").WithTags("users", "admin")
	if len(op.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(op.Tags))
	}
}

func TestOperation_WithDeprecated(t *testing.T) {
	op := NewOperation("Test").WithDeprecated()
	if !op.Deprecated {
		t.Error("expected deprecated")
	}
}

func TestOperation_AddParameter(t *testing.T) {
	op := NewOperation("Test").
		AddParameter("page", "query", "Page number", false, IntegerSchema()).
		AddParameter("id", "path", "User ID", true, StringSchema())

	if len(op.Parameters) != 2 {
		t.Fatalf("expected 2 params, got %d", len(op.Parameters))
	}
	if op.Parameters[0].Name != "page" || op.Parameters[0].In != "query" {
		t.Error("first param mismatch")
	}
	if !op.Parameters[1].Required {
		t.Error("expected id to be required")
	}
}

func TestOperation_WithRequestBody(t *testing.T) {
	schema := ObjectSchema(map[string]*Schema{
		"name":  StringSchema(),
		"email": StringSchema(),
	})
	op := NewOperation("Create user").WithRequestBody("User data", true, schema)

	if op.RequestBody == nil {
		t.Fatal("expected request body")
	}
	if !op.RequestBody.Required {
		t.Error("expected required body")
	}
	if _, ok := op.RequestBody.Content["application/json"]; !ok {
		t.Error("expected application/json content")
	}
}

func TestOperation_AddResponse(t *testing.T) {
	op := NewOperation("Test").
		AddResponse("200", "Success", StringSchema()).
		AddResponse("404", "Not found", nil)

	if len(op.Responses) != 2 {
		t.Errorf("expected 2 responses, got %d", len(op.Responses))
	}
	if op.Responses["200"].Content == nil {
		t.Error("expected content for 200")
	}
	if op.Responses["404"].Content != nil {
		t.Error("expected no content for 404")
	}
}

func TestOperation_WithSecurity(t *testing.T) {
	op := NewOperation("Test").
		WithSecurity("bearerAuth").
		WithSecurity("oauth2", "read:users", "write:users")

	if len(op.Security) != 2 {
		t.Errorf("expected 2 security reqs, got %d", len(op.Security))
	}
}

func TestSchemaHelpers(t *testing.T) {
	tests := []struct {
		name   string
		schema *Schema
		want   string
	}{
		{"string", StringSchema(), "string"},
		{"integer", IntegerSchema(), "integer"},
		{"boolean", BooleanSchema(), "boolean"},
		{"array", ArraySchema(StringSchema()), "array"},
		{"object", ObjectSchema(nil), "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.schema.Type != tt.want {
				t.Errorf("expected type %q, got %q", tt.want, tt.schema.Type)
			}
		})
	}
}

func TestArraySchema_Items(t *testing.T) {
	s := ArraySchema(IntegerSchema())
	if s.Items == nil || s.Items.Type != "integer" {
		t.Error("expected integer items")
	}
}

func TestRefSchema(t *testing.T) {
	s := RefSchema("#/components/schemas/User")
	if s.Ref != "#/components/schemas/User" {
		t.Errorf("expected ref, got %q", s.Ref)
	}
}

func TestFullSpec_Serialization(t *testing.T) {
	s := NewSpec("Pet Store", "1.0.0").
		WithDescription("A sample pet store").
		AddServer("https://petstore.example.com/v1", "Production")

	listOp := NewOperation("List pets").
		WithOperationID("listPets").
		WithTags("pets").
		AddParameter("limit", "query", "Max items", false, IntegerSchema()).
		AddResponse("200", "A list of pets", ArraySchema(RefSchema("#/components/schemas/Pet")))

	createOp := NewOperation("Create a pet").
		WithOperationID("createPet").
		WithTags("pets").
		WithRequestBody("Pet to add", true, ObjectSchema(map[string]*Schema{
			"name": StringSchema(),
		})).
		AddResponse("201", "Created", nil).
		WithSecurity("bearerAuth")

	s.AddPath("/pets", "get", listOp)
	s.AddPath("/pets", "post", createOp)

	data, err := s.JSONIndent()
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["openapi"] != "3.0.3" {
		t.Error("missing openapi version")
	}
	paths, ok := parsed["paths"].(map[string]any)
	if !ok {
		t.Fatal("missing paths")
	}
	if _, ok := paths["/pets"]; !ok {
		t.Error("missing /pets path")
	}
}

func BenchmarkSpec_JSON(b *testing.B) {
	s := NewSpec("API", "1.0.0")
	for i := 0; i < 50; i++ {
		s.AddPath("/path"+string(rune(i)), "get", NewOperation("Op").AddResponse("200", "OK", nil))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.JSON()
	}
}
