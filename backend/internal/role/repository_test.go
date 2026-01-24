package role

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`INSERT INTO roles`).
		WithArgs(sqlmock.AnyArg(), clubID, "Test Role", false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID, clubID, "Test Role", false, now, now))

	ctx := context.Background()
	role, err := repo.Create(ctx, CreateRoleParams{
		ClubID:      clubID,
		Name:        "Test Role",
		PaysBaseFee: false,
	})

	if err != nil {
		t.Fatalf("create role: %v", err)
	}

	if role.ID != roleID {
		t.Fatalf("expected id %s, got %s", roleID, role.ID)
	}
	if role.Name != "Test Role" {
		t.Fatalf("expected name 'Test Role', got %s", role.Name)
	}
	if role.ClubID != clubID {
		t.Fatalf("expected club_id %s, got %s", clubID, role.ClubID)
	}
	if role.PaysBaseFee != false {
		t.Fatalf("expected pays_base_fee false, got %v", role.PaysBaseFee)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee, created_at, updated_at`).
		WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID, clubID, "Test Role", true, now, now))

	ctx := context.Background()
	role, err := repo.GetByID(ctx, roleID)

	if err != nil {
		t.Fatalf("get role: %v", err)
	}

	if role.ID != roleID {
		t.Fatalf("expected id %s, got %s", roleID, role.ID)
	}
	if role.Name != "Test Role" {
		t.Fatalf("expected name 'Test Role', got %s", role.Name)
	}
	if role.PaysBaseFee != true {
		t.Fatalf("expected pays_base_fee true, got %v", role.PaysBaseFee)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee, created_at, updated_at`).
		WithArgs(roleID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = repo.GetByID(ctx, roleID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByClubID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID := uuid.NewString()
	roleID1 := uuid.NewString()
	roleID2 := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID1, clubID, "Role 1", false, now, now).
			AddRow(roleID2, clubID, "Role 2", true, now, now))

	ctx := context.Background()
	roles, err := repo.GetByClubID(ctx, clubID)

	if err != nil {
		t.Fatalf("get roles: %v", err)
	}

	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if roles[0].ID != roleID1 {
		t.Fatalf("expected first role id %s, got %s", roleID1, roles[0].ID)
	}
	if roles[1].ID != roleID2 {
		t.Fatalf("expected second role id %s, got %s", roleID2, roles[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`UPDATE roles`).
		WithArgs("Updated Role", true, sqlmock.AnyArg(), roleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID, clubID, "Updated Role", true, now, now))

	ctx := context.Background()
	role, err := repo.Update(ctx, UpdateRoleParams{
		ID:          roleID,
		Name:        "Updated Role",
		PaysBaseFee: true,
	})

	if err != nil {
		t.Fatalf("update role: %v", err)
	}

	if role.Name != "Updated Role" {
		t.Fatalf("expected name 'Updated Role', got %s", role.Name)
	}
	if role.PaysBaseFee != true {
		t.Fatalf("expected pays_base_fee true, got %v", role.PaysBaseFee)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	// First call: GetRoleByID to check if role exists
	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee, created_at, updated_at FROM roles`).
		WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID, clubID, "Test Role", false, now, now))

	// Second call: DeleteRole
	mock.ExpectExec(`DELETE FROM roles`).
		WithArgs(roleID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.Delete(ctx, roleID)

	if err != nil {
		t.Fatalf("delete role: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()

	// GetRoleByID returns ErrNoRows
	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee, created_at, updated_at FROM roles`).
		WithArgs(roleID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	err = repo.Delete(ctx, roleID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_AddPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	permID := uuid.NewString()
	roleID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`INSERT INTO role_permissions`).
		WithArgs(sqlmock.AnyArg(), roleID, EntityTypeRoles, PermissionTypeCreate, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role_id", "entity_type", "permission_type", "created_at"}).
			AddRow(permID, roleID, string(EntityTypeRoles), string(PermissionTypeCreate), now))

	ctx := context.Background()
	perm, err := repo.AddPermission(ctx, AddPermissionParams{
		RoleID:         roleID,
		EntityType:     EntityTypeRoles,
		PermissionType: PermissionTypeCreate,
	})

	if err != nil {
		t.Fatalf("add permission: %v", err)
	}

	if perm.ID != permID {
		t.Fatalf("expected id %s, got %s", permID, perm.ID)
	}
	if perm.RoleID != roleID {
		t.Fatalf("expected role_id %s, got %s", roleID, perm.RoleID)
	}
	if perm.EntityType != EntityTypeRoles {
		t.Fatalf("expected entity_type %s, got %s", EntityTypeRoles, perm.EntityType)
	}
	if perm.PermissionType != PermissionTypeCreate {
		t.Fatalf("expected permission_type %s, got %s", PermissionTypeCreate, perm.PermissionType)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_RemovePermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()

	// First call: HasRolePermission returns true
	mock.ExpectQuery(`-- name: HasRolePermission :one`).
		WithArgs(roleID, string(EntityTypeRoles), string(PermissionTypeCreate)).
		WillReturnRows(sqlmock.NewRows([]string{"column_1"}).AddRow(true))

	// Second call: RemoveRolePermission
	mock.ExpectExec(`DELETE FROM role_permissions`).
		WithArgs(roleID, string(EntityTypeRoles), string(PermissionTypeCreate)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.RemovePermission(ctx, RemovePermissionParams{
		RoleID:         roleID,
		EntityType:     EntityTypeRoles,
		PermissionType: PermissionTypeCreate,
	})

	if err != nil {
		t.Fatalf("remove permission: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()
	permID1 := uuid.NewString()
	permID2 := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, role_id, entity_type, permission_type, created_at`).
		WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role_id", "entity_type", "permission_type", "created_at"}).
			AddRow(permID1, roleID, string(EntityTypeRoles), string(PermissionTypeCreate), now).
			AddRow(permID2, roleID, string(EntityTypePlayers), string(PermissionTypeView), now))

	ctx := context.Background()
	perms, err := repo.GetPermissions(ctx, roleID)

	if err != nil {
		t.Fatalf("get permissions: %v", err)
	}

	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(perms))
	}
	if perms[0].ID != permID1 {
		t.Fatalf("expected first permission id %s, got %s", permID1, perms[0].ID)
	}
	if perms[1].ID != permID2 {
		t.Fatalf("expected second permission id %s, got %s", permID2, perms[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_HasPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	roleID := uuid.NewString()

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(roleID, EntityTypeRoles, PermissionTypeCreate).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(true))

	ctx := context.Background()
	has, err := repo.HasPermission(ctx, roleID, EntityTypeRoles, PermissionTypeCreate)

	if err != nil {
		t.Fatalf("has permission: %v", err)
	}

	if !has {
		t.Fatal("expected has permission to be true, got false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestEntityType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		entity   EntityType
		expected bool
	}{
		{"valid roles", EntityTypeRoles, true},
		{"valid players", EntityTypePlayers, true},
		{"valid game_nights", EntityTypeGameNights, true},
		{"valid penalties", EntityTypePenalties, true},
		{"invalid", EntityType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entity.IsValid(); got != tt.expected {
				t.Fatalf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPermissionType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		perm     PermissionType
		expected bool
	}{
		{"valid list", PermissionTypeList, true},
		{"valid view", PermissionTypeView, true},
		{"valid create", PermissionTypeCreate, true},
		{"valid update", PermissionTypeUpdate, true},
		{"valid delete", PermissionTypeDelete, true},
		{"invalid", PermissionType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.perm.IsValid(); got != tt.expected {
				t.Fatalf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

