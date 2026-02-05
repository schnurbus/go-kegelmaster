import * as React from "react";
import { AppLayout } from "@/components/AppLayout";
import { CompetitionsDataTable } from "@/components/competitions-data-table";
import { CompetitionDialog } from "@/components/competition-dialog";
import { DeleteCompetitionDialog } from "@/components/delete-competition-dialog";
import { useClub } from "@/context/ClubContext";
import { usePermissions } from "@/hooks/use-permissions";
import type { Competition } from "@/types/competition";
import { toast } from "sonner";

function CompetitionsPage() {
  const { activeClub } = useClub();
  const { canCreate, canUpdate, canDelete } = usePermissions(activeClub?.id ?? null);
  const [competitions, setCompetitions] = React.useState<Competition[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isDialogOpen, setIsDialogOpen] = React.useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);
  const [selectedCompetition, setSelectedCompetition] =
    React.useState<Competition | null>(null);
  const [competitionToDelete, setCompetitionToDelete] =
    React.useState<Competition | null>(null);

  const fetchCompetitions = React.useCallback(async () => {
    if (!activeClub) return;

    setIsLoading(true);
    try {
      const response = await fetch(
        `/api/clubs/${activeClub.id}/competitions`,
        {
          credentials: "include",
        }
      );

      if (!response.ok) {
        throw new Error("Fehler beim Laden der Wettbewerbe");
      }

      const data: Competition[] = await response.json();
      setCompetitions(data);
    } catch (error) {
      console.error("Error fetching competitions:", error);
      toast.error("Fehler beim Laden der Wettbewerbe");
      setCompetitions([]);
    } finally {
      setIsLoading(false);
    }
  }, [activeClub]);

  React.useEffect(() => {
    if (activeClub) {
      fetchCompetitions();
    }
  }, [activeClub, fetchCompetitions]);

  const handleCreate = () => {
    setSelectedCompetition(null);
    setIsDialogOpen(true);
  };

  const handleEdit = (competition: Competition) => {
    setSelectedCompetition(competition);
    setIsDialogOpen(true);
  };

  const handleDelete = (competition: Competition) => {
    setCompetitionToDelete(competition);
    setIsDeleteDialogOpen(true);
  };

  const handleDialogSuccess = () => {
    fetchCompetitions();
  };

  const handleDeleteSuccess = () => {
    fetchCompetitions();
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
          <h1 className="text-3xl font-bold">Wettbewerbe</h1>
          <p className="text-muted-foreground">
            Verwalten Sie die Wettbewerbe für {activeClub.name}
          </p>
        </div>

        <CompetitionsDataTable
          competitions={competitions}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onCreate={handleCreate}
          isLoading={isLoading}
          clubId={activeClub.id}
          canCreate={canCreate("competitions")}
          canUpdate={canUpdate("competitions")}
          canDelete={canDelete("competitions")}
        />

        <CompetitionDialog
          open={isDialogOpen}
          onOpenChange={setIsDialogOpen}
          competition={selectedCompetition}
          clubId={activeClub.id}
          onSuccess={handleDialogSuccess}
        />

        <DeleteCompetitionDialog
          open={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
          competition={competitionToDelete}
          clubId={activeClub.id}
          onSuccess={handleDeleteSuccess}
        />
      </div>
    </AppLayout>
  );
}

export default CompetitionsPage;
