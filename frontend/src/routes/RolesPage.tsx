import * as React from "react";
import { useNavigate } from "react-router-dom";
import { RolesDataTable } from "@/components/roles-data-table";
import { RoleDialog } from "@/components/role-dialog";
import { DeleteRoleDialog } from "@/components/delete-role-dialog";
import { useClub } from "@/context/ClubContext";
import { usePermissions } from "@/hooks/use-permissions";
import type { Role } from "@/types/role";
import { toast } from "sonner";

function RolesPage() {
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const { canCreate, canUpdate, canDelete } = usePermissions(activeClub?.id ?? null);
  const [roles, setRoles] = React.useState<Role[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isDialogOpen, setIsDialogOpen] = React.useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);
  const [selectedRole, setSelectedRole] = React.useState<Role | null>(null);
  const [roleToDelete, setRoleToDelete] = React.useState<Role | null>(null);

  const fetchRoles = React.useCallback(async () => {
    if (!activeClub) return;

    setIsLoading(true);
    try {
      const response = await fetch(`/api/clubs/${activeClub.id}/roles`, {
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Laden der Rollen");
      }

      const data: Role[] = await response.json();
      setRoles(data);
    } catch (error) {
      console.error("Error fetching roles:", error);
      toast.error("Fehler beim Laden der Rollen");
      setRoles([]);
    } finally {
      setIsLoading(false);
    }
  }, [activeClub]);

  React.useEffect(() => {
    if (activeClub) {
      fetchRoles();
    }
  }, [activeClub, fetchRoles]);

  const handleCreate = () => {
    setSelectedRole(null);
    setIsDialogOpen(true);
  };

  const handleView = (role: Role) => {
    navigate(`/app/roles/${role.id}`);
  };

  const handleEdit = (role: Role) => {
    setSelectedRole(role);
    setIsDialogOpen(true);
  };

  const handleDelete = (role: Role) => {
    setRoleToDelete(role);
    setIsDeleteDialogOpen(true);
  };

  const handleSuccess = () => {
    fetchRoles();
  };

  if (!activeClub) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-muted-foreground">
          Bitte wählen Sie einen Club aus, um die Rollen zu sehen.
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="px-4 lg:px-6">
        <RolesDataTable
          roles={roles}
          onView={handleView}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onCreate={handleCreate}
          isLoading={isLoading}
          clubId={activeClub.id}
          canCreate={canCreate("roles")}
          canUpdate={canUpdate("roles")}
          canDelete={canDelete("roles")}
        />
      </div>

      <RoleDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        role={selectedRole}
        clubId={activeClub.id}
        onSuccess={handleSuccess}
      />

      <DeleteRoleDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        role={roleToDelete}
        clubId={activeClub.id}
        onSuccess={handleSuccess}
      />
    </div>
  );
}

export default RolesPage;

