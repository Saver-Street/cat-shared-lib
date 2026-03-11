package openapi

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNewSpec(t *testing.T) {
	s := NewSpec("TestAPI", "1.0.0")
	testkit.AssertEqual(t, s.OpenAPI, "3.0.3")
	testkit.AssertEqual(t, s.Info.Title, "TestAPI")
	testkit.AssertEqual(t, s.Info.Version, "1.0.0")
	testkit.AssertNotNil(t, s.Paths)
}

func TestSpec_WithDescription(t *testing.T) {
	s := NewSpec("API", "1.0.0").WithDescription("A test API")
	testkit.AssertEqual(t, s.Info.Description, "A test API")
}

func TestSpec_AddServer(t *testing.T) {
	s := NewSpec("API", "1.0.0").AddServer("https://api.example.com", "Production")
	testkit.RequireLen(t, s.Servers, 1)
	testkit.AssertEqual(t, s.Servers[0].URL, "https://api.example.com")
}

func TestSpec_AddPath(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	op := NewOperation("List users")
	s.AddPath("/users", "get", op)

	if _, ok := s.Paths["/users"]; !ok {
		t.Fatal("expected /users path")
	}
	testkit.AssertEqual(t, s.Paths["/users"]["get"].Summary, "List users")
}

func TestSpec_AddPath_MultipleMethods(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	s.AddPath("/users", "get", NewOperation("List"))
	s.AddPath("/users", "post", NewOperation("Create"))

	testkit.AssertLen(t, s.Paths["/users"], 2)
}

func TestSpec_JSON(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	data, err := s.JSON()
	testkit.RequireNoError(t, err)

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	testkit.AssertEqual(t, result["openapi"], "3.0.3")
}

func TestSpec_JSONIndent(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	data, err := s.JSONIndent()
	testkit.RequireNoError(t, err)
	testkit.AssertTrue(t, len(data) > 0)
	// Should contain indentation.
	testkit.AssertEqual(t, string(data[0:1]), "{")
}

func TestSpec_Handler(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	s.AddPath("/health", "get", NewOperation("Health check").AddResponse("200", "OK", nil))

	handler := s.Handler()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertEqual(t, rr.Code, http.StatusOK)
	testkit.AssertEqual(t, rr.Header().Get("Content-Type"), "application/json")

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

	testkit.AssertEqual(t, rr.Code, http.StatusInternalServerError)
}

func TestNewOperation(t *testing.T) {
	op := NewOperation("Test op")
	testkit.AssertEqual(t, op.Summary, "Test op")
	testkit.AssertNotNil(t, op.Responses)
}

func TestOperation_WithDescription(t *testing.T) {
	op := NewOperation("Test").WithDescription("Detailed desc")
	testkit.AssertEqual(t, op.Description, "Detailed desc")
}

func TestOperation_WithOperationID(t *testing.T) {
	op := NewOperation("Test").WithOperationID("listUsers")
	testkit.AssertEqual(t, op.OperationID, "listUsers")
}

func TestOperation_WithTags(t *testing.T) {
	op := NewOperation("Test").WithTags("users", "admin")
	testkit.AssertLen(t, op.Tags, 2)
}

func TestOperation_WithDeprecated(t *testing.T) {
	op := NewOperation("Test").WithDeprecated()
	testkit.AssertTrue(t, op.Deprecated)
}

func TestOperation_AddParameter(t *testing.T) {
	op := NewOperation("Test").
		AddParameter("page", "query", "Page number", false, IntegerSchema()).
		AddParameter("id", "path", "User ID", true, StringSchema())

	testkit.RequireLen(t, op.Parameters, 2)
	testkit.AssertEqual(t, op.Parameters[0].Name, "page")
	testkit.AssertEqual(t, op.Parameters[0].In, "query")
	testkit.AssertTrue(t, op.Parameters[1].Required)
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
	testkit.AssertTrue(t, op.RequestBody.Required)
	_, ok := op.RequestBody.Content["application/json"]
	testkit.AssertTrue(t, ok)
}

func TestOperation_AddResponse(t *testing.T) {
	op := NewOperation("Test").
		AddResponse("200", "Success", StringSchema()).
		AddResponse("404", "Not found", nil)

	testkit.AssertLen(t, op.Responses, 2)
	testkit.AssertNotNil(t, op.Responses["200"].Content)
	testkit.AssertNil(t, op.Responses["404"].Content)
}

func TestOperation_WithSecurity(t *testing.T) {
	op := NewOperation("Test").
		WithSecurity("bearerAuth").
		WithSecurity("oauth2", "read:users", "write:users")

	testkit.AssertLen(t, op.Security, 2)
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
			testkit.AssertEqual(t, tt.schema.Type, tt.want)
		})
	}
}

func TestArraySchema_Items(t *testing.T) {
	s := ArraySchema(IntegerSchema())
	if s.Items == nil {
		t.Fatal("expected non-nil items")
	}
	testkit.AssertEqual(t, s.Items.Type, "integer")
}

func TestRefSchema(t *testing.T) {
	s := RefSchema("#/components/schemas/User")
	testkit.AssertEqual(t, s.Ref, "#/components/schemas/User")
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
	testkit.RequireNoError(t, err)

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	testkit.AssertEqual(t, parsed["openapi"], "3.0.3")
	paths, ok := parsed["paths"].(map[string]any)
	if !ok {
		t.Fatal("missing paths")
	}
	_, ok2 := paths["/pets"]
	testkit.AssertTrue(t, ok2)
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
