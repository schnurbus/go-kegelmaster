import * as React from "react";
import { AppLayout } from "@/components/AppLayout";
import { PenaltyTypesDataTable } from "@/components/penalty-types-data-table";
import { PenaltyTypeDialog } from "@/components/penalty-type-dialog";
import { DeletePenaltyTypeDialog } from "@/components/delete-penalty-type-dialog";
import { useClub } from "@/context/ClubContext";
import { usePermissions } from "@/hooks/use-permissions";
import type { PenaltyType } from "@/types/penalty-type";
import { toast } from "sonner";

function PenaltyTypesPage() {
  const { activeClub } = useClub();
  const { canCreate, canUpdate, canDelete } = usePermissions(activeClub?.id ?? null);
  const [penaltyTypes, setPenaltyTypes] = React.useState<PenaltyType[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isDialogOpen, setIsDialogOpen] = React.useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);
  const [selectedPenaltyType, setSelectedPenaltyType] =
    React.useState<PenaltyType | null>(null);
  const [penaltyTypeToDelete, setPenaltyTypeToDelete] =
    React.useState<PenaltyType | null>(null);

  const fetchPenaltyTypes = React.useCallback(async () => {
    if (!activeClub) return;

    setIsLoading(true);
    try {
      const response = await fetch(
        `/api/clubs/${activeClub.id}/penalty-types`,
        {
          credentials: "include",
        }
      );

      if (!response.ok) {
        throw new Error("Fehler beim Laden der Strafentypen");
      }

      const data: PenaltyType[] = await response.json();
      setPenaltyTypes(data);
    } catch (error) {
      console.error("Error fetching penalty types:", error);
      toast.error("Fehler beim Laden der Strafentypen");
      setPenaltyTypes([]);
    } finally {
      setIsLoading(false);
    }
  }, [activeClub]);

  React.useEffect(() => {
    if (activeClub) {
      fetchPenaltyTypes();
    }
  }, [activeClub, fetchPenaltyTypes]);

  const handleCreate = () => {
    setSelectedPenaltyType(null);
    setIsDialogOpen(true);
  };

  const handleEdit = (penaltyType: PenaltyType) => {
    setSelectedPenaltyType(penaltyType);
    setIsDialogOpen(true);
  };

  const handleDelete = (penaltyType: PenaltyType) => {
    setPenaltyTypeToDelete(penaltyType);
    setIsDeleteDialogOpen(true);
  };

  const handleDialogSuccess = () => {
    fetchPenaltyTypes();
  };

  const handleDeleteSuccess = () => {
    fetchPenaltyTypes();
  };

  const handleUpdateDisplayOrder = async (
    penaltyType: PenaltyType,
    newOrder: number
  ) => {
    if (!activeClub) return;

    try {
      const response = await fetch(
        `/api/clubs/${activeClub.id}/penalty-types/${penaltyType.id}/display-order`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": document.cookie
              .split("; ")
              .find((row) => row.startsWith("csrf_token="))
              ?.split("=")[1] || "",
          },
          credentials: "include",
          body: JSON.stringify({ display_order: newOrder }),
        }
      );

      if (!response.ok) {
        throw new Error("Fehler beim Aktualisieren der Sortierung");
      }

      toast.success("Sortierung aktualisiert");
      fetchPenaltyTypes();
    } catch (error) {
      console.error("Error updating display order:", error);
      toast.error("Fehler beim Aktualisieren der Sortierung");
    }
  };

  if (!activeClub) {
    return (
      <AppLayout>
        <div className="flex h-full items-center justify-center">
          <p className="text-muted-foreground">
            Bitte wählen Sie einen Club aus.
          </p>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="flex flex-col gap-6 p-6">
        <div>
          <h1 className="text-3xl font-bold">Strafentypen</h1>
          <p className="text-muted-foreground">
            Verwalten Sie die Strafentypen für {activeClub.name}
          </p>
        </div>

        <PenaltyTypesDataTable
          penaltyTypes={penaltyTypes}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onCreate={handleCreate}
          onUpdateDisplayOrder={handleUpdateDisplayOrder}
          isLoading={isLoading}
          canCreate={canCreate("penalty_types")}
          canUpdate={canUpdate("penalty_types")}
          canDelete={canDelete("penalty_types")}
        />

        <PenaltyTypeDialog
          open={isDialogOpen}
          onOpenChange={setIsDialogOpen}
          penaltyType={selectedPenaltyType}
          clubId={activeClub.id}
          onSuccess={handleDialogSuccess}
        />

        <DeletePenaltyTypeDialog
          open={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
          penaltyType={penaltyTypeToDelete}
          clubId={activeClub.id}
          onSuccess={handleDeleteSuccess}
        />
      </div>
    </AppLayout>
  );
}

export default PenaltyTypesPage;

