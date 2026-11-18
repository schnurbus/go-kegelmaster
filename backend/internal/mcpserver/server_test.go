package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	goodToken = "kgm_good"
	clubID    = "club-1"
)

type recordedCall struct {
	Method, Path string
	Query        url.Values
	Body         map[string]any
}

type route struct {
	status int
	body   any
}

// fakeAPI answers by "METHOD path" and records every call.
type fakeAPI struct {
	mu     sync.Mutex
	routes map[string]route
	calls  []recordedCall
}

func newFakeAPI() *fakeAPI {
	f := &fakeAPI{routes: map[string]route{}}
	f.on("GET /api/auth/me", 200, map[string]any{"id": "user-1", "email": "a@example.test"})
	return f
}

func (f *fakeAPI) on(key string, status int, body any) { f.routes[key] = route{status, body} }

func (f *fakeAPI) Do(_ context.Context, token, method, path string, query url.Values, body any) (int, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var b map[string]any
	if body != nil {
		raw, _ := json.Marshal(body)
		_ = json.Unmarshal(raw, &b)
	}
	f.calls = append(f.calls, recordedCall{method, path, query, b})
	if token != goodToken {
		return 401, []byte(`{"message":"Nicht angemeldet"}`), nil
	}
	r, ok := f.routes[method+" "+path]
	if !ok {
		return 404, []byte(`{"message":"nicht gefunden: ` + method + " " + path + `"}`), nil
	}
	raw, _ := json.Marshal(r.body)
	return r.status, raw, nil
}

func (f *fakeAPI) called(method, path string) []recordedCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []recordedCall
	for _, c := range f.calls {
		if c.Method == method && c.Path == path {
			out = append(out, c)
		}
	}
	return out
}

type bearerTransport struct{ token string }

func (b bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(r)
}

func connect(t *testing.T, api API) *mcp.ClientSession {
	t.Helper()
	srv := httptest.NewServer(New(api, "test"))
	t.Cleanup(srv.Close)
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint:   srv.URL,
		HTTPClient: &http.Client{Transport: bearerTransport{goodToken}},
		MaxRetries: -1,
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func callTool(t *testing.T, s *mcp.ClientSession, name string, args any) (string, bool) {
	t.Helper()
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	var sb strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String(), res.IsError
}

func TestAuth_RejectsMissingAndInvalidToken(t *testing.T) {
	srv := httptest.NewServer(New(newFakeAPI(), "test"))
	defer srv.Close()

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	for name, header := range map[string]string{"missing": "", "invalid": "Bearer kgm_wrong"} {
		req, _ := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s token: expected 401, got %d", name, resp.StatusCode)
		}
	}
}

func TestListTools(t *testing.T) {
	s := connect(t, newFakeAPI())
	res, err := s.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"klubs_auflisten": false, "klub_ueberblick": false, "spieltage_auflisten": false, "spieltag_details": false,
		"spieler_konto": false, "transaktionen_suchen": false, "mein_spieler": false, "meine_statistik": false,
		"spieltag_erfassen": false, "spieltag_veroeffentlichen": false, "buchung_erfassen": false,
	}
	for _, tool := range res.Tools {
		if _, ok := want[tool.Name]; !ok {
			t.Errorf("unexpected tool %s", tool.Name)
			continue
		}
		want[tool.Name] = true
		destructive := tool.Annotations != nil && tool.Annotations.DestructiveHint != nil && *tool.Annotations.DestructiveHint
		readOnly := tool.Annotations != nil && tool.Annotations.ReadOnlyHint
		switch tool.Name {
		case "spieltag_veroeffentlichen", "buchung_erfassen":
			if !destructive || readOnly {
				t.Errorf("%s must be destructive", tool.Name)
			}
		case "spieltag_erfassen":
			if destructive || readOnly {
				t.Errorf("%s must be a non-destructive write", tool.Name)
			}
		default:
			if !readOnly {
				t.Errorf("%s must be read-only", tool.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("missing tool %s", name)
		}
	}
}

func TestListClubs(t *testing.T) {
	api := newFakeAPI()
	api.on("GET /api/clubs", 200, []map[string]any{{"id": clubID, "name": "Pudelkönige", "balance": 123456, "base_fee": 500}})
	api.on("GET /api/clubs/club-1/permissions/me", 200, map[string]any{"is_owner": false, "permissions": []map[string]string{
		{"entity_type": "game_days", "permission_type": "list"}, {"entity_type": "game_days", "permission_type": "create"},
	}})
	out, isErr := callTool(t, connect(t, api), "klubs_auflisten", map[string]any{})
	if isErr {
		t.Fatal(out)
	}
	for _, want := range []string{`"club_id": "club-1"`, `"1.234,56 €"`, `"5,00 €"`, `"game_days"`, `"create"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %s:\n%s", want, out)
		}
	}
}

func setupDraftAPI() *fakeAPI {
	api := newFakeAPI()
	gdPath := "/api/clubs/club-1/gamedays/gd-1"
	api.on("POST /api/clubs/club-1/gamedays", 201, map[string]any{"id": "gd-1", "date": "2026-11-13", "is_draft": true})
	api.on("POST "+gdPath+"/participants", 201, map[string]any{})
	api.on("PUT "+gdPath+"/participants/p-1/fees", 200, map[string]any{})
	api.on("PUT "+gdPath+"/participants/p-1/competition-values", 200, map[string]any{})
	api.on("PUT "+gdPath+"/participants/p-2/fees", 200, map[string]any{})
	api.on("DELETE "+gdPath, 204, nil)
	api.on("GET "+gdPath, 200, map[string]any{
		"game_day": map[string]any{"id": "gd-1", "date": "2026-11-13", "notes": "Heimspiel", "is_draft": true},
		"participants": []map[string]any{{
			"participant":        map[string]any{"player_id": "p-1", "player_name": "Anna"},
			"fees":               []map[string]any{{"penalty_type_id": "pt-1", "penalty_type_name": "Pumpe", "penalty_type_price": 50, "count": 3, "quantity_scale": 1, "quantity": 3}},
			"competition_values": []map[string]any{{"competition_id": "c-1", "value": 7}},
		}},
	})
	api.on("GET /api/clubs/club-1/penalty-types", 200, []map[string]any{
		{"id": "pt-1", "name": "Pumpe", "price": 50},
		{"id": "pt-2", "name": "Rinne", "price": 20, "allows_decimal_quantity": true},
	})
	api.on("GET /api/clubs/club-1/competitions", 200, []map[string]any{{"id": "c-1", "name": "Pudelkönig", "scoring_type": "winner"}})
	api.on("GET /api/clubs/club-1", 200, map[string]any{"id": clubID, "name": "K", "base_fee": 300})
	api.on("GET /api/clubs/club-1/roles", 200, []map[string]any{{"id": "r-pay", "pays_base_fee": true}, {"id": "r-free", "pays_base_fee": false}})
	api.on("GET /api/clubs/club-1/players", 200, []map[string]any{
		{"id": "p-1", "name": "Anna", "role_id": "r-pay"},
		{"id": "p-2", "name": "Bernd", "role_id": "r-pay"},
		{"id": "p-3", "name": "Carla", "role_id": "r-pay", "inactive": true},
		{"id": "p-4", "name": "Dieter", "role_id": "r-free"},
	})
	return api
}

func TestCreateGameDay_BuildsDraftAndPreview(t *testing.T) {
	api := setupDraftAPI()
	out, isErr := callTool(t, connect(t, api), "spieltag_erfassen", map[string]any{
		"club_id": clubID, "datum": "2026-11-13", "notizen": "Heimspiel",
		"teilnehmer": []map[string]any{
			{"spieler_id": "p-1", "strafen": []map[string]any{{"strafentyp_id": "pt-1", "anzahl": 3}}, "wettbewerbe": []map[string]any{{"wettbewerb_id": "c-1", "wert": 7}}},
			{"spieler_id": "p-2", "strafen": []map[string]any{{"strafentyp_id": "pt-2", "anzahl": 0.5}}},
		},
	})
	if isErr {
		t.Fatal(out)
	}

	created := api.called("POST", "/api/clubs/club-1/gamedays")
	if len(created) != 1 || created[0].Body["is_draft"] != true || created[0].Body["date"] != "2026-11-13" {
		t.Fatalf("game day must be created as draft: %+v", created)
	}
	if n := len(api.called("POST", "/api/clubs/club-1/gamedays/gd-1/participants")); n != 2 {
		t.Errorf("expected 2 participants, got %d", n)
	}
	fees := api.called("PUT", "/api/clubs/club-1/gamedays/gd-1/participants/p-2/fees")
	if len(fees) != 1 || !strings.Contains(mustJSON(fees[0].Body), `"count":0.5`) {
		t.Errorf("decimal count not forwarded: %+v", fees)
	}
	if len(api.called("DELETE", "/api/clubs/club-1/gamedays/gd-1")) != 0 {
		t.Error("draft must not be deleted on success")
	}
	// Preview: penalties 3 × 0,50 € plus base fee for Anna and Bernd only
	// (Carla inactive, Dieter's role does not pay).
	for _, want := range []string{`"entwurf"`, `"1,50 €"`, `"anzahl_spieler": 2`, `"6,00 €"`, `"7,50 €"`, `"Pudelkönig"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %s:\n%s", want, out)
		}
	}
}

func TestCreateGameDay_RollsBackDraftOnFailure(t *testing.T) {
	api := setupDraftAPI()
	api.on("PUT /api/clubs/club-1/gamedays/gd-1/participants/p-2/fees", 400, map[string]any{"message": "Strafentyp nicht gefunden"})

	out, isErr := callTool(t, connect(t, api), "spieltag_erfassen", map[string]any{
		"club_id": clubID, "datum": "2026-11-13",
		"teilnehmer": []map[string]any{
			{"spieler_id": "p-1"},
			{"spieler_id": "p-2", "strafen": []map[string]any{{"strafentyp_id": "pt-1", "anzahl": 1}}},
		},
	})
	if !isErr || !strings.Contains(out, "Strafentyp nicht gefunden") || !strings.Contains(out, "verworfen") {
		t.Fatalf("expected rollback error, got %v: %s", isErr, out)
	}
	if len(api.called("DELETE", "/api/clubs/club-1/gamedays/gd-1")) != 1 {
		t.Error("half-filled draft must be deleted")
	}
}

func TestCreateGameDay_Validation(t *testing.T) {
	api := setupDraftAPI()
	s := connect(t, api)
	cases := map[string]map[string]any{
		"no participants":         {"club_id": clubID, "datum": "2026-11-13", "teilnehmer": []any{}},
		"bad date":                {"club_id": clubID, "datum": "13.11.2026", "teilnehmer": []map[string]any{{"spieler_id": "p-1"}}},
		"duplicate":               {"club_id": clubID, "datum": "2026-11-13", "teilnehmer": []map[string]any{{"spieler_id": "p-1"}, {"spieler_id": "p-1"}}},
		"decimal on integer type": {"club_id": clubID, "datum": "2026-11-13", "teilnehmer": []map[string]any{{"spieler_id": "p-1", "strafen": []map[string]any{{"strafentyp_id": "pt-1", "anzahl": 1.5}}}}},
		"three decimals":          {"club_id": clubID, "datum": "2026-11-13", "teilnehmer": []map[string]any{{"spieler_id": "p-1", "strafen": []map[string]any{{"strafentyp_id": "pt-2", "anzahl": 1.255}}}}},
		"unknown penalty type":    {"club_id": clubID, "datum": "2026-11-13", "teilnehmer": []map[string]any{{"spieler_id": "p-1", "strafen": []map[string]any{{"strafentyp_id": "pt-x", "anzahl": 1}}}}},
		"zero penalty":            {"club_id": clubID, "datum": "2026-11-13", "teilnehmer": []map[string]any{{"spieler_id": "p-1", "strafen": []map[string]any{{"strafentyp_id": "pt-1", "anzahl": 0}}}}},
	}
	for name, args := range cases {
		if out, isErr := callTool(t, s, "spieltag_erfassen", args); !isErr {
			t.Errorf("%s: expected error, got %s", name, out)
		}
	}
	if len(api.called("POST", "/api/clubs/club-1/gamedays")) != 0 {
		t.Error("invalid input must not reach the API")
	}
}

func TestPublishGameDay(t *testing.T) {
	api := setupDraftAPI()
	api.on("PUT /api/clubs/club-1/gamedays/gd-1", 200, map[string]any{})
	api.on("GET /api/clubs/club-1/gamedays/gd-1/transaction-summary", 200, map[string]any{"base_fee_total": -600, "penalty_fee_total": -150, "total": -750})

	_, isErr := callTool(t, connect(t, api), "spieltag_veroeffentlichen", map[string]any{"club_id": clubID, "spieltag_id": "gd-1"})
	if isErr {
		t.Fatal("publish failed")
	}
	puts := api.called("PUT", "/api/clubs/club-1/gamedays/gd-1")
	if len(puts) != 1 || puts[0].Body["is_draft"] != false || puts[0].Body["date"] != "2026-11-13" || puts[0].Body["notes"] != "Heimspiel" {
		t.Fatalf("unexpected publish request: %+v", puts)
	}
}

func TestPublishGameDay_AlreadyPublishedBooksNothing(t *testing.T) {
	api := newFakeAPI()
	api.on("GET /api/clubs/club-1/gamedays/gd-1", 200, map[string]any{"game_day": map[string]any{"id": "gd-1", "date": "2026-11-13", "is_draft": false}})

	out, isErr := callTool(t, connect(t, api), "spieltag_veroeffentlichen", map[string]any{"club_id": clubID, "spieltag_id": "gd-1"})
	if !isErr || !strings.Contains(out, "bereits veröffentlicht") {
		t.Fatalf("expected error, got %s", out)
	}
	if len(api.called("PUT", "/api/clubs/club-1/gamedays/gd-1")) != 0 {
		t.Error("published game day must not be updated again")
	}
}

func TestCreateTransaction(t *testing.T) {
	api := newFakeAPI()
	after := 1500
	api.on("POST /api/clubs/club-1/transactions", 201, map[string]any{
		"id": "tx-1", "player_id": "p-1", "player_name": "Anna", "transaction_type": "deposit", "amount": 2000,
		"player_balance_after": after, "club_balance_after": 50000, "transaction_date": "2026-11-17",
	})
	s := connect(t, api)

	out, isErr := callTool(t, s, "buchung_erfassen", map[string]any{"club_id": clubID, "typ": "einzahlung", "betrag_cent": 2000, "spieler_id": "p-1"})
	if isErr {
		t.Fatal(out)
	}
	posted := api.called("POST", "/api/clubs/club-1/transactions")
	if len(posted) != 1 || posted[0].Body["transaction_type"] != "deposit" || posted[0].Body["amount"] != float64(2000) || posted[0].Body["player_id"] != "p-1" {
		t.Fatalf("unexpected request: %+v", posted)
	}
	for _, want := range []string{`"einzahlung"`, `"20,00 €"`, `"15,00 €"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %s:\n%s", want, out)
		}
	}

	for name, args := range map[string]map[string]any{
		"fee type":           {"club_id": clubID, "typ": "strafe", "betrag_cent": 100, "spieler_id": "p-1"},
		"negative":           {"club_id": clubID, "typ": "einzahlung", "betrag_cent": -100, "spieler_id": "p-1"},
		"expense w/ player":  {"club_id": clubID, "typ": "ausgabe", "betrag_cent": 100, "spieler_id": "p-1"},
		"both player fields": {"club_id": clubID, "typ": "einzahlung", "betrag_cent": 100, "spieler_id": "p-1", "spieler_ids": []string{"p-1", "p-2"}},
	} {
		if out, isErr := callTool(t, s, "buchung_erfassen", args); !isErr {
			t.Errorf("%s: expected error, got %s", name, out)
		}
	}
	if n := len(api.called("POST", "/api/clubs/club-1/transactions")); n != 1 {
		t.Errorf("invalid bookings must not reach the API (got %d calls)", n)
	}
}

func TestCreateTransaction_AutoTipSplit(t *testing.T) {
	api := newFakeAPI()
	api.on("POST /api/clubs/club-1/transactions", 201, []map[string]any{
		{"id": "tx-1", "player_name": "Anna", "transaction_type": "deposit", "amount": 450, "club_balance_after": 450},
		{"id": "tx-2", "player_name": "Anna", "transaction_type": "tip", "amount": -550, "club_balance_after": 1000},
	})
	out, isErr := callTool(t, connect(t, api), "buchung_erfassen", map[string]any{"club_id": clubID, "typ": "einzahlung", "betrag_cent": 1000, "spieler_id": "p-1"})
	if isErr {
		t.Fatalf("split deposit must be reported as success: %s", out)
	}
	for _, want := range []string{`"tx-1"`, `"tx-2"`, `"trinkgeld"`, "aufgeteilt"} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %s:\n%s", want, out)
		}
	}
}

func TestAPIErrorsAreShownToModel(t *testing.T) {
	api := newFakeAPI()
	api.on("GET /api/clubs/club-1", 403, map[string]any{"message": "Keine Berechtigung"})
	out, isErr := callTool(t, connect(t, api), "klub_ueberblick", map[string]any{"club_id": clubID})
	if !isErr || !strings.Contains(out, "Keine Berechtigung") {
		t.Fatalf("expected permission error, got %s", out)
	}
}

func TestEuro(t *testing.T) {
	for cents, want := range map[int]string{0: "0,00 €", 5: "0,05 €", -350: "-3,50 €", 123456789: "1.234.567,89 €", -100000: "-1.000,00 €"} {
		if got := euro(cents); got != want {
			t.Errorf("euro(%d) = %q, want %q", cents, got, want)
		}
	}
}

func mustJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func TestMyStatistics(t *testing.T) {
	api := newFakeAPI()
	api.on("GET /api/clubs/club-1/players/me/statistics", 200, map[string]any{
		"player":    map[string]any{"id": "p-1", "name": "Anna"},
		"penalties": map[string]any{"total_amount": 1250, "rank": 2, "rank_of": 8},
	})
	s := connect(t, api)
	out, isErr := callTool(t, s, "meine_statistik", map[string]any{"club_id": clubID, "von": "2026-01-01", "bis": "2026-06-30"})
	if isErr {
		t.Fatal(out)
	}
	if !strings.Contains(out, `"total_amount": 1250`) {
		t.Errorf("output misses statistics:\n%s", out)
	}
	calls := api.called("GET", "/api/clubs/club-1/players/me/statistics")
	if len(calls) != 1 || calls[0].Query.Get("from") != "2026-01-01" || calls[0].Query.Get("to") != "2026-06-30" {
		t.Fatalf("unexpected calls: %+v", calls)
	}

	out, isErr = callTool(t, s, "meine_statistik", map[string]any{"club_id": clubID, "von": "01.01.2026"})
	if !isErr || !strings.Contains(out, "JJJJ-MM-TT") {
		t.Errorf("expected date validation error, got %s", out)
	}
}
