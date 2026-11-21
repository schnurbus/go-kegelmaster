package role

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var roleRowCols = []string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}
var permRowCols = []string{"id", "role_id", "entity_type", "permission_type", "created_at"}

func TestRepository_Delete_InUse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	roleID, clubID, now := uuid.NewString(), uuid.NewString(), time.Now().UTC()

	mock.ExpectQuery(`FROM roles`).WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows(roleRowCols).AddRow(roleID, clubID, "Mitglied", true, now, now))
	mock.ExpectExec(`DELETE FROM roles`).WithArgs(roleID).
		WillReturnError(&pgconn.PgError{Code: "23503"})

	if err := NewRepository(db).Delete(context.Background(), roleID); !errors.Is(err, ErrInUse) {
		t.Fatalf("expected ErrInUse, got %v", err)
	}
}

func TestRepository_CreateFromTemplates_Commits(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	clubID, roleID, now := uuid.NewString(), uuid.NewString(), time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO roles`).WithArgs(sqlmock.AnyArg(), clubID, "Gast", false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(roleRowCols).AddRow(roleID, clubID, "Gast", false, now, now))
	mock.ExpectQuery(`INSERT INTO role_permissions`).WithArgs(sqlmock.AnyArg(), roleID, "players", "list", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(permRowCols).AddRow(uuid.NewString(), roleID, "players", "list", now))
	mock.ExpectCommit()

	created, err := NewRepository(db).CreateFromTemplates(context.Background(), clubID, []Template{
		{Name: "Gast", Permissions: []TemplatePermission{{EntityType: EntityTypePlayers, PermissionType: PermissionTypeList}}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(created) != 1 || created[0].Role.Name != "Gast" || len(created[0].Permissions) != 1 {
		t.Fatalf("unexpected result: %+v", created)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_CreateFromTemplates_RollsBackOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	clubID, roleID, now := uuid.NewString(), uuid.NewString(), time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO roles`).
		WillReturnRows(sqlmock.NewRows(roleRowCols).AddRow(roleID, clubID, "Gast", false, now, now))
	mock.ExpectQuery(`INSERT INTO role_permissions`).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	_, err = NewRepository(db).CreateFromTemplates(context.Background(), clubID, []Template{
		{Name: "Gast", Permissions: []TemplatePermission{{EntityType: EntityTypePlayers, PermissionType: PermissionTypeList}}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
