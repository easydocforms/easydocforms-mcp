package tools

import (
	"context"
	"time"

	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createFillLinkInput struct {
	TemplateID    string `json:"template_id" jsonschema:"the template to link (from list_templates or get_import_status)"`
	ExpiresInDays int    `json:"expires_in_days,omitempty" jsonschema:"days until the link stops accepting responses; 0 or omitted = no expiry"`
	MaxResponses  int    `json:"max_responses,omitempty" jsonschema:"maximum number of submissions; 0 or omitted = unlimited"`
	ExternalRef   string `json:"external_ref,omitempty" jsonschema:"opaque correlation id (max 256 chars) echoed on submissions and webhook events — MUST NOT contain PHI (no names, birth dates, record numbers)"`
}

type createFillLinkOutput struct {
	FillLinkID string     `json:"fill_link_id"`
	URL        string     `json:"url" jsonschema:"the hosted form URL to hand to the patient"`
	ShortCode  string     `json:"short_code"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Note       string     `json:"note"`
}

func addCreateFillLink(server *mcp.Server, client *easydocforms.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "create_fill_link",
		Description: "Mint a hosted URL where a patient fills the form — no EasyDocForms account " +
			"needed on their side. Links always serve the template's latest version. Set " +
			"external_ref to your own visit/order id to correlate the eventual submission; " +
			"external_ref must never contain PHI.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createFillLinkInput) (*mcp.CallToolResult, createFillLinkOutput, error) {
		var zero createFillLinkOutput
		link, err := client.CreateFillLink(ctx, easydocforms.CreateFillLinkParams{
			TemplateID:    input.TemplateID,
			ExpiresInDays: input.ExpiresInDays,
			MaxResponses:  input.MaxResponses,
			ExternalRef:   input.ExternalRef,
		})
		if err != nil {
			return nil, zero, err
		}
		return nil, createFillLinkOutput{
			FillLinkID: link.FillLinkID,
			URL:        link.URL,
			ShortCode:  link.ShortCode,
			ExpiresAt:  link.ExpiresAt,
			Note:       "Hand this URL to the patient. When they submit, look up the result with get_submission.",
		}, nil
	})
}
