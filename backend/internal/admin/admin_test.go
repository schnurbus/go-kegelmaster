package admin

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseUserListQuery_Defaults(t *testing.T) {
	q, err := ParseUserListQuery(ListParams{})
	if err != nil {
		t.Fatal(err)
	}
	if q.Sort != "created_at" || q.Order != OrderDesc || q.Page != 1 || q.Limit != DefaultPageLimit || q.Filter != "" {
		t.Errorf("unexpected defaults: %+v", q)
	}
	q, err = ParseUserListQuery(ListParams{Sort: "email"})
	if err != nil || q.Order != OrderAsc {
		t.Errorf("email should default to asc: %+v %v", q, err)
	}
}

func TestParseUserListQuery_Invalid(t *testing.T) {
	cases := []ListParams{
		{Filter: "inactive"},
		{Sort: "password_hash"},
		{Order: "up"},
		{Page: "0"},
		{Page: "x"},
		{Limit: "101"},
		{Search: strings.Repeat("a", 201)},
	}
	for _, p := range cases {
		if _, err := ParseUserListQuery(p); !errors.As(err, new(ValidationError)) {
			t.Errorf("%+v: expected ValidationError, got %v", p, err)
		}
	}
}

func TestParseClubListQuery(t *testing.T) {
	q, err := ParseClubListQuery(ListParams{Search: "  KC ", Filter: "INACTIVE", Sort: "last_activity_at", Page: "3", Limit: "50"})
	if err != nil {
		t.Fatal(err)
	}
	if q.Search != "KC" || q.Filter != ClubFilterInactive || q.Sort != "last_activity_at" || q.Order != OrderDesc || q.Page != 3 || q.Limit != 50 {
		t.Errorf("unexpected query: %+v", q)
	}
	if _, err := ParseClubListQuery(ListParams{Filter: "orphaned"}); err == nil {
		t.Error("orphaned is no club filter")
	}
}

func TestEscapeLike(t *testing.T) {
	if got := escapeLike(`50%_a\b`); got != `50\%\_a\\b` {
		t.Errorf("got %q", got)
	}
}

func TestNewPage(t *testing.T) {
	p := newPage([]int{1}, 1, 25, 51)
	if p.TotalPages != 3 {
		t.Errorf("total pages: got %d", p.TotalPages)
	}
	if p := newPage([]int{}, 1, 25, 0); p.TotalPages != 0 {
		t.Errorf("empty: got %d", p.TotalPages)
	}
}

func TestAnyTime(t *testing.T) {
	if v, err := anyTime(nil); v != nil || err != nil {
		t.Errorf("nil: %v %v", v, err)
	}
	now := time.Now()
	if v, err := anyTime(now); err != nil || !v.Equal(now) {
		t.Errorf("time: %v %v", v, err)
	}
	if _, err := anyTime("2026-01-01"); err == nil {
		t.Error("expected error for string")
	}
}

type fakeStore struct {
	admins map[string]bool // email -> is admin; missing key = no account
	calls  []string
}

func (f *fakeStore) SetInstanceAdmin(_ context.Context, email string, isAdmin bool) error {
	f.calls = append(f.calls, email)
	if _, ok := f.admins[email]; !ok {
		return ErrUserNotFound
	}
	f.admins[email] = isAdmin
	return nil
}

func (f *fakeStore) ListInstanceAdmins(context.Context) ([]InstanceAdmin, error) {
	var out []InstanceAdmin
	for email, isAdmin := range f.admins {
		if isAdmin {
			out = append(out, InstanceAdmin{ID: "id", Email: email, CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)})
		}
	}
	return out, nil
}

func runCLI(store AdminStore, args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := RunCLI(context.Background(), store, args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestRunCLI_GrantRevokeList(t *testing.T) {
	store := &fakeStore{admins: map[string]bool{"op@example.com": false}}

	if code, out, _ := runCLI(store, "grant", " OP@example.com "); code != 0 || !strings.Contains(out, "op@example.com ist jetzt Instanz-Admin") {
		t.Fatalf("grant: code=%d out=%q", code, out)
	}
	if !store.admins["op@example.com"] {
		t.Fatal("flag not set")
	}
	if code, out, _ := runCLI(store, "list"); code != 0 || !strings.Contains(out, "op@example.com") || !strings.Contains(out, "2026-01-02T03:04:05Z") {
		t.Fatalf("list: code=%d out=%q", code, out)
	}
	if code, _, _ := runCLI(store, "revoke", "op@example.com"); code != 0 || store.admins["op@example.com"] {
		t.Fatalf("revoke: code=%d", code)
	}
	if code, out, _ := runCLI(store, "list"); code != 0 || !strings.Contains(out, "Keine Instanz-Admins") {
		t.Fatalf("empty list: code=%d out=%q", code, out)
	}
}

func TestRunCLI_Errors(t *testing.T) {
	store := &fakeStore{admins: map[string]bool{}}
	if code, _, errOut := runCLI(store, "grant", "missing@example.com"); code != 1 || !strings.Contains(errOut, "Kein Konto") {
		t.Errorf("unknown account: code=%d err=%q", code, errOut)
	}
	for _, args := range [][]string{{}, {"grant"}, {"grant", "no-email"}, {"list", "x"}, {"delete", "a@b.de"}} {
		if code, _, _ := runCLI(store, args...); code != 2 {
			t.Errorf("%v: expected exit code 2, got %d", args, code)
		}
	}
}

func TestBootstrapFromEnv_GrantsExistingOnly(t *testing.T) {
	store := &fakeStore{admins: map[string]bool{"a@example.com": false, "b@example.com": true}}
	BootstrapFromEnv(context.Background(), store, []string{" A@example.com", "", "missing@example.com"})
	if !store.admins["a@example.com"] {
		t.Error("a@example.com should be admin")
	}
	if !store.admins["b@example.com"] {
		t.Error("bootstrap must not revoke")
	}
	if len(store.calls) != 2 {
		t.Errorf("calls: %v", store.calls)
	}
}
