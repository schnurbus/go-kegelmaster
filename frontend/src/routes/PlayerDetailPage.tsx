import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { PlayerDialog } from "@/components/player-dialog";
import { DeletePlayerDialog } from "@/components/delete-player-dialog";
import { useClub } from "@/context/ClubContext";
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
  AlertTriangleIcon,
  ArrowLeftIcon,
  CalculatorIcon,
  EditIcon,
  Loader2Icon,
  TrashIcon,
  UserIcon,
  WalletIcon,
} from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import { toast } from "sonner";

import type { Player, Role } from "@/types/player";
import { formatCentsToEuro } from "@/types/player";

function PlayerDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const { csrfToken, refreshCsrf } = useAuth();
  const [player, setPlayer] = React.useState<Player | null>(null);
  const [roles, setRoles] = React.useState<Role[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isEditDialogOpen, setIsEditDialogOpen] = React.useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = React.useState(false);
  const [isRecalculating, setIsRecalculating] = React.useState(false);

  const fetchPlayer = React.useCallback(async () => {
    if (!activeClub || !id) return;

    setIsLoading(true);
    try {
      const response = await fetch(
        `/api/clubs/${activeClub.id}/players/${id}`,
        {
          credentials: "include",
        }
      );

      if (!response.ok) {
        if (response.status === 404) {
          toast.error("Player nicht gefunden");
          navigate("/app/players");
          return;
        }
        throw new Error("Fehler beim Laden des Players");
      }

      const data: Player = await response.json();
      setPlayer(data);
    } catch (error) {
      console.error("Error fetching player:", error);
      toast.error("Fehler beim Laden des Players");
    } finally {
      setIsLoading(false);
    }
  }, [activeClub, id, navigate]);

  const fetchRoles = React.useCallback(async () => {
    if (!activeClub) return;

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
    }
  }, [activeClub]);

  React.useEffect(() => {
    if (activeClub) {
      fetchPlayer();
      fetchRoles();
    }
  }, [activeClub, fetchPlayer, fetchRoles]);

  const getRoleName = (roleId: string | null) => {
    if (!roleId) return "Keine Rolle zugewiesen";
    const role = roles.find((r) => r.id === roleId);
    return role?.name || "Rolle nicht gefunden";
  };

  const handleSuccess = () => {
    fetchPlayer();
  };

  const handleDeleteSuccess = () => {
    navigate("/app/players");
  };

  if (!activeClub) {
    return (
      <AppLayout title="Player Details">
        <div className="flex flex-col items-center justify-center py-12">
          <p className="text-muted-foreground">
            Bitte wählen Sie einen Club aus.
          </p>
        </div>
      </AppLayout>
    );
  }

  if (isLoading) {
    return (
      <AppLayout title="Player Details">
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
      </AppLayout>
    );
  }

  if (!player) {
    return (
      <AppLayout title="Player Details">
        <div className="flex flex-col items-center justify-center py-12">
          <p className="text-muted-foreground">Player nicht gefunden</p>
          <Button
            variant="outline"
            className="mt-4"
            onClick={() => navigate("/app/players")}
          >
            Zurück zur Übersicht
          </Button>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout title={player.name}>
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          {/* Header with back button and actions */}
          <div className="mb-6 flex items-center justify-between">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate("/app/players")}
            >
              <ArrowLeftIcon className="mr-2 size-4" />
              Zurück
            </Button>
            <div className="flex flex-wrap gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsEditDialogOpen(true)}
              >
                <EditIcon className="mr-2 size-4" />
                Bearbeiten
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={isRecalculating}
                onClick={async () => {
                  if (!activeClub || !player) return;
                  setIsRecalculating(true);
                  try {
                    const token = csrfToken || (await refreshCsrf());
                    const response = await fetch(
                      `/api/clubs/${activeClub.id}/players/${player.id}/recalculate-balance`,
                      {
                        method: "POST",
                        credentials: "include",
                        headers: { "X-CSRF-Token": token },
                      }
                    );
                    const body = await response.json().catch(() => ({}));
                    if (!response.ok) {
                      const message =
                        (body as { message?: string }).message ||
                        "Fehler beim Neuberechnen des Saldos";
                      toast.error(message);
                      return;
                    }
                    const updated = body as Player;
                    setPlayer(updated);
                    toast.success("Saldo wurde neu berechnet.");
                  } catch (error) {
                    console.error("Recalculate balance:", error);
                    toast.error(
                      (error as Error).message || "Etwas ist schiefgelaufen."
                    );
                  } finally {
                    setIsRecalculating(false);
                  }
                }}
              >
                {isRecalculating ? (
                  <Loader2Icon className="mr-2 size-4 animate-spin" />
                ) : (
                  <CalculatorIcon className="mr-2 size-4" />
                )}
                Balance neu berechnen
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsDeleteDialogOpen(true)}
                className="text-red-600 hover:text-red-700"
              >
                <TrashIcon className="mr-2 size-4" />
                Löschen
              </Button>
            </div>
          </div>

          {/* Player Information Cards */}
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <UserIcon className="size-5" />
                  Spieler-Informationen
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <p className="text-sm text-muted-foreground">Name</p>
                  <p className="text-lg font-medium">{player.name}</p>
                </div>
                <Separator />
                <div>
                  <p className="text-sm text-muted-foreground">Rolle</p>
                  <Badge variant={player.role_id ? "outline" : "secondary"}>
                    {getRoleName(player.role_id)}
                  </Badge>
                </div>
                <Separator />
                <div>
                  <p className="text-sm text-muted-foreground">Erstellt am</p>
                  <p className="text-sm">
                    {new Date(player.created_at).toLocaleDateString("de-DE", {
                      year: "numeric",
                      month: "long",
                      day: "numeric",
                    })}
                  </p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <WalletIcon className="size-5" />
                  Finanzinformationen
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <p className="text-sm text-muted-foreground">
                    Aktuelle Balance
                  </p>
                  <div className="flex items-center gap-2">
                    <p
                      className={`text-2xl font-bold ${
                        player.balance < 0
                          ? "text-red-500"
                          : player.balance > 0
                            ? "text-green-500"
                            : ""
                      }`}
                    >
                      {formatCentsToEuro(player.balance)}
                    </p>
                    {player.balance > 0 && (
                      <span title="Positives Guthaben - sollte mit Auto-Tip nicht vorkommen">
                        <AlertTriangleIcon className="size-5 text-yellow-500" />
                      </span>
                    )}
                  </div>
                  {player.balance > 0 && (
                    <p className="text-sm text-yellow-600 dark:text-yellow-500 mt-1">
                      ⚠️ Achtung: Spieler hat positives Guthaben
                    </p>
                  )}
                </div>
                <Separator />
                <div>
                  <p className="text-sm text-muted-foreground">Start-Balance</p>
                  <p className="text-lg font-medium">
                    {formatCentsToEuro(player.start_balance)}
                  </p>
                </div>
                <Separator />
                <div>
                  <p className="text-sm text-muted-foreground">Differenz</p>
                  <p
                    className={`text-lg font-medium ${
                      player.balance - player.start_balance < 0
                        ? "text-red-500"
                        : player.balance - player.start_balance > 0
                          ? "text-green-500"
                          : ""
                    }`}
                  >
                    {formatCentsToEuro(player.balance - player.start_balance)}
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Transaction History Placeholder */}
          <Card className="mt-6">
            <CardHeader>
              <CardTitle>Transaktionshistorie</CardTitle>
              <CardDescription>
                Hier werden zukünftig alle Transaktionen des Players angezeigt.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-center rounded-lg border border-dashed py-12">
                <p className="text-sm text-muted-foreground">
                  Transaktionshistorie kommt bald
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <PlayerDialog
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
        player={player}
        clubId={activeClub.id}
        roles={roles}
        onSuccess={handleSuccess}
      />

      <DeletePlayerDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        player={player}
        clubId={activeClub.id}
        onSuccess={handleDeleteSuccess}
      />
    </AppLayout>
  );
}

export default PlayerDetailPage;

