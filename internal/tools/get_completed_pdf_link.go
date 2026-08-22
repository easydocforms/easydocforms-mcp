package tools

import (
	"context"
	"time"

	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getCompletedPDFLinkInput struct {
	SubmissionID string `json:"submission_id" jsonschema:"the submission whose completed PDF to link"`
}

type getCompletedPDFLinkOutput struct {
	URL       string     `json:"url,omitempty" jsonschema:"time-limited signed download URL (~10 minutes), works without any auth header"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Pending   bool       `json:"pending,omitempty" jsonschema:"true when the frozen PDF is not available yet"`
	Note      string     `json:"note"`
}

func addGetCompletedPDFLink(server *mcp.Server, client *easydocforms.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:  "get_completed_pdf_link",
		Title: "Get completed-PDF download link",
		Description: "Get a short-lived signed download URL for a submission's completed, pixel-exact " +
			"PDF — safe to hand to the user or an EMR without exposing the API key. The URL expires " +
			"in about 10 minutes; treat it like a fax (anyone holding it can download this one " +
			"document until then). Every call is audit-logged.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: boolPtr(false),
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getCompletedPDFLinkInput) (*mcp.CallToolResult, getCompletedPDFLinkOutput, error) {
		var zero getCompletedPDFLinkOutput
		link, err := client.GetSubmissionPDFLink(ctx, input.SubmissionID)
		if easydocforms.IsPDFPending(err) {
			return nil, getCompletedPDFLinkOutput{
				Pending: true,
				Note:    "The completed PDF is still being frozen — this is brief. Retry in ~30 seconds, or fetch the PDF server-side via GET /api/v1/submissions/{id}/pdf, which renders on demand.",
			}, nil
		}
		if err != nil {
			return nil, zero, err
		}
		expiresAt := link.ExpiresAt
		return nil, getCompletedPDFLinkOutput{
			URL:       link.URL,
			ExpiresAt: &expiresAt,
			Note:      "Signed URL, valid ~10 minutes. Treat like a fax: anyone with the URL can download this one document until it expires.",
		}, nil
	})
}
