import * as React from "react";
import { useNavigate } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { useClub } from "@/context/ClubContext";
import { usePermissions } from "@/hooks/use-permissions";
import { GameDaysDataTable } from "@/components/game-days-data-table";
import type { GameDaySummary } from "@/types/gameday";
import { toast } from "sonner";

function GameDaysPage() {
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const { canCreate, canUpdate, canDelete } = usePermissions(activeClub?.id ?? null);
  const [gameDays, setGameDays] = React.useState<GameDaySummary[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);

  const fetchCSRFToken = React.useCallback(async () => {
    const response = await fetch("/api/auth/csrf-token", {
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error("Failed to fetch CSRF token");
    }
    const data = await response.json();
    return data.csrf_token;
  }, []);

  const fetchGameDays = React.useCallback(async () => {
    if (!activeClub) return;

    setIsLoading(true);
    try {
      const response = await fetch(`/api/clubs/${activeClub.id}/gamedays/summaries`, {
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Laden der Spieltage");
      }

      const data: GameDaySummary[] = await response.json();
      setGameDays(data || []);
    } catch (error) {
      console.error("Error fetching game days:", error);
      toast.error("Fehler beim Laden der Spieltage");
    } finally {
      setIsLoading(false);
    }
  }, [activeClub]);

  React.useEffect(() => {
    fetchGameDays();
  }, [fetchGameDays]);

  const handleCreateGameDay = () => {
    navigate("/app/gamedays/new");
  };

  const handleViewGameDay = (gameDay: GameDaySummary) => {
    navigate(`/app/gamedays/${gameDay.id}`);
  };

  const handleDeleteGameDay = async (gameDay: GameDaySummary) => {
    const formattedDate = new Date(gameDay.date).toLocaleDateString("de-DE", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });

    if (!confirm(`Möchten Sie den Spieltag vom ${formattedDate} wirklich löschen?`)) {
      return;
    }

    try {
      const csrfToken = await fetchCSRFToken();
      const response = await fetch(`/api/clubs/${activeClub!.id}/gamedays/${gameDay.id}`, {
        method: "DELETE",
        headers: {
          "X-CSRF-Token": csrfToken,
        },
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Löschen");
      }

      toast.success("Spieltag gelöscht");
      fetchGameDays();
    } catch (error: any) {
      console.error("Error deleting game day:", error);
      toast.error(error.message || "Fehler beim Löschen");
    }
  };

  if (!activeClub) {
    return (
      <AppLayout>
        <div className="p-6">
          <p>Bitte wählen Sie einen Klub aus.</p>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="p-6 space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Spieltage</h1>
          <p className="text-muted-foreground">
            Verwalten Sie die Spieltage für {activeClub.name}
          </p>
        </div>

        <GameDaysDataTable
          gameDays={gameDays}
          onView={handleViewGameDay}
          onDelete={handleDeleteGameDay}
          onCreate={handleCreateGameDay}
          isLoading={isLoading}
          clubId={activeClub.id}
          canCreate={canCreate("game_days")}
          canUpdate={canUpdate("game_days")}
          canDelete={canDelete("game_days")}
        />
      </div>
    </AppLayout>
  );
}

export default GameDaysPage;
