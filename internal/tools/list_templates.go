package tools

import (
	"context"
	"time"

	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listTemplatesInput struct{}

type templateSummary struct {
	TemplateID string    `json:"template_id"`
	Title      string    `json:"title"`
	PageCount  int       `json:"page_count"`
	FieldCount int       `json:"field_count"`
	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
}

type listTemplatesOutput struct {
	Templates []templateSummary `json:"templates"`
}

func addListTemplates(server *mcp.Server, client *easydocforms.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_templates",
		Description: "List the organization's active PDF form templates, newest first — templates " +
			"imported via this server and templates created in the EasyDocForms app alike. Use a " +
			"template_id with create_fill_link.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listTemplatesInput) (*mcp.CallToolResult, listTemplatesOutput, error) {
		var zero listTemplatesOutput
		templates, err := client.ListTemplates(ctx)
		if err != nil {
			return nil, zero, err
		}
		out := listTemplatesOutput{Templates: make([]templateSummary, 0, len(templates))}
		for _, t := range templates {
			out.Templates = append(out.Templates, templateSummary{
				TemplateID: t.TemplateID,
				Title:      t.Title,
				PageCount:  t.PageCount,
				FieldCount: t.FieldCount,
				Version:    t.Version,
				CreatedAt:  t.CreatedAt,
			})
		}
		return nil, out, nil
	})
}
