import * as React from "react";
import { useNavigate } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { PlayersDataTable } from "@/components/players-data-table";
import { PlayerDialog } from "@/components/player-dialog";
import { DeletePlayerDialog } from "@/components/delete-player-dialog";
import { useClub } from "@/context/ClubContext";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ShieldIcon } from "lucide-react";
import type { Player, Role } from "@/types/player";
import { toast } from "sonner";

function PlayersPage() {
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const [players, setPlayers] = React.useState<Player[]>([]);
  const [roles, setRoles] = React.useState<Role[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isRolesLoading, setIsRolesLoading] = React.useState(true);
  const [isDialogOpen, setIsDialogOpen] = React.useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);
  const [selectedPlayer, setSelectedPlayer] = React.useState<Player | null>(
    null
  );
  const [playerToDelete, setPlayerToDelete] = React.useState<Player | null>(
    null
  );

  const fetchPlayers = React.useCallback(async () => {
    if (!activeClub) return;

    setIsLoading(true);
    try {
      const response = await fetch(`/api/clubs/${activeClub.id}/players`, {
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Laden der Player");
      }

      const data: Player[] = await response.json();
      setPlayers(data);
    } catch (error) {
      console.error("Error fetching players:", error);
      toast.error("Fehler beim Laden der Player");
      setPlayers([]);
    } finally {
      setIsLoading(false);
    }
  }, [activeClub]);

  const fetchRoles = React.useCallback(async () => {
    if (!activeClub) return;

    setIsRolesLoading(true);
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
      setRoles([]);
    } finally {
      setIsRolesLoading(false);
    }
  }, [activeClub]);

  React.useEffect(() => {
    if (activeClub) {
      fetchPlayers();
      fetchRoles();
    }
  }, [activeClub, fetchPlayers, fetchRoles]);

  const handleCreate = () => {
    setSelectedPlayer(null);
    setIsDialogOpen(true);
  };

  const handleEdit = (player: Player) => {
    setSelectedPlayer(player);
    setIsDialogOpen(true);
  };

  const handleDelete = (player: Player) => {
    setPlayerToDelete(player);
    setIsDeleteDialogOpen(true);
  };

  const handleSuccess = () => {
    fetchPlayers();
  };

  if (!activeClub) {
    return (
      <AppLayout title="Spieler">
        <div className="flex flex-col items-center justify-center py-12">
          <p className="text-muted-foreground">
            Bitte wählen Sie einen Club aus, um die Spieler zu sehen.
          </p>
        </div>
      </AppLayout>
    );
  }

  // Show empty state if no roles exist
  if (!isRolesLoading && roles.length === 0) {
    return (
      <AppLayout title="Spieler">
        <div className="flex flex-col items-center justify-center py-12 px-4">
          <Card className="max-w-md">
            <CardHeader className="text-center">
              <div className="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-muted">
                <ShieldIcon className="size-6 text-muted-foreground" />
              </div>
              <CardTitle>Keine Rollen vorhanden</CardTitle>
              <CardDescription>
                Bevor Sie Spieler erstellen können, müssen Sie mindestens eine
                Rolle definieren. Rollen bestimmen die Berechtigungen und
                Eigenschaften der Spieler.
              </CardDescription>
            </CardHeader>
            <CardContent className="flex justify-center">
              <Button onClick={() => navigate("/app/roles")}>
                Erste Rolle erstellen
              </Button>
            </CardContent>
          </Card>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout title="Spieler">
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <PlayersDataTable
            players={players}
            roles={roles}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onCreate={handleCreate}
            isLoading={isLoading}
          />
        </div>
      </div>

      <PlayerDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        player={selectedPlayer}
        clubId={activeClub.id}
        roles={roles}
        onSuccess={handleSuccess}
      />

      <DeletePlayerDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        player={playerToDelete}
        clubId={activeClub.id}
        onSuccess={handleSuccess}
      />
    </AppLayout>
  );
}

export default PlayersPage;

