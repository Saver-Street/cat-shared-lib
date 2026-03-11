package openapi

import (
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
	testkit.AssertJSON(t, data, &result)
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
	testkit.AssertJSON(t, rr.Body.Bytes(), &result)
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

	testkit.RequireNotNil(t, op.RequestBody)
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
	testkit.RequireNotNil(t, s.Items)
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
	testkit.AssertJSON(t, data, &parsed)

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

// ---- Components / Schema / Security Scheme tests ----

func TestSpec_AddSchema(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	s.AddSchema("User", ObjectSchema(map[string]*Schema{
		"id":   IntegerSchema(),
		"name": StringSchema(),
	}))

	testkit.AssertNotNil(t, s.Components)
	testkit.AssertNotNil(t, s.Components.Schemas)
	testkit.AssertNotNil(t, s.Components.Schemas["User"])
	testkit.AssertEqual(t, s.Components.Schemas["User"].Type, "object")
}

func TestSpec_AddSchema_Multiple(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	s.AddSchema("User", ObjectSchema(map[string]*Schema{"id": IntegerSchema()}))
	s.AddSchema("Error", ObjectSchema(map[string]*Schema{"message": StringSchema()}))

	testkit.AssertEqual(t, len(s.Components.Schemas), 2)
	testkit.AssertNotNil(t, s.Components.Schemas["User"])
	testkit.AssertNotNil(t, s.Components.Schemas["Error"])
}

func TestSpec_AddSchema_Chainable(t *testing.T) {
	s := NewSpec("API", "1.0.0").
		AddSchema("User", StringSchema()).
		AddSchema("Error", StringSchema())
	testkit.AssertEqual(t, len(s.Components.Schemas), 2)
}

func TestSpec_AddSecurityScheme(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	s.AddSecurityScheme("bearerAuth", BearerAuth("JWT"))

	testkit.AssertNotNil(t, s.Components)
	testkit.AssertNotNil(t, s.Components.SecuritySchemes)
	scheme := s.Components.SecuritySchemes["bearerAuth"]
	testkit.AssertNotNil(t, scheme)
	testkit.AssertEqual(t, scheme.Type, "http")
	testkit.AssertEqual(t, scheme.Scheme, "bearer")
	testkit.AssertEqual(t, scheme.BearerFormat, "JWT")
}

func TestSpec_AddSecurityScheme_APIKey(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	s.AddSecurityScheme("apiKey", APIKeyAuth("X-API-Key", "header"))

	scheme := s.Components.SecuritySchemes["apiKey"]
	testkit.AssertNotNil(t, scheme)
	testkit.AssertEqual(t, scheme.Type, "apiKey")
	testkit.AssertEqual(t, scheme.Name, "X-API-Key")
	testkit.AssertEqual(t, scheme.In, "header")
}

func TestSpec_AddSecurityScheme_Chainable(t *testing.T) {
	s := NewSpec("API", "1.0.0").
		AddSecurityScheme("bearer", BearerAuth("JWT")).
		AddSecurityScheme("apiKey", APIKeyAuth("X-Key", "header"))
	testkit.AssertEqual(t, len(s.Components.SecuritySchemes), 2)
}

func TestSpec_Components_JSON(t *testing.T) {
	s := NewSpec("API", "1.0.0").
		AddSchema("User", ObjectSchema(map[string]*Schema{
			"id":   IntegerSchema(),
			"name": StringSchema(),
		})).
		AddSecurityScheme("bearerAuth", BearerAuth("JWT")).
		AddPath("/users", "get", NewOperation("List users").
			AddResponse("200", "Success", RefSchema("#/components/schemas/User")).
			WithSecurity("bearerAuth"))

	data, err := s.JSON()
	testkit.AssertNoError(t, err)

	var parsed map[string]any
	testkit.AssertJSON(t, data, &parsed)

	components, ok := parsed["components"].(map[string]any)
	testkit.AssertTrue(t, ok)
	testkit.AssertNotNil(t, components["schemas"])
	testkit.AssertNotNil(t, components["securitySchemes"])
}

func TestSpec_NilComponents_OmittedFromJSON(t *testing.T) {
	s := NewSpec("API", "1.0.0")
	data, err := s.JSON()
	testkit.AssertNoError(t, err)
	testkit.AssertNotContains(t, string(data), "components")
}

func TestBearerAuth(t *testing.T) {
	scheme := BearerAuth("JWT")
	testkit.AssertEqual(t, scheme.Type, "http")
	testkit.AssertEqual(t, scheme.Scheme, "bearer")
	testkit.AssertEqual(t, scheme.BearerFormat, "JWT")
}

func TestAPIKeyAuth(t *testing.T) {
	scheme := APIKeyAuth("X-API-Key", "header")
	testkit.AssertEqual(t, scheme.Type, "apiKey")
	testkit.AssertEqual(t, scheme.Name, "X-API-Key")
	testkit.AssertEqual(t, scheme.In, "header")
}

func TestBearerAuth_EmptyFormat(t *testing.T) {
	scheme := BearerAuth("")
	testkit.AssertEqual(t, scheme.Type, "http")
	testkit.AssertEqual(t, scheme.Scheme, "bearer")
	testkit.AssertEmpty(t, scheme.BearerFormat)
}
