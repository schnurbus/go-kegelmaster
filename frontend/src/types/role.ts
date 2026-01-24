export type EntityType = "roles" | "players";
export type PermissionType = "list" | "view" | "create" | "update" | "delete";

export type Permission = {
  id: string;
  role_id: string;
  entity_type: EntityType;
  permission_type: PermissionType;
  created_at: string;
};

export type Role = {
  id: string;
  club_id: string;
  name: string;
  pays_base_fee: boolean;
  created_at: string;
  updated_at: string;
  permissions: Permission[];
};

export type CreateRoleRequest = {
  name: string;
  pays_base_fee: boolean;
};

export type UpdateRoleRequest = {
  name: string;
  pays_base_fee: boolean;
};

// Helper function to get German label for entity type
export function getEntityTypeLabel(entityType: EntityType): string {
  const labels: Record<EntityType, string> = {
    roles: "Rollen",
    players: "Spieler",
  };
  return labels[entityType];
}

// Helper function to get German label for permission type
export function getPermissionTypeLabel(permissionType: PermissionType): string {
  const labels: Record<PermissionType, string> = {
    list: "Auflisten",
    view: "Ansehen",
    create: "Erstellen",
    update: "Bearbeiten",
    delete: "Löschen",
  };
  return labels[permissionType];
}

// Helper function to check if a role has a specific permission
export function hasPermission(
  role: Role,
  entityType: EntityType,
  permissionType: PermissionType
): boolean {
  return role.permissions.some(
    (p) =>
      p.entity_type === entityType && p.permission_type === permissionType
  );
}

