package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var playerColumns = []string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}

func expectStatsUser(mock sqlmock.Sqlmock, userID string, now time.Time) {
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at", "password_changed_at"}).
			AddRow(userID, "user@example.com", "hash", now, now, nil))
}

func TestHandleGetMyStatistics_Unauthorized(t *testing.T) {
	srv, _, _ := newTestServer(t)
	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+uuid.NewString()+"/players/me/statistics", "")
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHandleGetMyStatistics_InvalidPeriod(t *testing.T) {
	cases := map[string]string{
		"bad from":      "?from=01.01.2026",
		"bad to":        "?to=foo",
		"from after to": "?from=2026-05-01&to=2026-01-01",
	}
	for name, query := range cases {
		t.Run(name, func(t *testing.T) {
			srv, mock, authSvc := newTestServer(t)
			userID := uuid.NewString()
			token, err := authSvc.GenerateToken(userID, true)
			if err != nil {
				t.Fatalf("generate token: %v", err)
			}
			expectStatsUser(mock, userID, time.Now().UTC())

			req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+uuid.NewString()+"/players/me/statistics"+query, "")
			req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
			resp := doRequest(t, srv, req)
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.StatusCode)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestHandleGetMyStatistics_NoPlayer(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	clubID := uuid.NewString()
	token, _ := authSvc.GenerateToken(userID, true)
	expectStatsUser(mock, userID, time.Now().UTC())
	mock.ExpectQuery(`SELECT .* FROM players\s+WHERE user_id = \$1 AND club_id = \$2`).
		WithArgs(&userID, clubID).
		WillReturnRows(sqlmock.NewRows(playerColumns))

	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+clubID+"/players/me/statistics", "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleGetMyStatistics_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
	otherID := uuid.NewString()
	compID := uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	gd := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)

	expectStatsUser(mock, userID, now)
	mock.ExpectQuery(`SELECT .* FROM players\s+WHERE user_id = \$1 AND club_id = \$2`).
		WithArgs(&userID, clubID).
		WillReturnRows(sqlmock.NewRows(playerColumns).
			AddRow(playerID, clubID, userID, nil, "Hans", 0, 1000, "male", false, nil, now, now))
	mock.ExpectQuery(`FROM game_days gd\s+WHERE gd.club_id = \$1 AND gd.is_draft = false`).
		WithArgs(clubID, from, to).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date"}).AddRow("gd1", gd))
	mock.ExpectQuery(`FROM game_day_participants gdp`).
		WithArgs(clubID, from, to).
		WillReturnRows(sqlmock.NewRows([]string{"participant_id", "game_day_id", "player_id", "attended", "gender"}).
			AddRow("a1", "gd1", playerID, true, "male").
			AddRow("a2", "gd1", otherID, true, "male"))
	mock.ExpectQuery(`FROM game_day_fees gdf`).
		WithArgs(clubID, from, to).
		WillReturnRows(sqlmock.NewRows([]string{"game_day_participant_id", "penalty_type_id", "penalty_type_name", "penalty_type_price", "count", "quantity_scale"}).
			AddRow("a1", "pt1", "Pudel", 50, 4, 1))
	mock.ExpectQuery(`FROM game_day_competition_values gdcv`).
		WithArgs(clubID, from, to).
		WillReturnRows(sqlmock.NewRows([]string{"game_day_participant_id", "competition_id", "value"}).
			AddRow("a1", compID, 9).
			AddRow("a2", compID, 3))
	mock.ExpectQuery(`SELECT .* FROM competitions\s+WHERE club_id = \$1`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "scoring_type", "is_gender_specific", "display_order", "created_at", "updated_at"}).
			AddRow(compID, clubID, "Pokal", "winner", false, 0, now, now))
	mock.ExpectQuery(`transaction_date < \$2`).
		WithArgs(&playerID, from).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(int64(-300)))
	mock.ExpectQuery(`SELECT transaction_date, transaction_type, amount`).
		WithArgs(&playerID, from, to).
		WillReturnRows(sqlmock.NewRows([]string{"transaction_date", "transaction_type", "amount"}).
			AddRow(gd, "fee", -200).
			AddRow(gd, "deposit", 1000))

	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+clubID+"/players/me/statistics?from=2026-01-01&to=2026-06-30", "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload struct {
		Player     struct{ Name string } `json:"player"`
		Attendance struct {
			Attended int     `json:"attended"`
			Rate     float64 `json:"rate"`
		} `json:"attendance"`
		Penalties struct {
			TotalAmount int `json:"total_amount"`
			Rank        int `json:"rank"`
		} `json:"penalties"`
		Competitions []struct {
			Name string `json:"name"`
			Wins int    `json:"wins"`
		} `json:"competitions"`
		Finance struct {
			StartBalance int `json:"start_balance"`
			EndBalance   int `json:"end_balance"`
		} `json:"finance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Player.Name != "Hans" || payload.Attendance.Attended != 1 || payload.Attendance.Rate != 1 {
		t.Fatalf("unexpected player/attendance: %+v", payload)
	}
	if payload.Penalties.TotalAmount != 200 || payload.Penalties.Rank != 1 {
		t.Fatalf("unexpected penalties: %+v", payload.Penalties)
	}
	if len(payload.Competitions) != 1 || payload.Competitions[0].Wins != 1 {
		t.Fatalf("unexpected competitions: %+v", payload.Competitions)
	}
	if payload.Finance.StartBalance != 700 || payload.Finance.EndBalance != 1500 {
		t.Fatalf("unexpected finance: %+v", payload.Finance)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
