"use client";

import { useCallback, useEffect, useState } from "react";
import type {
  EntityType,
  MyPermissionsResponse,
  UsePermissionsReturn,
} from "@/types/permissions";

function hasPerm(
  permissions: MyPermissionsResponse["permissions"],
  entityType: EntityType,
  permissionType: string
): boolean {
  return permissions.some(
    (p) =>
      p.entity_type === entityType && p.permission_type === permissionType
  );
}

export function usePermissions(clubId: string | null): UsePermissionsReturn {
  const [isOwner, setIsOwner] = useState(false);
  const [permissions, setPermissions] = useState<
    MyPermissionsResponse["permissions"]
  >([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!clubId) {
      setIsOwner(false);
      setPermissions([]);
      setIsLoading(false);
      setError(null);
      return;
    }

    let cancelled = false;
    setError(null);
    setIsLoading(true);

    fetch(`/api/clubs/${clubId}/permissions/me`, { credentials: "include" })
      .then((res) => {
        if (!res.ok) {
          if (res.status === 404) {
            setPermissions([]);
            setIsOwner(false);
            return null;
          }
          throw new Error(res.status === 403 ? "Keine Berechtigung" : "Fehler beim Laden");
        }
        return res.json() as Promise<MyPermissionsResponse>;
      })
      .then((data) => {
        if (cancelled || data == null) return;
        setIsOwner(data.is_owner);
        setPermissions(data.permissions ?? []);
      })
      .catch((e) => {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "Fehler beim Laden");
          setPermissions([]);
          setIsOwner(false);
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [clubId]);

  const canCreate = useCallback(
    (entityType: EntityType) => isOwner || hasPerm(permissions, entityType, "create"),
    [isOwner, permissions]
  );
  const canUpdate = useCallback(
    (entityType: EntityType) => isOwner || hasPerm(permissions, entityType, "update"),
    [isOwner, permissions]
  );
  const canDelete = useCallback(
    (entityType: EntityType) => isOwner || hasPerm(permissions, entityType, "delete"),
    [isOwner, permissions]
  );
  const canList = useCallback(
    (entityType: EntityType) => isOwner || hasPerm(permissions, entityType, "list"),
    [isOwner, permissions]
  );
  const canView = useCallback(
    (entityType: EntityType) => isOwner || hasPerm(permissions, entityType, "view"),
    [isOwner, permissions]
  );

  return {
    isOwner,
    permissions,
    isLoading,
    error,
    canCreate,
    canUpdate,
    canDelete,
    canList,
    canView,
  };
}
