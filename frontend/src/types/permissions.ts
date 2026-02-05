/**
 * Response from GET /api/clubs/:clubId/permissions/me
 */
export type MyPermissionsResponse = {
  is_owner: boolean;
  permissions: Array<{
    entity_type: string;
    permission_type: string;
  }>;
};

export type EntityType =
  | "roles"
  | "players"
  | "game_nights"
  | "game_days"
  | "penalties"
  | "penalty_types"
  | "competitions"
  | "transactions";

export type PermissionType = "list" | "view" | "create" | "update" | "delete";

export type PermissionsState = {
  isOwner: boolean;
  permissions: MyPermissionsResponse["permissions"];
  isLoading: boolean;
  error: string | null;
};

export type UsePermissionsReturn = PermissionsState & {
  canCreate: (entityType: EntityType) => boolean;
  canUpdate: (entityType: EntityType) => boolean;
  canDelete: (entityType: EntityType) => boolean;
  canList: (entityType: EntityType) => boolean;
  canView: (entityType: EntityType) => boolean;
};
