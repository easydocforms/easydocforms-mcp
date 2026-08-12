package tools

import (
	"context"
	"time"

	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getSubmissionInput struct {
	SubmissionID string `json:"submission_id" jsonschema:"the submission id, e.g. from a submission.created webhook"`
}

// getSubmissionOutput deliberately has no answers field: the PHI boundary of
// this server is that patient answers never pass through agent context.
type getSubmissionOutput struct {
	SubmissionID       string    `json:"submission_id"`
	SubmittedAt        time.Time `json:"submitted_at"`
	TemplateID         string    `json:"template_id"`
	FillLinkID         string    `json:"fill_link_id,omitempty"`
	ExternalRef        string    `json:"external_ref,omitempty" jsonschema:"the correlation id set when the fill link was created"`
	CompletedPDFStatus string    `json:"completed_pdf_status" jsonschema:"ready = get_completed_pdf_link can sign a download URL now; pending = the PDF is still being frozen"`
	AnswerCount        int       `json:"answer_count" jsonschema:"how many fields the patient answered"`
	Note               string    `json:"note"`
}

func addGetSubmission(server *mcp.Server, client *easydocforms.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_submission",
		Description: "Look up a patient submission: when it arrived, which template and fill link " +
			"produced it, your external_ref, and whether the completed PDF is ready. PHI boundary: " +
			"returns metadata and an answer count only — never the patient's answers. Answers are " +
			"retrieved server-side via GET /api/v1/submissions/{id} with the API key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getSubmissionInput) (*mcp.CallToolResult, getSubmissionOutput, error) {
		var zero getSubmissionOutput
		sub, err := client.GetSubmission(ctx, input.SubmissionID)
		if err != nil {
			return nil, zero, err
		}
		return nil, getSubmissionOutput{
			SubmissionID:       sub.SubmissionID,
			SubmittedAt:        sub.SubmittedAt,
			TemplateID:         sub.TemplateID,
			FillLinkID:         sub.FillLinkID,
			ExternalRef:        sub.ExternalRef,
			CompletedPDFStatus: sub.CompletedPDFStatus,
			AnswerCount:        len(sub.Answers),
			Note:               "Patient answers are not exposed through MCP tools by design. Use get_completed_pdf_link for the completed PDF, or fetch answers server-side with the API key.",
		}, nil
	})
}
