package role

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

var (
	ErrNotFound        = errors.New("role not found")
	ErrPermissionExist = errors.New("permission already exists")
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{
		queries: db.New(dbConn),
	}
}

type CreateRoleParams struct {
	ClubID      string
	Name        string
	PaysBaseFee bool
}

func (r *Repository) Create(ctx context.Context, params CreateRoleParams) (Role, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbRole, err := r.queries.CreateRole(ctx, db.CreateRoleParams{
		ID:          id,
		ClubID:      params.ClubID,
		Name:        params.Name,
		PaysBaseFee: params.PaysBaseFee,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return Role{}, err
	}

	return dbRoleToRole(dbRole), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Role, error) {
	dbRole, err := r.queries.GetRoleByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Role{}, ErrNotFound
	}
	if err != nil {
		return Role{}, err
	}

	return dbRoleToRole(dbRole), nil
}

func (r *Repository) GetByClubID(ctx context.Context, clubID string) ([]Role, error) {
	dbRoles, err := r.queries.GetRolesByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	roles := make([]Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = dbRoleToRole(dbRole)
	}

	return roles, nil
}

type UpdateRoleParams struct {
	ID          string
	Name        string
	PaysBaseFee bool
}

func (r *Repository) Update(ctx context.Context, params UpdateRoleParams) (Role, error) {
	now := time.Now().UTC()

	dbRole, err := r.queries.UpdateRole(ctx, db.UpdateRoleParams{
		ID:          params.ID,
		Name:        params.Name,
		PaysBaseFee: params.PaysBaseFee,
		UpdatedAt:   now,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Role{}, ErrNotFound
	}
	if err != nil {
		return Role{}, err
	}

	return dbRoleToRole(dbRole), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	// Check if role exists first
	_, err := r.queries.GetRoleByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Delete the role
	err = r.queries.DeleteRole(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

type AddPermissionParams struct {
	RoleID         string
	EntityType     EntityType
	PermissionType PermissionType
}

func (r *Repository) AddPermission(ctx context.Context, params AddPermissionParams) (RolePermission, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbPerm, err := r.queries.AddRolePermission(ctx, db.AddRolePermissionParams{
		ID:             id,
		RoleID:         params.RoleID,
		EntityType:     string(params.EntityType),
		PermissionType: string(params.PermissionType),
		CreatedAt:      now,
	})
	if err != nil {
		// Check if it's a unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return RolePermission{}, ErrPermissionExist
		}
		// PostgreSQL unique constraint violation (check error message)
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") {
			return RolePermission{}, ErrPermissionExist
		}
		return RolePermission{}, err
	}

	return dbRolePermissionToRolePermission(dbPerm), nil
}

type RemovePermissionParams struct {
	RoleID         string
	EntityType     EntityType
	PermissionType PermissionType
}

func (r *Repository) RemovePermission(ctx context.Context, params RemovePermissionParams) error {
	// Check if permission exists first
	hasPermission, err := r.HasPermission(ctx, params.RoleID, params.EntityType, params.PermissionType)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrNotFound
	}

	// Remove the permission
	err = r.queries.RemoveRolePermission(ctx, db.RemoveRolePermissionParams{
		RoleID:         params.RoleID,
		EntityType:     string(params.EntityType),
		PermissionType: string(params.PermissionType),
	})
	return err
}

func (r *Repository) GetPermissions(ctx context.Context, roleID string) ([]RolePermission, error) {
	dbPerms, err := r.queries.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permissions := make([]RolePermission, len(dbPerms))
	for i, dbPerm := range dbPerms {
		permissions[i] = dbRolePermissionToRolePermission(dbPerm)
	}

	return permissions, nil
}

func (r *Repository) HasPermission(ctx context.Context, roleID string, entityType EntityType, permissionType PermissionType) (bool, error) {
	return r.queries.HasRolePermission(ctx, db.HasRolePermissionParams{
		RoleID:         roleID,
		EntityType:     string(entityType),
		PermissionType: string(permissionType),
	})
}

// dbRoleToRole converts a db.Role to a role.Role
func dbRoleToRole(dbRole db.Role) Role {
	return Role{
		ID:          dbRole.ID,
		ClubID:      dbRole.ClubID,
		Name:        dbRole.Name,
		PaysBaseFee: dbRole.PaysBaseFee,
		CreatedAt:   dbRole.CreatedAt,
		UpdatedAt:   dbRole.UpdatedAt,
	}
}

// dbRolePermissionToRolePermission converts a db.RolePermission to a role.RolePermission
func dbRolePermissionToRolePermission(dbPerm db.RolePermission) RolePermission {
	return RolePermission{
		ID:             dbPerm.ID,
		RoleID:         dbPerm.RoleID,
		EntityType:     EntityType(dbPerm.EntityType),
		PermissionType: PermissionType(dbPerm.PermissionType),
		CreatedAt:      dbPerm.CreatedAt,
	}
}

