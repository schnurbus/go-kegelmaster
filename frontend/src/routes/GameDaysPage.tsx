import * as React from "react";
import { useNavigate } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { useClub } from "@/context/ClubContext";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CalendarIcon, PlusIcon } from "lucide-react";
import { toast } from "sonner";

interface GameDay {
  id: string;
  club_id: string;
  date: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

function GameDaysPage() {
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const [gameDays, setGameDays] = React.useState<GameDay[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);

  const fetchGameDays = React.useCallback(async () => {
    if (!activeClub) return;

    setIsLoading(true);
    try {
      const response = await fetch(`/api/clubs/${activeClub.id}/gamedays`, {
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Laden der Spieltage");
      }

      const data: GameDay[] = await response.json();
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
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Spieltage</h1>
            <p className="text-muted-foreground">
              Verwalten Sie die Spieltage für {activeClub.name}
            </p>
          </div>
          <Button onClick={handleCreateGameDay}>
            <PlusIcon className="mr-2 h-4 w-4" />
            Neuer Spieltag
          </Button>
        </div>

        {isLoading ? (
          <p>Laden...</p>
        ) : gameDays.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-center text-muted-foreground">
                Noch keine Spieltage vorhanden
              </p>
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-4">
            {gameDays.map((gameDay) => (
              <Card
                key={gameDay.id}
                className="cursor-pointer hover:border-primary transition-colors"
                onClick={() => navigate(`/app/gamedays/${gameDay.id}`)}
              >
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <CalendarIcon className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <CardTitle>
                          {new Date(gameDay.date).toLocaleDateString("de-DE", {
                            weekday: "long",
                            year: "numeric",
                            month: "long",
                            day: "numeric",
                          })}
                        </CardTitle>
                        {gameDay.notes && (
                          <CardDescription className="mt-1">
                            {gameDay.notes}
                          </CardDescription>
                        )}
                      </div>
                    </div>
                  </div>
                </CardHeader>
              </Card>
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  );
}

export default GameDaysPage;
