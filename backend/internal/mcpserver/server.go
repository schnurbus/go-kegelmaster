// Package mcpserver exposes the Kegelmaster API as a Model Context Protocol
// server (Streamable HTTP, stateless) at /mcp. Clients authenticate with a
// personal API token ("Authorization: Bearer kgm_...") or an OAuth access
// token ("kgo_...", see internal/oauth); every tool forwards to the REST API
// with that token, so permissions, validation, audit logging and the balance
// invariants are exactly those of the web frontend.
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const tokenExtraKey = "kegelmaster_token"

const serverInstructions = `Kegelmaster verwaltet Kegelklubs: Spieler, Spieltage (Spielabende) mit Strafen und Wettbewerben sowie die Klubkasse.

Vorgehen:
1. klubs_auflisten liefert die Klubs und deine Rechte darin.
2. klub_ueberblick liefert alle IDs, die schreibende Tools brauchen (Spieler, Strafentypen, Wettbewerbe). Ordne Namen aus Nutzereingaben immer anhand dieser Liste zu und frage nach, wenn ein Name nicht eindeutig ist.
3. Geldbeträge sind immer ganze Cent (betrag_cent). Ein negativer Saldo bedeutet, dass der Spieler der Kasse Geld schuldet. Strafen und Grundgebühren belasten nur den Spielersaldo; der Kassenstand ändert sich erst durch Einzahlungen, Trinkgeld und Ausgaben.

Spieltage werden in zwei Schritten erfasst: spieltag_erfassen legt einen Entwurf an (es wird nichts gebucht). Erst spieltag_veroeffentlichen bucht Grundgebühren und Strafen. Zeige dem Nutzer vorher die Vorschau aus spieltag_details und lass dir die Veröffentlichung bestätigen.`

// New returns the HTTP handler for the MCP endpoint. resourceMetadataURL is
// announced in the WWW-Authenticate challenge of 401 responses (RFC 9728), so
// OAuth clients can discover the authorization server; empty omits it.
func New(api API, version, resourceMetadataURL string) http.Handler {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "kegelmaster",
		Title:   "Kegelmaster",
		Version: version,
	}, &mcp.ServerOptions{Instructions: serverInstructions})

	registerTools(server, api)

	streamable := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	})

	verifier := func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		var me struct {
			ID string `json:"id"`
		}
		if err := call(ctx, api, token, http.MethodGet, "/api/auth/me", nil, nil, &me); err != nil {
			if isStatus(err, http.StatusUnauthorized) {
				return nil, auth.ErrInvalidToken
			}
			return nil, err
		}
		return &auth.TokenInfo{UserID: me.ID, Extra: map[string]any{tokenExtraKey: token}}, nil
	}

	protected := auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL: resourceMetadataURL,
		// Personal API tokens may have no expiry; OAuth access tokens expire,
		// which the REST API checks on every request.
		AllowMissingExpiration: true,
	})(streamable)
	return withBearerChallenge(protected)
}

// withBearerChallenge adds the standard "WWW-Authenticate: Bearer" challenge
// to 401 responses when the SDK did not set one (no OAuth metadata configured).
func withBearerChallenge(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&challengeWriter{ResponseWriter: w}, r)
	})
}

type challengeWriter struct{ http.ResponseWriter }

func (c *challengeWriter) WriteHeader(code int) {
	if code == http.StatusUnauthorized && c.Header().Get("WWW-Authenticate") == "" {
		c.Header().Set("WWW-Authenticate", `Bearer realm="kegelmaster", error="invalid_token", error_description="send a personal API token (kgm_...) as bearer token"`)
	}
	c.ResponseWriter.WriteHeader(code)
}

func (c *challengeWriter) Flush() {
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func tokenFrom(req *mcp.CallToolRequest) (string, error) {
	if req.Extra != nil && req.Extra.TokenInfo != nil {
		if t, ok := req.Extra.TokenInfo.Extra[tokenExtraKey].(string); ok && t != "" {
			return t, nil
		}
	}
	return "", errors.New("nicht angemeldet")
}

// toolFunc is the signature of a tool implementation: it returns a value that
// is rendered as JSON for the model, or an error shown to the model.
type toolFunc[In any] func(ctx context.Context, c *client, in In) (any, error)

type client struct {
	api   API
	token string
}

func (c *client) get(ctx context.Context, path string, query map[string]string, out any) error {
	return call(ctx, c.api, c.token, http.MethodGet, path, toValues(query), nil, out)
}

func (c *client) send(ctx context.Context, method, path string, body, out any) error {
	return call(ctx, c.api, c.token, method, path, nil, body, out)
}

func addTool[In any](s *mcp.Server, api API, tool *mcp.Tool, fn toolFunc[In]) {
	mcp.AddTool(s, tool, func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		token, err := tokenFrom(req)
		if err != nil {
			return errorResult(err), nil, nil
		}
		out, err := fn(ctx, &client{api: api, token: token}, in)
		if err != nil {
			return errorResult(err), nil, nil
		}
		raw, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}, nil, nil
	})
}

func errorResult(err error) *mcp.CallToolResult {
	msg := err.Error()
	var ae *apiError
	if errors.As(err, &ae) && ae.Status >= 500 {
		msg = "Interner Fehler im Kegelmaster-Server. Bitte später erneut versuchen."
	}
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "Fehler: " + strings.TrimPrefix(msg, "Fehler: ")}}}
}

func boolPtr(b bool) *bool { return &b }
