// Command easydocforms-mcp is an MCP server exposing the EasyDocForms
// Partner API as agent tools over stdio: import a blank PDF intake form, get
// a hosted patient fill link, and retrieve completed-PDF download links.
//
// Configuration comes from the environment, per the MCP spec's guidance for
// stdio servers:
//
//	EASYDOCFORMS_API_KEY   required — an edfk_live_* Partner API key
//	EASYDOCFORMS_BASE_URL  optional — API base URL override
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	easydocforms "github.com/easydocforms/easydocforms-go"
	"github.com/easydocforms/easydocforms-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is stamped by release builds via -ldflags "-X main.version=…".
// Unstamped builds (`go install …@vX.Y.Z`) resolve it from module build
// info instead, so serverInfo can never drift from the tag again.
var version = ""

func resolveVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}

func main() {
	apiKey := os.Getenv("EASYDOCFORMS_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "easydocforms-mcp: EASYDOCFORMS_API_KEY is not set.")
		fmt.Fprintln(os.Stderr, "Create an API key in the EasyDocForms app under Settings → Integrations → Partner API,")
		fmt.Fprintln(os.Stderr, "then set EASYDOCFORMS_API_KEY=edfk_live_... in this server's environment.")
		os.Exit(2)
	}

	opts := []easydocforms.Option{
		easydocforms.WithUserAgent("easydocforms-mcp/" + resolveVersion()),
	}
	if baseURL := os.Getenv("EASYDOCFORMS_BASE_URL"); baseURL != "" {
		opts = append(opts, easydocforms.WithBaseURL(baseURL))
	}
	client := easydocforms.NewClient(apiKey, opts...)

	server := tools.NewServer(client, resolveVersion())
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("easydocforms-mcp: %v", err)
	}
}
