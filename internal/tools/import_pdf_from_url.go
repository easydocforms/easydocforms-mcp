package tools

import (
	"context"
	"errors"

	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type importPDFFromURLInput struct {
	PDFURL               string `json:"pdf_url" jsonschema:"public HTTPS URL of the blank PDF to import (private/internal addresses are rejected)"`
	Filename             string `json:"filename" jsonschema:"original filename, e.g. new-patient-intake.pdf"`
	Title                string `json:"title,omitempty" jsonschema:"optional display title for the resulting template"`
	BlankFormAttestation bool   `json:"blank_form_attestation" jsonschema:"must be true — attests the PDF is a blank form template containing no patient information (PHI). Confirm with the user before setting it."`
}

type importPDFFromURLOutput struct {
	ImportID string `json:"import_id"`
	Status   string `json:"status"`
	Note     string `json:"note"`
}

func addImportPDFFromURL(server *mcp.Server, client *easydocforms.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "import_pdf_from_url",
		Description: "Import a blank PDF intake form into EasyDocForms, turning it into a hosted, " +
			"mobile-friendly fillable form. Async: returns an import_id immediately; processing " +
			"typically takes 1–10 minutes — poll get_import_status. BLANK FORMS ONLY: the PDF must " +
			"contain no patient information (PHI), and blank_form_attestation must be true.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input importPDFFromURLInput) (*mcp.CallToolResult, importPDFFromURLOutput, error) {
		var zero importPDFFromURLOutput
		if !input.BlankFormAttestation {
			return nil, zero, errors.New("blank_form_attestation must be true: confirm with the user that the PDF is a blank form template containing no patient information (PHI), then retry")
		}
		created, err := client.CreateImport(ctx, easydocforms.CreateImportParams{
			PDFURL:               input.PDFURL,
			Filename:             input.Filename,
			Title:                input.Title,
			BlankFormAttestation: true,
		})
		if err != nil {
			return nil, zero, err
		}
		return nil, importPDFFromURLOutput{
			ImportID: created.ImportID,
			Status:   created.Status,
			Note:     "Processing typically takes 1–10 minutes. Poll get_import_status with this import_id.",
		}, nil
	})
}
