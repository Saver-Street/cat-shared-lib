package openapi

import "testing"

func BenchmarkNewSpec(b *testing.B) {
	for b.Loop() {
		NewSpec("Pet Store API", "1.0.0")
	}
}

func BenchmarkSpecJSON(b *testing.B) {
	s := NewSpec("Pet Store API", "1.0.0").
		WithDescription("A sample API for pets").
		AddServer("https://api.example.com", "Production")

	op := NewOperation("List pets").
		WithDescription("Returns all pets").
		WithOperationID("listPets").
		WithTags("pets").
		AddParameter("limit", "query", "Max items", false,
			&Schema{Type: "integer", Format: "int32"}).
		AddResponse("200", "Success", &Schema{Type: "array",
			Items: &Schema{Type: "object"}})
	s.AddPath("/pets", "get", op)

	b.ResetTimer()
	for b.Loop() {
		s.JSON()
	}
}

func BenchmarkSpecJSONIndent(b *testing.B) {
	s := NewSpec("API", "1.0").
		WithDescription("Test").
		AddServer("https://api.example.com", "Prod")
	b.ResetTimer()
	for b.Loop() {
		s.JSONIndent()
	}
}

func BenchmarkNewOperation(b *testing.B) {
	for b.Loop() {
		NewOperation("Create pet").
			WithDescription("Creates a new pet").
			WithOperationID("createPet").
			WithTags("pets").
			WithRequestBody("Pet to create", true,
				&Schema{Type: "object"}).
			AddResponse("201", "Created", &Schema{Type: "object"})
	}
}

func BenchmarkAddPath(b *testing.B) {
	op := NewOperation("Get").WithOperationID("get")
	for b.Loop() {
		s := NewSpec("API", "1.0")
		s.AddPath("/items", "get", op)
		s.AddPath("/items", "post", op)
		s.AddPath("/items/{id}", "get", op)
		s.AddPath("/items/{id}", "put", op)
		s.AddPath("/items/{id}", "delete", op)
	}
}
