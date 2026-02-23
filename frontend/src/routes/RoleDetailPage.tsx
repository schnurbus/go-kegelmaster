import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { RoleDialog } from "@/components/role-dialog";
import { DeleteRoleDialog } from "@/components/delete-role-dialog";
import { PermissionsTable } from "@/components/permissions-table";
import { useClub } from "@/context/ClubContext";
import { usePermissions } from "@/hooks/use-permissions";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import {
  ArrowLeftIcon,
  EditIcon,
  TrashIcon,
  ShieldIcon,
  LockIcon,
} from "lucide-react";
import { usePageTitle } from "@/context/PageTitleContext";
import { toast } from "sonner";

import type { Role } from "@/types/role";

function RoleDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const { canUpdate, canDelete } = usePermissions(activeClub?.id ?? null);
  const { setTitle } = usePageTitle();
  const [role, setRole] = React.useState<Role | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isEditDialogOpen, setIsEditDialogOpen] = React.useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);

  const fetchRole = React.useCallback(async () => {
    if (!activeClub || !id) return;

    setIsLoading(true);
    try {
      const response = await fetch(
        `/api/clubs/${activeClub.id}/roles/${id}`,
        {
          credentials: "include",
        }
      );

      if (!response.ok) {
        if (response.status === 404) {
          toast.error("Rolle nicht gefunden");
          navigate("/app/roles");
          return;
        }
        throw new Error("Fehler beim Laden der Rolle");
      }

      const data: Role = await response.json();
      setRole(data);
    } catch (error) {
      console.error("Error fetching role:", error);
      toast.error("Fehler beim Laden der Rolle");
    } finally {
      setIsLoading(false);
    }
  }, [activeClub, id, navigate]);

  React.useEffect(() => {
    if (activeClub) {
      fetchRole();
    }
  }, [activeClub, fetchRole]);

  React.useEffect(() => {
    if (role) setTitle(role.name);
  }, [role, setTitle]);

  const handleSuccess = () => {
    fetchRole();
  };

  const handleDeleteSuccess = () => {
    navigate("/app/roles");
  };

  if (!activeClub) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-muted-foreground">
          Bitte wählen Sie einen Club aus.
        </p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <div className="mb-6">
            <Skeleton className="h-10 w-32" />
          </div>
          <div className="grid gap-6 md:grid-cols-2">
            <Skeleton className="h-[200px]" />
            <Skeleton className="h-[200px]" />
          </div>
        </div>
      </div>
    );
  }

  if (!role) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-muted-foreground">Rolle nicht gefunden</p>
        <Button
          variant="outline"
          className="mt-4"
          onClick={() => navigate("/app/roles")}
        >
          Zurück zur Übersicht
        </Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          {/* Header with back button and actions */}
          <div className="mb-6 flex items-center justify-between">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate("/app/roles")}
            >
              <ArrowLeftIcon className="mr-2 size-4" />
              Zurück
            </Button>
            {(canUpdate("roles") || canDelete("roles")) && (
              <div className="flex gap-2">
                {canUpdate("roles") && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setIsEditDialogOpen(true)}
                  >
                    <EditIcon className="mr-2 size-4" />
                    Bearbeiten
                  </Button>
                )}
                {canDelete("roles") && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setIsDeleteDialogOpen(true)}
                    className="text-red-600 hover:text-red-700"
                  >
                    <TrashIcon className="mr-2 size-4" />
                    Löschen
                  </Button>
                )}
              </div>
            )}
          </div>

          {/* Role Information Card */}
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <ShieldIcon className="size-5" />
                Rollen-Informationen
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="text-sm text-muted-foreground">Name</p>
                <p className="text-lg font-medium">{role.name}</p>
              </div>
              <Separator />
              <div>
                <p className="text-sm text-muted-foreground">Grundgebühr</p>
                <Badge variant={role.pays_base_fee ? "default" : "secondary"}>
                  {role.pays_base_fee
                    ? "Zahlt Grundgebühr"
                    : "Keine Grundgebühr"}
                </Badge>
              </div>
              <Separator />
              <div>
                <p className="text-sm text-muted-foreground">Erstellt am</p>
                <p className="text-sm">
                  {new Date(role.created_at).toLocaleDateString("de-DE", {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })}
                </p>
              </div>
              <Separator />
              <div>
                <p className="text-sm text-muted-foreground">
                  Anzahl Berechtigungen
                </p>
                <Badge variant="outline">{role.permissions.length}</Badge>
              </div>
            </CardContent>
          </Card>

          {/* Permissions Card */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <LockIcon className="size-5" />
                Berechtigungen
              </CardTitle>
              <CardDescription>
                Verwalten Sie die Berechtigungen für diese Rolle. Änderungen
                werden sofort gespeichert.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <PermissionsTable
                role={role}
                clubId={activeClub.id}
                onUpdate={handleSuccess}
                canEdit={canUpdate("roles")}
              />
            </CardContent>
          </Card>
        </div>

      <RoleDialog
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
        role={role}
        clubId={activeClub.id}
        onSuccess={handleSuccess}
      />

      <DeleteRoleDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        role={role}
        clubId={activeClub.id}
        onSuccess={handleDeleteSuccess}
      />
    </div>
  );
}

export default RoleDetailPage;

