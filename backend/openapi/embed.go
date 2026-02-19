package openapi

import _ "embed"

// OpenAPIYAML is the raw OpenAPI 3.0 specification (YAML).
// Served at /api/openapi.yaml and /api/openapi.json.
//go:embed openapi.yaml
var OpenAPIYAML []byte
