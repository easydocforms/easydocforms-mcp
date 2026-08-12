package tools

import (
	"context"

	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getImportStatusInput struct {
	ImportID string `json:"import_id" jsonschema:"the import_id returned by import_pdf_from_url"`
}

type getImportStatusOutput struct {
	ImportID string `json:"import_id"`
	Status   string `json:"status" jsonschema:"queued | processing | succeeded | failed"`

	TemplateID     string   `json:"template_id,omitempty" jsonschema:"present when succeeded — use with create_fill_link"`
	PageCount      int      `json:"page_count,omitempty"`
	FieldCount     int      `json:"field_count,omitempty"`
	ReviewRequired bool     `json:"review_required,omitempty" jsonschema:"true when a human should double-check the template in the EasyDocForms editor before sending it to patients"`
	ReviewReasons  []string `json:"review_reasons,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	Error          string   `json:"error,omitempty" jsonschema:"present when failed"`

	Note string `json:"note,omitempty"`
}

func addGetImportStatus(server *mcp.Server, client *easydocforms.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_import_status",
		Description: "Check on a PDF import started by import_pdf_from_url. Imports never fail for " +
			"quality reasons: the template is always created, and review_required tells staff what " +
			"to double-check in the EasyDocForms editor.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getImportStatusInput) (*mcp.CallToolResult, getImportStatusOutput, error) {
		var zero getImportStatusOutput
		imp, err := client.GetImport(ctx, input.ImportID)
		if err != nil {
			return nil, zero, err
		}
		out := getImportStatusOutput{
			ImportID:       imp.ImportID,
			Status:         string(imp.Status),
			TemplateID:     imp.TemplateID,
			PageCount:      imp.PageCount,
			FieldCount:     imp.FieldCount,
			ReviewRequired: imp.ReviewRequired,
			ReviewReasons:  imp.ReviewReasons,
			Warnings:       imp.Warnings,
			Error:          imp.ErrorMessage,
		}
		switch imp.Status {
		case easydocforms.ImportQueued, easydocforms.ImportProcessing:
			out.Note = "Still working — check again in about 30 seconds."
		case easydocforms.ImportSucceeded:
			if imp.ReviewRequired {
				out.Note = "Template created. A staff member should review it in the EasyDocForms editor before it goes to patients (see review_reasons)."
			} else {
				out.Note = "Template ready. Create a patient link with create_fill_link."
			}
		}
		return nil, out, nil
	})
}
