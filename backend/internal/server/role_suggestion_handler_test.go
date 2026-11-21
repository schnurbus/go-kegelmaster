package server

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// recordArg is a sqlmock argument matcher that accepts any value and records it.
type recordArg struct{ into *string }

func (r recordArg) Match(v driver.Value) bool {
	if s, ok := v.(string); ok {
		*r.into = s
	}
	return true
}

type grantedPermission struct{ roleID, entity, perm string }

// permissionRecorder records the role, entity and permission of every inserted permission.
type permissionRecorder struct {
	granted []grantedPermission
	cur     grantedPermission
}

func (p *permissionRecorder) arg(field string) sqlmock.Argument {
	return matcherFunc(func(v driver.Value) bool {
		s, _ := v.(string)
		switch field {
		case "role":
			p.cur = grantedPermission{roleID: s}
		case "entity":
			p.cur.entity = s
		case "perm":
			p.cur.perm = s
			p.granted = append(p.granted, p.cur)
		}
		return true
	})
}

type matcherFunc func(driver.Value) bool

func (f matcherFunc) Match(v driver.Value) bool { return f(v) }

func suggestionRequest(t *testing.T, token, clubID string) *http.Request {
	t.Helper()
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/roles/suggestion", "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	return req
}

func TestRoleSuggestion_CreatesFourRoles(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	mock.ExpectQuery(`SELECT .* FROM roles\s+WHERE club_id`).WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows(roleCols))
	mock.ExpectBegin()

	names := make([]string, 4)
	fees := make([]bool, 4)
	roleIDs := make([]string, 4)
	rec := &permissionRecorder{}
	// Enough permission inserts per role; the exact matrix is checked below via the recorder.
	for i := range 4 {
		roleIDs[i] = uuid.NewString()
		i := i
		mock.ExpectQuery(`INSERT INTO roles`).
			WithArgs(sqlmock.AnyArg(), clubID, recordArg{&names[i]}, matcherFunc(func(v driver.Value) bool { fees[i], _ = v.(bool); return true }), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows(roleCols).AddRow(roleIDs[i], clubID, "", false, now, now))
		for range 40 {
			mock.ExpectQuery(`INSERT INTO role_permissions`).
				WithArgs(sqlmock.AnyArg(), rec.arg("role"), rec.arg("entity"), rec.arg("perm"), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id", "role_id", "entity_type", "permission_type", "created_at"}).
					AddRow(uuid.NewString(), roleIDs[i], "players", "list", now))
		}
	}
	mock.ExpectCommit()
	mock.MatchExpectationsInOrder(false)

	resp := doRequest(t, srv, suggestionRequest(t, token, clubID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var roles []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(roles) != 4 {
		t.Fatalf("expected 4 roles, got %d", len(roles))
	}

	nameByID := map[string]string{}
	feeByName := map[string]bool{}
	for i := range 4 {
		nameByID[roleIDs[i]] = names[i]
		feeByName[names[i]] = fees[i]
	}
	for _, want := range []string{"Präsident", "Kassenwart", "Mitglied", "Gast"} {
		if _, ok := feeByName[want]; !ok {
			t.Fatalf("role %s not created, got %v", want, names)
		}
	}
	if feeByName["Gast"] || !feeByName["Mitglied"] {
		t.Fatalf("Gast must not pay the base fee, Mitglied must: %v", feeByName)
	}

	has := func(roleName, entity, perm string) bool {
		for _, g := range rec.granted {
			if nameByID[g.roleID] == roleName && g.entity == entity && g.perm == perm {
				return true
			}
		}
		return false
	}
	checks := []struct {
		role, entity, perm string
		want               bool
	}{
		{"Präsident", "roles", "delete", true},
		{"Präsident", "transactions", "create", true},
		{"Kassenwart", "transactions", "create", true},
		{"Kassenwart", "game_days", "update", true},
		{"Kassenwart", "roles", "view", true},
		{"Kassenwart", "roles", "create", false},
		{"Mitglied", "game_days", "view", true},
		{"Mitglied", "game_days", "create", false},
		{"Mitglied", "roles", "list", false},
		{"Gast", "transactions", "list", true},
		{"Gast", "players", "update", false},
	}
	for _, c := range checks {
		if got := has(c.role, c.entity, c.perm); got != c.want {
			t.Errorf("%s %s/%s: got %v, want %v", c.role, c.entity, c.perm, got, c.want)
		}
	}
}

func TestRoleSuggestion_ConflictWhenRolesExist(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	mock.ExpectQuery(`SELECT .* FROM roles\s+WHERE club_id`).WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows(roleCols).AddRow(uuid.NewString(), clubID, "Mitglied", true, now, now))

	resp := doRequest(t, srv, suggestionRequest(t, token, clubID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRoleSuggestion_ForbiddenWithoutPermission(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, ownerID, clubID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, ownerID, nil, now)
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name`).WithArgs(userID, clubID).WillReturnError(sql.ErrNoRows)

	resp := doRequest(t, srv, suggestionRequest(t, token, clubID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
