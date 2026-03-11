// Package openapi provides helpers for building OpenAPI 3.0 specification
// documents programmatically. It offers a fluent builder API for constructing
// paths, operations, schemas, and responses.
package openapi

import (
	"encoding/json"
	"net/http"
)

// Spec represents an OpenAPI 3.0 specification document.
type Spec struct {
	OpenAPI    string          `json:"openapi"`
	Info       Info            `json:"info"`
	Servers    []Server        `json:"servers,omitempty"`
	Paths      map[string]Path `json:"paths"`
	Components *Components     `json:"components,omitempty"`
}

// Info provides metadata about the API.
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// Server represents a server URL.
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Path represents operations available on a single path.
type Path map[string]*Operation

// Operation describes a single API operation on a path.
type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
	Security    []SecurityReq       `json:"security,omitempty"`
	Deprecated  bool                `json:"deprecated,omitempty"`
}

// Parameter describes a single operation parameter.
type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"` // query, header, path, cookie
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// RequestBody describes a request body.
type RequestBody struct {
	Description string             `json:"description,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Content     map[string]Content `json:"content"`
}

// Content describes the content type and schema.
type Content struct {
	Schema *Schema `json:"schema"`
}

// Response describes a single response from an API operation.
type Response struct {
	Description string             `json:"description"`
	Content     map[string]Content `json:"content,omitempty"`
}

// Schema represents a JSON Schema object.
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Enum        []string           `json:"enum,omitempty"`
	Example     any                `json:"example,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
}

// SecurityReq is a map of security scheme names to scopes.
type SecurityReq map[string][]string

// Components holds reusable schemas and security scheme definitions.
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme describes an authentication mechanism.
type SecurityScheme struct {
	Type         string `json:"type"`                   // apiKey, http, oauth2, openIdConnect
	Scheme       string `json:"scheme,omitempty"`       // e.g. "bearer"
	BearerFormat string `json:"bearerFormat,omitempty"` // e.g. "JWT"
	Name         string `json:"name,omitempty"`         // for apiKey
	In           string `json:"in,omitempty"`           // for apiKey: query, header, cookie
	Description  string `json:"description,omitempty"`
}

// NewSpec creates a new OpenAPI 3.0 specification with the given info.
func NewSpec(title, version string) *Spec {
	return &Spec{
		OpenAPI: "3.0.3",
		Info:    Info{Title: title, Version: version},
		Paths:   make(map[string]Path),
	}
}

// WithDescription sets the API description.
func (s *Spec) WithDescription(desc string) *Spec {
	s.Info.Description = desc
	return s
}

// AddServer adds a server URL to the spec.
func (s *Spec) AddServer(url, description string) *Spec {
	s.Servers = append(s.Servers, Server{URL: url, Description: description})
	return s
}

// AddPath adds an operation to the given path and method.
func (s *Spec) AddPath(path, method string, op *Operation) *Spec {
	if s.Paths[path] == nil {
		s.Paths[path] = make(Path)
	}
	s.Paths[path][method] = op
	return s
}

// AddSchema adds a reusable schema to the components section.
// Reference it in operations with RefSchema("#/components/schemas/<name>").
func (s *Spec) AddSchema(name string, schema *Schema) *Spec {
	if s.Components == nil {
		s.Components = &Components{}
	}
	if s.Components.Schemas == nil {
		s.Components.Schemas = make(map[string]*Schema)
	}
	s.Components.Schemas[name] = schema
	return s
}

// AddSecurityScheme adds a reusable security scheme to the components section.
// Reference it in operations with WithSecurity("<name>").
func (s *Spec) AddSecurityScheme(name string, scheme *SecurityScheme) *Spec {
	if s.Components == nil {
		s.Components = &Components{}
	}
	if s.Components.SecuritySchemes == nil {
		s.Components.SecuritySchemes = make(map[string]*SecurityScheme)
	}
	s.Components.SecuritySchemes[name] = scheme
	return s
}

// BearerAuth returns a SecurityScheme for HTTP Bearer token authentication.
func BearerAuth(format string) *SecurityScheme {
	return &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: format,
	}
}

// APIKeyAuth returns a SecurityScheme for API key authentication.
func APIKeyAuth(name, in string) *SecurityScheme {
	return &SecurityScheme{
		Type: "apiKey",
		Name: name,
		In:   in,
	}
}

// JSON returns the spec as a JSON byte slice.
func (s *Spec) JSON() ([]byte, error) {
	return json.Marshal(s)
}

// JSONIndent returns the spec as a pretty-printed JSON byte slice.
func (s *Spec) JSONIndent() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// Handler returns an http.HandlerFunc that serves the spec as JSON.
func (s *Spec) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := s.JSON()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// NewOperation creates a new Operation with the given summary.
func NewOperation(summary string) *Operation {
	return &Operation{
		Summary:   summary,
		Responses: make(map[string]Response),
	}
}

// WithDescription sets the operation description.
func (o *Operation) WithDescription(desc string) *Operation {
	o.Description = desc
	return o
}

// WithOperationID sets the operation ID.
func (o *Operation) WithOperationID(id string) *Operation {
	o.OperationID = id
	return o
}

// WithTags sets the operation tags.
func (o *Operation) WithTags(tags ...string) *Operation {
	o.Tags = tags
	return o
}

// WithDeprecated marks the operation as deprecated.
func (o *Operation) WithDeprecated() *Operation {
	o.Deprecated = true
	return o
}

// AddParameter adds a parameter to the operation.
func (o *Operation) AddParameter(name, in, desc string, required bool, schema *Schema) *Operation {
	o.Parameters = append(o.Parameters, Parameter{
		Name:        name,
		In:          in,
		Description: desc,
		Required:    required,
		Schema:      schema,
	})
	return o
}

// WithRequestBody sets the request body.
func (o *Operation) WithRequestBody(desc string, required bool, schema *Schema) *Operation {
	o.RequestBody = &RequestBody{
		Description: desc,
		Required:    required,
		Content: map[string]Content{
			"application/json": {Schema: schema},
		},
	}
	return o
}

// AddResponse adds a response to the operation.
func (o *Operation) AddResponse(statusCode, desc string, schema *Schema) *Operation {
	resp := Response{Description: desc}
	if schema != nil {
		resp.Content = map[string]Content{
			"application/json": {Schema: schema},
		}
	}
	o.Responses[statusCode] = resp
	return o
}

// WithSecurity adds a security requirement.
func (o *Operation) WithSecurity(scheme string, scopes ...string) *Operation {
	o.Security = append(o.Security, SecurityReq{scheme: scopes})
	return o
}

// StringSchema returns a string schema.
func StringSchema() *Schema { return &Schema{Type: "string"} }

// IntegerSchema returns an integer schema.
func IntegerSchema() *Schema { return &Schema{Type: "integer"} }

// BooleanSchema returns a boolean schema.
func BooleanSchema() *Schema { return &Schema{Type: "boolean"} }

// ArraySchema returns an array schema with the given items schema.
func ArraySchema(items *Schema) *Schema {
	return &Schema{Type: "array", Items: items}
}

// ObjectSchema returns an object schema with the given properties.
func ObjectSchema(props map[string]*Schema) *Schema {
	return &Schema{Type: "object", Properties: props}
}

// RefSchema returns a $ref schema.
func RefSchema(ref string) *Schema {
	return &Schema{Ref: ref}
}
