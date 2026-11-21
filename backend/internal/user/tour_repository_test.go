package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var userRowCols = []string{"id", "email", "password_hash", "created_at", "updated_at", "password_changed_at", "notify_gameday_settled", "is_instance_admin", "last_login_at", "disabled_at", "sessions_valid_after", "tour_seen_at", "setup_hint_seen_at"}

func TestRepository_MarkTourFlags(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	userID, now := uuid.NewString(), time.Now().UTC()

	// An existing timestamp is kept (COALESCE), only the requested flags change.
	mock.ExpectQuery(`COALESCE\(tour_seen_at`).WithArgs(true, sqlmock.AnyArg(), false, userID).
		WillReturnRows(sqlmock.NewRows(userRowCols).AddRow(userID, "a@example.com", "hash", now, now, nil, false, false, nil, nil, nil, now, nil))

	u, err := NewRepository(db).MarkTourFlags(context.Background(), userID, true, false)
	if err != nil {
		t.Fatalf("mark: %v", err)
	}
	if u.TourSeenAt == nil || u.SetupHintSeenAt != nil {
		t.Fatalf("unexpected flags: tour=%v hint=%v", u.TourSeenAt, u.SetupHintSeenAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_MarkTourFlags_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(`UPDATE users`).WillReturnError(sql.ErrNoRows)

	if _, err := NewRepository(db).MarkTourFlags(context.Background(), uuid.NewString(), true, true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
