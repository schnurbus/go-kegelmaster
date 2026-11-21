package club

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestRepository_SetupCounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	clubID, userID := uuid.NewString(), uuid.NewString()

	// Drafts must not count as a game day.
	mock.ExpectQuery(`NOT g\.is_draft`).WithArgs(clubID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"penalty_types", "roles", "players", "published_game_days", "invitations", "competitions", "own_players"}).
			AddRow(1, 2, 3, 4, 5, 6, 0))

	got, err := NewRepository(db).SetupCounts(context.Background(), clubID, userID)
	if err != nil {
		t.Fatalf("setup counts: %v", err)
	}
	want := SetupCounts{PenaltyTypes: 1, Roles: 2, Players: 3, PublishedGameDays: 4, Invitations: 5, Competitions: 6, OwnPlayers: 0}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_DismissSetup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	clubID := uuid.NewString()

	mock.ExpectExec(`SET setup_dismissed_at = COALESCE\(setup_dismissed_at`).WithArgs(sqlmock.AnyArg(), clubID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := NewRepository(db).DismissSetup(context.Background(), clubID); err != nil {
		t.Fatalf("dismiss: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByID_SetupDismissed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	clubID, userID, now := uuid.NewString(), uuid.NewString(), time.Now().UTC()

	mock.ExpectQuery(`FROM clubs`).WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at", "setup_dismissed_at"}).
			AddRow(clubID, "KC", 0, 0, 0, true, false, userID, now, now, now))

	c, err := NewRepository(db).GetByID(context.Background(), clubID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if c.SetupDismissedAt == nil {
		t.Fatal("expected SetupDismissedAt to be set")
	}
}
