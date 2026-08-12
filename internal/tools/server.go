// Package tools registers the EasyDocForms agent tools on an MCP server.
//
// PHI stance, enforced by construction: tools move template ids, hosted
// links, and status — never patient answers. get_submission returns metadata
// and an answer count; the answers themselves are only retrievable
// server-side, over the authenticated Partner API. There is deliberately no
// tool that uploads document bytes: imports take a URL, and blank templates
// only.
package tools

import (
	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer returns an MCP server with all six EasyDocForms tools registered.
func NewServer(client *easydocforms.Client, version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "easydocforms",
		Title:   "EasyDocForms — PDF intake forms",
		Version: version,
	}, nil)
	addImportPDFFromURL(server, client)
	addGetImportStatus(server, client)
	addListTemplates(server, client)
	addCreateFillLink(server, client)
	addGetSubmission(server, client)
	addGetCompletedPDFLink(server, client)
	return server
}
