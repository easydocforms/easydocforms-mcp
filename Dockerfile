FROM golang:1.25 AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/easydocforms-mcp ./cmd/easydocforms-mcp

FROM gcr.io/distroless/static-debian12:nonroot
# The official MCP registry verifies image ownership via this label; it must
# match the server name in server.json exactly.
LABEL io.modelcontextprotocol.server.name="com.easydocforms/easydocforms-mcp"
LABEL org.opencontainers.image.source="https://github.com/easydocforms/easydocforms-mcp"
LABEL org.opencontainers.image.description="EasyDocForms MCP server: healthcare intake forms as agent tools"
LABEL org.opencontainers.image.licenses="MIT"
COPY --from=build /out/easydocforms-mcp /easydocforms-mcp
ENTRYPOINT ["/easydocforms-mcp"]
