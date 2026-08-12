module github.com/easydocforms/easydocforms-mcp

go 1.25.0

require (
	github.com/easydocforms/easydocforms-go v0.1.0
	github.com/modelcontextprotocol/go-sdk v1.7.0
)

require (
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/time v0.15.0 // indirect
)

// TODO(publish): drop this replace once easydocforms-go v0.1.0 is tagged and
// on the module proxy — the launch checklist publishes the SDK first.
replace github.com/easydocforms/easydocforms-go => ../easydocforms-go
