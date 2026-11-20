package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/oauth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

// OAuth 2.1 authorization server for MCP clients (KEG-12), following the MCP
// authorization spec (2025-06-18): protected resource metadata (RFC 9728),
// authorization server metadata (RFC 8414), dynamic client registration
// (RFC 7591), authorization code + PKCE S256, resource indicators (RFC 8707),
// rotating refresh tokens and token revocation (RFC 7009).

const (
	// ConsentPath is the frontend page that asks the user for consent.
	ConsentPath = "/oauth/consent"

	defaultOAuthClientName = "Unbenannte Anwendung"
	defaultOAuthIssuer     = "http://localhost:8080"
)

// OAuthIssuer is the issuer identifier and base URL of the authorization
// server: PUBLIC_BASE_URL without a trailing slash. It comes from the
// configuration (never from request headers), so it is correct behind a
// reverse proxy and cannot be spoofed via the Host header.
func (h *Handler) OAuthIssuer() string {
	base := strings.TrimRight(strings.TrimSpace(h.Config.BaseURL), "/")
	if base == "" {
		return defaultOAuthIssuer
	}
	return base
}

// MCPResourceURL is the canonical URI of the MCP server (RFC 8707 resource).
func (h *Handler) MCPResourceURL() string { return h.OAuthIssuer() + "/mcp" }

// ProtectedResourceMetadataURL is announced in the WWW-Authenticate header of /mcp.
func (h *Handler) ProtectedResourceMetadataURL() string {
	return h.OAuthIssuer() + "/.well-known/oauth-protected-resource"
}

// HandleOAuthProtectedResourceMetadata serves RFC 9728 metadata for /mcp.
func (h *Handler) HandleOAuthProtectedResourceMetadata(c fiber.Ctx) error {
	if h.OAuthRepo == nil {
		return fiber.ErrNotFound
	}
	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.JSON(fiber.Map{
		"resource":                 h.MCPResourceURL(),
		"authorization_servers":    []string{h.OAuthIssuer()},
		"bearer_methods_supported": []string{"header"},
		"scopes_supported":         []string{oauth.Scope},
		"resource_name":            "Kegelmaster",
	})
}

// HandleOAuthAuthorizationServerMetadata serves RFC 8414 metadata.
func (h *Handler) HandleOAuthAuthorizationServerMetadata(c fiber.Ctx) error {
	if h.OAuthRepo == nil {
		return fiber.ErrNotFound
	}
	issuer := h.OAuthIssuer()
	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.JSON(fiber.Map{
		"issuer":                                         issuer,
		"authorization_endpoint":                         issuer + "/oauth/authorize",
		"token_endpoint":                                 issuer + "/oauth/token",
		"registration_endpoint":                          issuer + "/oauth/register",
		"revocation_endpoint":                            issuer + "/oauth/revoke",
		"response_types_supported":                       []string{"code"},
		"response_modes_supported":                       []string{"query"},
		"grant_types_supported":                          []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":               []string{"S256"},
		"token_endpoint_auth_methods_supported":          []string{"none"},
		"revocation_endpoint_auth_methods_supported":     []string{"none"},
		"scopes_supported":                               []string{oauth.Scope},
		"authorization_response_iss_parameter_supported": true,
	})
}

// oauthError is an OAuth error response ({"error": ..., "error_description": ...}).
func oauthError(c fiber.Ctx, status int, code, description string) error {
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Set("Pragma", "no-cache")
	return c.Status(status).JSON(fiber.Map{"error": code, "error_description": description})
}

// parseOAuthParams parses form- or query-encoded parameters. Parameters must
// not be repeated (RFC 6749, section 3.1 and 3.2).
func parseOAuthParams(raw string) (map[string]string, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, errors.New("Parameter nicht lesbar")
	}
	out := make(map[string]string, len(values))
	for k, v := range values {
		if len(v) != 1 {
			return nil, errors.New("Parameter " + k + " ist mehrfach angegeben")
		}
		out[k] = v[0]
	}
	return out, nil
}

// --- Dynamic client registration (RFC 7591) ---

type oauthRegisterRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	ClientName              string   `json:"client_name"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

// HandleOAuthRegister registers a public client (no client secret).
func (h *Handler) HandleOAuthRegister(c fiber.Ctx) error {
	if h.OAuthRepo == nil {
		return fiber.ErrNotFound
	}
	var req oauthRegisterRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return oauthError(c, fiber.StatusBadRequest, "invalid_client_metadata", "Ungültiges JSON")
	}
	if len(req.RedirectURIs) == 0 || len(req.RedirectURIs) > oauth.MaxRedirectURIs {
		return oauthError(c, fiber.StatusBadRequest, "invalid_redirect_uri", "Es sind 1 bis 5 redirect_uris erforderlich")
	}
	seen := make(map[string]bool, len(req.RedirectURIs))
	uris := make([]string, 0, len(req.RedirectURIs))
	for _, u := range req.RedirectURIs {
		if err := oauth.ValidateRedirectURI(u); err != nil {
			return oauthError(c, fiber.StatusBadRequest, "invalid_redirect_uri", "redirect_uri muss https oder http auf localhost/127.0.0.1/[::1] verwenden und darf kein Fragment enthalten")
		}
		if !seen[u] {
			seen[u] = true
			uris = append(uris, u)
		}
	}
	if m := req.TokenEndpointAuthMethod; m != "" && m != "none" {
		return oauthError(c, fiber.StatusBadRequest, "invalid_client_metadata", "Nur öffentliche Clients (token_endpoint_auth_method \"none\") werden unterstützt")
	}
	for _, g := range req.GrantTypes {
		if g != "authorization_code" && g != "refresh_token" {
			return oauthError(c, fiber.StatusBadRequest, "invalid_client_metadata", "Nicht unterstützter grant_type")
		}
	}
	for _, r := range req.ResponseTypes {
		if r != "code" {
			return oauthError(c, fiber.StatusBadRequest, "invalid_client_metadata", "Nicht unterstützter response_type")
		}
	}
	name := strings.TrimSpace(req.ClientName)
	if name == "" {
		name = defaultOAuthClientName
	}
	if utf8.RuneCountInString(name) > oauth.MaxClientNameLength || strings.ContainsFunc(name, unicode.IsControl) || !utf8.ValidString(name) {
		return oauthError(c, fiber.StatusBadRequest, "invalid_client_metadata", "client_name ist ungültig (höchstens 100 Zeichen)")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()
	client, err := h.OAuthRepo.CreateClient(ctx, name, uris, time.Now().UTC())
	if err != nil {
		slog.Error("register oauth client", "error", err)
		return oauthError(c, fiber.StatusInternalServerError, "server_error", "Interner Fehler")
	}
	audit.LogAudit(c, "", audit.ActionOAuthClientRegistered, "client_id", client.ID)

	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"client_id":                  client.ID,
		"client_id_issued_at":        client.CreatedAt.Unix(),
		"client_name":                client.Name,
		"redirect_uris":              client.RedirectURIs,
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
		"scope":                      oauth.Scope,
	})
}

// --- Authorization endpoint and consent ---

// authorizeRequest holds the parameters of an authorization request.
type authorizeRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	State               string `json:"state"`
	Scope               string `json:"scope"`
	Resource            string `json:"resource"`
}

func authorizeRequestFromParams(p map[string]string) authorizeRequest {
	return authorizeRequest{
		ClientID:            p["client_id"],
		RedirectURI:         p["redirect_uri"],
		ResponseType:        p["response_type"],
		CodeChallenge:       p["code_challenge"],
		CodeChallengeMethod: p["code_challenge_method"],
		State:               p["state"],
		Scope:               p["scope"],
		Resource:            p["resource"],
	}
}

func (r authorizeRequest) query() string {
	v := url.Values{}
	for k, val := range map[string]string{
		"client_id":             r.ClientID,
		"redirect_uri":          r.RedirectURI,
		"response_type":         r.ResponseType,
		"code_challenge":        r.CodeChallenge,
		"code_challenge_method": r.CodeChallengeMethod,
		"state":                 r.State,
		"scope":                 r.Scope,
		"resource":              r.Resource,
	} {
		if val != "" {
			v.Set(k, val)
		}
	}
	return v.Encode()
}

// authorizeError is an error of an authorization request. Only when Redirect
// is true were client and redirect_uri verified, so the error may be sent to
// the client's redirect_uri; otherwise it is shown to the user (no open redirect).
type authorizeError struct {
	Code        string
	Description string
	Redirect    bool
}

func (e *authorizeError) Error() string { return e.Description }

// validateAuthorizeRequest checks an authorization request. redirect_uri
// must exactly match a registered URI before anything is sent there.
func (h *Handler) validateAuthorizeRequest(c fiber.Ctx, req authorizeRequest) (oauth.Client, error) {
	ctx, cancel := h.RequestContext()
	defer cancel()
	client, err := h.OAuthRepo.GetClient(ctx, req.ClientID)
	if errors.Is(err, oauth.ErrNotFound) {
		return oauth.Client{}, &authorizeError{Code: "invalid_client", Description: "Unbekannte Anwendung (client_id)"}
	}
	if err != nil {
		slog.Error("get oauth client", "error", err)
		return oauth.Client{}, fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if req.RedirectURI == "" || !client.HasRedirectURI(req.RedirectURI) {
		return oauth.Client{}, &authorizeError{Code: "invalid_request", Description: "redirect_uri ist für diese Anwendung nicht registriert"}
	}
	fail := func(code, description string) (oauth.Client, error) {
		return oauth.Client{}, &authorizeError{Code: code, Description: description, Redirect: true}
	}
	if !oauth.ValidState(req.State) {
		// The state cannot be echoed safely; do not redirect with it.
		return oauth.Client{}, &authorizeError{Code: "invalid_request", Description: "state ist ungültig"}
	}
	if req.ResponseType != "code" {
		return fail("unsupported_response_type", "Nur response_type=code wird unterstützt")
	}
	if req.CodeChallengeMethod != "S256" || !oauth.ValidCodeChallenge(req.CodeChallenge) {
		return fail("invalid_request", "PKCE mit code_challenge_method=S256 ist erforderlich")
	}
	if !oauth.ValidScopeSyntax(req.Scope) {
		return fail("invalid_scope", "scope ist ungültig")
	}
	if req.Resource != "" && !oauth.SameResource(req.Resource, h.MCPResourceURL()) {
		return fail("invalid_target", "Unbekannte resource; erwartet wird "+h.MCPResourceURL())
	}
	return client, nil
}

// errorRedirect builds the redirect to the client for an authorization error.
func (h *Handler) errorRedirect(req authorizeRequest, code, description string) string {
	return oauth.AppendQuery(req.RedirectURI, map[string]string{
		"error":             code,
		"error_description": description,
		"state":             req.State,
		"iss":               h.OAuthIssuer(),
	})
}

// HandleOAuthAuthorize is the authorization endpoint (GET /oauth/authorize).
// It validates the request and forwards the browser to the consent page of the
// frontend, which sends users who are not logged in to the login first.
func (h *Handler) HandleOAuthAuthorize(c fiber.Ctx) error {
	if h.OAuthRepo == nil {
		return fiber.ErrNotFound
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	params, err := parseOAuthParams(string(c.Request().URI().QueryString()))
	if err != nil {
		return authorizeErrorPage(c, err.Error())
	}
	req := authorizeRequestFromParams(params)
	if _, err := h.validateAuthorizeRequest(c, req); err != nil {
		var ae *authorizeError
		if !errors.As(err, &ae) {
			return err
		}
		if !ae.Redirect {
			return authorizeErrorPage(c, ae.Description)
		}
		return redirectTo(c, h.errorRedirect(req, ae.Code, ae.Description))
	}
	return redirectTo(c, ConsentPath+"?"+req.query())
}

func redirectTo(c fiber.Ctx, location string) error {
	c.Set(fiber.HeaderLocation, location)
	return c.SendStatus(fiber.StatusFound)
}

// authorizeErrorPage tells the user about a request that cannot be sent back
// to the client (unknown client or redirect_uri).
func authorizeErrorPage(c fiber.Ctx, description string) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	return c.Status(fiber.StatusBadRequest).SendString("Ungültige Autorisierungsanfrage: " + description +
		"\n\nBitte starten Sie die Verbindung in Ihrer Anwendung erneut.")
}

// sessionUserForOAuth authenticates consent and grant management requests.
// Like API token management, they need a browser session: a bearer token
// must never approve further clients or manage connections.
func (h *Handler) sessionUserForOAuth(c fiber.Ctx) (user.User, error) {
	if IsAPITokenRequest(c) {
		return user.User{}, fiber.NewError(fiber.StatusForbidden, "Verbundene Anwendungen können nur über die Weboberfläche verwaltet werden")
	}
	if h.OAuthRepo == nil {
		return user.User{}, fiber.NewError(fiber.StatusServiceUnavailable, "OAuth ist nicht verfügbar")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()
	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return user.User{}, fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}
	return u, nil
}

// OAuthConsentInfo describes a pending authorization request for the consent page.
type OAuthConsentInfo struct {
	ClientID     string `json:"client_id"`
	ClientName   string `json:"client_name"`
	RedirectURI  string `json:"redirect_uri"`
	RedirectHost string `json:"redirect_host"`
	Scope        string `json:"scope"`
}

func redirectHost(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	return u.Host
}

// HandleOAuthConsentInfo returns what the consent page shows (GET /api/oauth/authorize).
func (h *Handler) HandleOAuthConsentInfo(c fiber.Ctx) error {
	if _, err := h.sessionUserForOAuth(c); err != nil {
		return err
	}
	params, err := parseOAuthParams(string(c.Request().URI().QueryString()))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	req := authorizeRequestFromParams(params)
	client, err := h.validateAuthorizeRequest(c, req)
	if err != nil {
		var ae *authorizeError
		if errors.As(err, &ae) {
			return fiber.NewError(fiber.StatusBadRequest, ae.Description)
		}
		return err
	}
	return c.JSON(OAuthConsentInfo{
		ClientID:     client.ID,
		ClientName:   client.Name,
		RedirectURI:  req.RedirectURI,
		RedirectHost: redirectHost(req.RedirectURI),
		Scope:        oauth.Scope,
	})
}

type oauthConsentRequest struct {
	authorizeRequest
	Approve bool `json:"approve"`
}

// HandleOAuthConsent records the user's decision (POST /api/oauth/authorize)
// and returns the redirect to the client: with a code when approved, with
// error=access_denied otherwise.
func (h *Handler) HandleOAuthConsent(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.sessionUserForOAuth(c)
	if err != nil {
		return err
	}
	var req oauthConsentRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	client, err := h.validateAuthorizeRequest(c, req.authorizeRequest)
	if err != nil {
		var ae *authorizeError
		if !errors.As(err, &ae) {
			return err
		}
		if !ae.Redirect {
			return fiber.NewError(fiber.StatusBadRequest, ae.Description)
		}
		return c.JSON(fiber.Map{"redirect_to": h.errorRedirect(req.authorizeRequest, ae.Code, ae.Description)})
	}

	if !req.Approve {
		audit.LogAudit(c, u.ID, audit.ActionOAuthConsentDenied, "client_id", client.ID)
		return c.JSON(fiber.Map{"redirect_to": h.errorRedirect(req.authorizeRequest, "access_denied", "Der Benutzer hat den Zugriff abgelehnt")})
	}

	resource := req.Resource
	if resource == "" {
		resource = h.MCPResourceURL()
	}
	ctx, cancel := h.RequestContext()
	defer cancel()
	code, err := h.OAuthRepo.CreateCode(ctx, oauth.Code{
		ClientID:      client.ID,
		UserID:        u.ID,
		RedirectURI:   req.RedirectURI,
		CodeChallenge: req.CodeChallenge,
		Scope:         oauth.Scope,
		Resource:      resource,
	}, time.Now().UTC())
	if err != nil {
		slog.Error("create oauth authorization code", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAudit(c, u.ID, audit.ActionOAuthConsentGranted, "client_id", client.ID, "redirect_host", redirectHost(req.RedirectURI))
	return c.JSON(fiber.Map{"redirect_to": oauth.AppendQuery(req.RedirectURI, map[string]string{
		"code":  code,
		"state": req.State,
		"iss":   h.OAuthIssuer(),
	})})
}

// --- Token endpoint ---

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// readOAuthForm parses an application/x-www-form-urlencoded request body.
func readOAuthForm(c fiber.Ctx) (map[string]string, error) {
	ct := strings.ToLower(strings.TrimSpace(c.Get(fiber.HeaderContentType)))
	if !strings.HasPrefix(ct, fiber.MIMEApplicationForm) {
		return nil, errors.New("Content-Type muss application/x-www-form-urlencoded sein")
	}
	return parseOAuthParams(string(c.Body()))
}

// HandleOAuthToken is the token endpoint (POST /oauth/token).
func (h *Handler) HandleOAuthToken(c fiber.Ctx) error {
	if h.OAuthRepo == nil {
		return fiber.ErrNotFound
	}
	params, err := readOAuthForm(c)
	if err != nil {
		return oauthError(c, fiber.StatusBadRequest, "invalid_request", err.Error())
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	client, err := h.OAuthRepo.GetClient(ctx, params["client_id"])
	if errors.Is(err, oauth.ErrNotFound) {
		return oauthError(c, fiber.StatusUnauthorized, "invalid_client", "Unbekannte client_id")
	}
	if err != nil {
		slog.Error("get oauth client", "error", err)
		return oauthError(c, fiber.StatusInternalServerError, "server_error", "Interner Fehler")
	}

	now := time.Now().UTC()
	switch params["grant_type"] {
	case "authorization_code":
		return h.exchangeAuthorizationCode(c, client, params, now)
	case "refresh_token":
		return h.exchangeRefreshToken(c, client, params, now)
	case "":
		return oauthError(c, fiber.StatusBadRequest, "invalid_request", "grant_type fehlt")
	default:
		return oauthError(c, fiber.StatusBadRequest, "unsupported_grant_type", "Nur authorization_code und refresh_token werden unterstützt")
	}
}

func (h *Handler) exchangeAuthorizationCode(c fiber.Ctx, client oauth.Client, params map[string]string, now time.Time) error {
	if params["code"] == "" || params["redirect_uri"] == "" || params["code_verifier"] == "" {
		return oauthError(c, fiber.StatusBadRequest, "invalid_request", "code, redirect_uri und code_verifier sind erforderlich")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	// The code is consumed first (atomically), so it is single use even if a
	// later check fails.
	code, err := h.OAuthRepo.ConsumeCode(ctx, params["code"], now)
	switch {
	case errors.Is(err, oauth.ErrReused):
		audit.LogAudit(c, code.UserID, audit.ActionOAuthCodeReused, "client_id", client.ID)
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Code wurde bereits verwendet")
	case errors.Is(err, oauth.ErrNotFound):
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Ungültiger Code")
	case err != nil:
		slog.Error("consume oauth code", "error", err)
		return oauthError(c, fiber.StatusInternalServerError, "server_error", "Interner Fehler")
	}
	if code.ClientID != client.ID {
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Code gehört zu einer anderen Anwendung")
	}
	if !now.Before(code.ExpiresAt) {
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Code ist abgelaufen")
	}
	if params["redirect_uri"] != code.RedirectURI {
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "redirect_uri stimmt nicht überein")
	}
	if !oauth.VerifyPKCE(params["code_verifier"], code.CodeChallenge) {
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "PKCE-Prüfung fehlgeschlagen")
	}
	if r := params["resource"]; r != "" && !oauth.SameResource(r, code.Resource) {
		return oauthError(c, fiber.StatusBadRequest, "invalid_target", "resource stimmt nicht überein")
	}
	u, err := h.UserRepo.GetByID(ctx, code.UserID)
	if err != nil || u.Disabled() {
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Konto nicht verfügbar")
	}

	grant, tokens, err := h.OAuthRepo.IssueForCode(ctx, code, now)
	if err != nil {
		slog.Error("issue oauth tokens", "error", err)
		return oauthError(c, fiber.StatusInternalServerError, "server_error", "Interner Fehler")
	}
	audit.LogAudit(c, u.ID, audit.ActionOAuthGrantCreated, "client_id", client.ID, "grant_id", grant.ID)
	return sendTokens(c, tokens)
}

func (h *Handler) exchangeRefreshToken(c fiber.Ctx, client oauth.Client, params map[string]string, now time.Time) error {
	if params["refresh_token"] == "" {
		return oauthError(c, fiber.StatusBadRequest, "invalid_request", "refresh_token ist erforderlich")
	}
	if !oauth.ValidScopeSyntax(params["scope"]) {
		return oauthError(c, fiber.StatusBadRequest, "invalid_scope", "scope ist ungültig")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	grant, err := h.OAuthRepo.UseRefreshToken(ctx, params["refresh_token"], now)
	switch {
	case errors.Is(err, oauth.ErrReused):
		// Token theft or a broken client: the whole chain is revoked.
		audit.LogAudit(c, grant.UserID, audit.ActionOAuthRefreshReused, "client_id", grant.ClientID, "grant_id", grant.ID)
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Refresh-Token wurde bereits verwendet; die Verbindung wurde widerrufen")
	case errors.Is(err, oauth.ErrNotFound), errors.Is(err, oauth.ErrExpired):
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Ungültiges oder abgelaufenes Refresh-Token")
	case err != nil:
		slog.Error("use oauth refresh token", "error", err)
		return oauthError(c, fiber.StatusInternalServerError, "server_error", "Interner Fehler")
	}
	if grant.ClientID != client.ID {
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Refresh-Token gehört zu einer anderen Anwendung")
	}
	if r := params["resource"]; r != "" && !oauth.SameResource(r, grant.Resource) {
		return oauthError(c, fiber.StatusBadRequest, "invalid_target", "resource stimmt nicht überein")
	}
	u, err := h.UserRepo.GetByID(ctx, grant.UserID)
	if err != nil || u.Disabled() || !u.SessionValid(grant.CreatedAt) {
		return oauthError(c, fiber.StatusBadRequest, "invalid_grant", "Die Verbindung ist nicht mehr gültig")
	}

	tokens, err := h.OAuthRepo.IssueForGrant(ctx, grant, now)
	if err != nil {
		slog.Error("issue oauth tokens", "error", err)
		return oauthError(c, fiber.StatusInternalServerError, "server_error", "Interner Fehler")
	}
	return sendTokens(c, tokens)
}

func sendTokens(c fiber.Ctx, t oauth.IssuedTokens) error {
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Set("Pragma", "no-cache")
	return c.JSON(oauthTokenResponse{
		AccessToken:  t.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    t.ExpiresIn,
		RefreshToken: t.RefreshToken,
		Scope:        t.Scope,
	})
}

// HandleOAuthRevoke is the revocation endpoint (POST /oauth/revoke, RFC 7009).
// Revoking an access or refresh token ends the whole connection (grant).
func (h *Handler) HandleOAuthRevoke(c fiber.Ctx) error {
	if h.OAuthRepo == nil {
		return fiber.ErrNotFound
	}
	params, err := readOAuthForm(c)
	if err != nil {
		return oauthError(c, fiber.StatusBadRequest, "invalid_request", err.Error())
	}
	if params["token"] == "" {
		return oauthError(c, fiber.StatusBadRequest, "invalid_request", "token ist erforderlich")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()
	client, err := h.OAuthRepo.GetClient(ctx, params["client_id"])
	if errors.Is(err, oauth.ErrNotFound) {
		return oauthError(c, fiber.StatusUnauthorized, "invalid_client", "Unbekannte client_id")
	}
	if err != nil {
		slog.Error("get oauth client", "error", err)
		return oauthError(c, fiber.StatusInternalServerError, "server_error", "Interner Fehler")
	}
	grantID, err := h.OAuthRepo.RevokeToken(ctx, params["token"], client.ID)
	if err != nil {
		slog.Error("revoke oauth token", "error", err)
		return oauthError(c, fiber.StatusServiceUnavailable, "temporarily_unavailable", "Bitte später erneut versuchen")
	}
	if grantID != "" {
		audit.LogAudit(c, "", audit.ActionOAuthTokenRevoked, "client_id", client.ID, "grant_id", grantID)
	}
	// Unknown tokens are answered with 200 as well (RFC 7009, section 2.2).
	return c.SendStatus(fiber.StatusOK)
}

// --- Connected applications (grant management) ---

// OAuthGrantResponse describes a connected application.
type OAuthGrantResponse struct {
	ID           string     `json:"id"`
	ClientID     string     `json:"client_id"`
	ClientName   string     `json:"client_name"`
	RedirectHost string     `json:"redirect_host"`
	Scope        string     `json:"scope"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at"`
}

// HandleListOAuthGrants lists the user's connected applications.
func (h *Handler) HandleListOAuthGrants(c fiber.Ctx) error {
	u, err := h.sessionUserForOAuth(c)
	if err != nil {
		return err
	}
	ctx, cancel := h.RequestContext()
	defer cancel()
	grants, err := h.OAuthRepo.ListGrants(ctx, u.ID)
	if err != nil {
		slog.Error("list oauth grants", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	out := make([]OAuthGrantResponse, 0, len(grants))
	for _, g := range grants {
		host := ""
		if len(g.RedirectURIs) > 0 {
			host = redirectHost(g.RedirectURIs[0])
		}
		out = append(out, OAuthGrantResponse{
			ID:           g.ID,
			ClientID:     g.ClientID,
			ClientName:   g.ClientName,
			RedirectHost: host,
			Scope:        g.Scope,
			CreatedAt:    g.CreatedAt,
			LastUsedAt:   g.LastUsedAt,
		})
	}
	return c.JSON(out)
}

// HandleDeleteOAuthGrant revokes one of the user's connected applications.
func (h *Handler) HandleDeleteOAuthGrant(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.sessionUserForOAuth(c)
	if err != nil {
		return err
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige ID")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()
	if err := h.OAuthRepo.DeleteGrantForUser(ctx, id, u.ID); err != nil {
		if errors.Is(err, oauth.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Verbindung nicht gefunden")
		}
		slog.Error("delete oauth grant", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAudit(c, u.ID, audit.ActionOAuthGrantRevoked, "grant_id", id)
	return c.SendStatus(fiber.StatusNoContent)
}
