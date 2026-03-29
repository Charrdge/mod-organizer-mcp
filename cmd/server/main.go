// Command server is the Mod Organizer 2 read-only MCP server: stdio (default) or HTTP streamable.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/charrdge/mod-organizer-mcp/internal/toolreg"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mod-organizer-mcp",
		Version: "0.3.0",
	}, nil)
	toolreg.Register(server)

	mode := strings.ToLower(strings.TrimSpace(os.Getenv("MCP_TRANSPORT")))
	if mode == "" || mode == "stdio" {
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("server: %v", err)
		}
		return
	}

	if mode != "http" {
		log.Fatalf("mod-organizer-mcp: unknown MCP_TRANSPORT=%q (use stdio or http)", mode)
	}

	addr := strings.TrimSpace(os.Getenv("MCP_HTTP_ADDR"))
	if addr == "" {
		addr = ":8080"
	}
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)
	log.Printf("mod-organizer-mcp streamable HTTP on %s (set MCP_TRANSPORT=stdio for Cursor)", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
