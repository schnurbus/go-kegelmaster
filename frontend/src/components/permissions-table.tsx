import * as React from "react";
import { toast } from "sonner";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  type Role,
  type EntityType,
  type PermissionType,
  hasPermission,
  getEntityTypeLabel,
  getPermissionTypeLabel,
} from "@/types/role";

type PermissionsTableProps = {
  role: Role;
  clubId: string;
  onUpdate: () => void;
};

const ENTITY_TYPES: EntityType[] = ["roles", "players", "game_days", "penalty_types"];
const PERMISSION_TYPES: PermissionType[] = [
  "list",
  "view",
  "create",
  "update",
  "delete",
];

export function PermissionsTable({
  role,
  clubId,
  onUpdate,
}: PermissionsTableProps) {
  const [loadingStates, setLoadingStates] = React.useState<
    Record<string, boolean>
  >({});

  const getLoadingKey = (
    entityType: EntityType,
    permissionType: PermissionType
  ) => {
    return `${entityType}-${permissionType}`;
  };

  const isLoading = (
    entityType: EntityType,
    permissionType: PermissionType
  ) => {
    return loadingStates[getLoadingKey(entityType, permissionType)] || false;
  };

  const handlePermissionChange = async (
    entityType: EntityType,
    permissionType: PermissionType,
    checked: boolean
  ) => {
    const loadingKey = getLoadingKey(entityType, permissionType);
    setLoadingStates((prev) => ({ ...prev, [loadingKey]: true }));

    try {
      const url = `/api/clubs/${clubId}/roles/${role.id}/permissions`;
      const method = checked ? "POST" : "DELETE";

      const response = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": document.cookie
            .split("; ")
            .find((row) => row.startsWith("csrf_token="))
            ?.split("=")[1] || "",
        },
        credentials: "include",
        body: JSON.stringify({
          entity_type: entityType,
          permission_type: permissionType,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
            `Fehler beim ${checked ? "Hinzufügen" : "Entfernen"} der Berechtigung`
        );
      }

      toast.success(
        checked
          ? "Berechtigung erfolgreich hinzugefügt"
          : "Berechtigung erfolgreich entfernt"
      );
      onUpdate();
    } catch (error) {
      console.error("Error updating permission:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Aktualisieren"
      );
      // Revert the change on error by triggering a refresh
      onUpdate();
    } finally {
      setLoadingStates((prev) => {
        const newState = { ...prev };
        delete newState[loadingKey];
        return newState;
      });
    }
  };

  return (
    <div className="space-y-4">
      <div className="text-sm text-muted-foreground">
        Legen Sie fest, welche Berechtigungen diese Rolle für verschiedene
        Entitäten hat. Aktivieren Sie die Checkboxen für die gewünschten
        Kombinationen.
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[150px]">Entität</TableHead>
              {PERMISSION_TYPES.map((permissionType) => (
                <TableHead key={permissionType} className="text-center">
                  {getPermissionTypeLabel(permissionType)}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {ENTITY_TYPES.map((entityType) => (
              <TableRow key={entityType}>
                <TableCell className="font-medium">
                  {getEntityTypeLabel(entityType)}
                </TableCell>
                {PERMISSION_TYPES.map((permissionType) => {
                  const isChecked = hasPermission(
                    role,
                    entityType,
                    permissionType
                  );
                  const loading = isLoading(entityType, permissionType);

                  return (
                    <TableCell key={permissionType} className="text-center">
                      <div className="flex justify-center">
                        <Checkbox
                          checked={isChecked}
                          disabled={loading}
                          onCheckedChange={(checked) =>
                            handlePermissionChange(
                              entityType,
                              permissionType,
                              checked === true
                            )
                          }
                          className={loading ? "opacity-50" : ""}
                        />
                      </div>
                    </TableCell>
                  );
                })}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <div className="text-xs text-muted-foreground">
        <p className="font-medium mb-1">Hinweise:</p>
        <ul className="list-disc list-inside space-y-1">
          <li>
            <strong>Auflisten:</strong> Berechtigung zum Abrufen einer Liste
            aller Einträge
          </li>
          <li>
            <strong>Ansehen:</strong> Berechtigung zum Anzeigen einzelner
            Einträge
          </li>
          <li>
            <strong>Erstellen:</strong> Berechtigung zum Erstellen neuer
            Einträge
          </li>
          <li>
            <strong>Bearbeiten:</strong> Berechtigung zum Ändern bestehender
            Einträge
          </li>
          <li>
            <strong>Löschen:</strong> Berechtigung zum Entfernen von Einträgen
          </li>
        </ul>
      </div>
    </div>
  );
}

