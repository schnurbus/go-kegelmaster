package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/valyala/fasthttp"
)

// API performs a request against the Kegelmaster REST API on behalf of the
// caller identified by token. Implementations must not fall back to any
// other credentials.
type API interface {
	Do(ctx context.Context, token, method, path string, query url.Values, body any) (status int, respBody []byte, err error)
}

// InProcessAPI dispatches requests directly into the Fiber request handler,
// so MCP tools go through exactly the same middleware, auth, permission
// checks, validation and audit logging as the web frontend, without a network
// round trip.
type InProcessAPI struct {
	handler fasthttp.RequestHandler
}

func NewInProcessAPI() *InProcessAPI { return &InProcessAPI{} }

// SetHandler must be called once all routes are registered.
func (a *InProcessAPI) SetHandler(h fasthttp.RequestHandler) { a.handler = h }

var loopbackAddr = &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}

func (a *InProcessAPI) Do(_ context.Context, token, method, path string, query url.Values, body any) (int, []byte, error) {
	if a.handler == nil {
		return 0, nil, errors.New("api handler not initialised")
	}
	var req fasthttp.Request
	uri := path
	if len(query) > 0 {
		uri += "?" + query.Encode()
	}
	req.SetRequestURI(uri)
	req.Header.SetMethod(method)
	req.Header.SetHost("localhost")
	req.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+token)
	req.Header.Set("X-Kegelmaster-Client", "mcp")
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		req.Header.SetContentType("application/json")
		req.SetBody(raw)
	}

	var rc fasthttp.RequestCtx
	rc.Init(&req, loopbackAddr, nil)
	a.handler(&rc)
	return rc.Response.StatusCode(), bytes.Clone(rc.Response.Body()), nil
}

// apiError is an error reported by the REST API; Message is user-facing (German).
type apiError struct {
	Status  int
	Message string
}

func (e *apiError) Error() string { return e.Message }

func isStatus(err error, status int) bool {
	var ae *apiError
	return errors.As(err, &ae) && ae.Status == status
}

// call performs a request and decodes a JSON response into out (if non-nil).
func call(ctx context.Context, api API, token, method, path string, query url.Values, body, out any) error {
	status, raw, err := api.Do(ctx, token, method, path, query, body)
	if err != nil {
		return err
	}
	if status >= 400 {
		var payload struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(raw, &payload)
		if payload.Message == "" {
			payload.Message = fmt.Sprintf("API-Fehler (HTTP %d)", status)
		}
		return &apiError{Status: status, Message: payload.Message}
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}
