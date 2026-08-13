# easydocforms-mcp

[![CI](https://github.com/easydocforms/easydocforms-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/easydocforms/easydocforms-mcp/actions/workflows/ci.yml)

An MCP server for **healthcare intake forms**: give an AI agent the ability to turn a blank PDF intake packet into a hosted, mobile-friendly fillable form, hand the patient a link, and retrieve the completed, pixel-exact PDF — without any PHI ever entering the agent's context.

Built on [EasyDocForms](https://easydocforms.com), whose document-understanding pipeline (field detection, checkbox semantics, consent blocks, signature models) runs healthcare intake in production. This server wraps the [Partner API](https://easydocforms.com/docs/api) via the [easydocforms-go](https://github.com/easydocforms/easydocforms-go) SDK.

## The PHI boundary (read this first)

This server is designed so protected health information never passes through the model:

- **Imports are blank forms only.** `import_pdf_from_url` requires `blank_form_attestation: true` and takes a URL, never document bytes.
- **`get_submission` returns metadata and an answer count — never the patient's answers.** Answers stay server-side, behind your API key.
- **`get_completed_pdf_link` returns a ~10-minute signed URL** the agent can hand to a human or an EMR without exposing your API key. Treat it like a fax.
- **`external_ref` must never contain PHI** — it's an opaque correlation id (your visit or order number).

## Tools

| Tool | What it does |
|---|---|
| `import_pdf_from_url` | Import a blank PDF intake form (async; requires blank-form attestation) |
| `get_import_status` | Poll an import until its template is ready |
| `list_templates` | List the organization's active form templates |
| `create_fill_link` | Mint a hosted URL where a patient fills the form |
| `get_submission` | Submission metadata + answer count — never answers |
| `get_completed_pdf_link` | Short-lived signed download URL for the completed PDF |

## Install

```sh
go install github.com/easydocforms/easydocforms-mcp/cmd/easydocforms-mcp@latest
```

You'll need an API key from the EasyDocForms app: **Settings → Integrations → Partner API** (shown exactly once).

Or run the Docker image (no Go toolchain needed):

```sh
docker run -i --rm -e EASYDOCFORMS_API_KEY=edfk_live_... ghcr.io/easydocforms/easydocforms-mcp:latest
```

### Claude Code

```sh
claude mcp add easydocforms --env EASYDOCFORMS_API_KEY=edfk_live_... -- easydocforms-mcp
```

### Claude Desktop (or any MCP client)

```json
{
  "mcpServers": {
    "easydocforms": {
      "command": "easydocforms-mcp",
      "env": { "EASYDOCFORMS_API_KEY": "edfk_live_..." }
    }
  }
}
```

### Environment

| Variable | Required | Purpose |
|---|---|---|
| `EASYDOCFORMS_API_KEY` | yes | `edfk_live_*` Partner API key |
| `EASYDOCFORMS_BASE_URL` | no | API base URL override |

## Try it

With [MCP Inspector](https://github.com/modelcontextprotocol/inspector):

```sh
EASYDOCFORMS_API_KEY=edfk_live_... npx @modelcontextprotocol/inspector easydocforms-mcp
```

A typical agent session: *"Import the intake form at https://clinic.example.com/forms/new-patient.pdf (it's blank), then give me a fill link for visit 8675309."* The agent imports, polls, mints a link with `external_ref: "visit-8675309"`, and hands back a URL. After the patient submits, *"Is the PDF for that visit ready?"* → `get_submission` + `get_completed_pdf_link`.

## Transport

stdio only, credentials from the environment — per the MCP spec's guidance for locally-run servers. A hosted remote variant (Streamable HTTP + OAuth 2.1) is planned as a second wave.

## Development

```sh
go test ./... -race
```

## License

MIT
