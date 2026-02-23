import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useClub } from "@/context/ClubContext";
import { useAuth } from "@/context/AuthContext";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import { ArrowLeftIcon, CalculatorIcon, Loader2Icon, SaveIcon, Trash2Icon, UserCogIcon } from "lucide-react";
import { DeleteClubDialog } from "@/components/delete-club-dialog";
import { TransferClubOwnerDialog } from "@/components/transfer-club-owner-dialog";
import { formatMoneyInput } from "@/types/player";

function ClubEditPage() {
  const { clubId } = useParams<{ clubId: string }>();
  const navigate = useNavigate();
  const { setActiveClub, refreshClubs } = useClub();
  const { user, csrfToken, refreshCsrf } = useAuth();
  const [name, setName] = React.useState("");
  const [balance, setBalance] = React.useState("");
  const [startBalance, setStartBalance] = React.useState("");
  const [baseFee, setBaseFee] = React.useState("");
  const [autoTipEnabled, setAutoTipEnabled] = React.useState(true);
  const [couplesModeEnabled, setCouplesModeEnabled] = React.useState(false);
  const [clubUserId, setClubUserId] = React.useState<string | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isSaving, setIsSaving] = React.useState(false);
  const [isRecalculating, setIsRecalculating] = React.useState(false);
  const [transferOwnerDialogOpen, setTransferOwnerDialogOpen] = React.useState(false);
  const [deleteClubDialogOpen, setDeleteClubDialogOpen] = React.useState(false);

  const isOwner = clubUserId != null && user != null && clubUserId === user.id;

  const fetchClub = React.useCallback(async () => {
    if (!clubId) return;
    setIsLoading(true);
    try {
      const response = await fetch(`/api/clubs/${clubId}`, {
        credentials: "include",
      });
      if (!response.ok) {
        if (response.status === 404) {
          toast.error("Club nicht gefunden");
          navigate("/app", { replace: true });
          return;
        }
        if (response.status === 403) {
          toast.error("Keine Berechtigung, diesen Club einzusehen");
          navigate("/app", { replace: true });
          return;
        }
        throw new Error("Fehler beim Laden des Clubs");
      }
      const data = await response.json();
      setName(data.name ?? "");
      setBalance(data.balance != null ? (data.balance / 100).toFixed(2) : "0.00");
      setStartBalance(data.start_balance != null ? (data.start_balance / 100).toFixed(2) : "0.00");
      setBaseFee(data.base_fee != null ? (data.base_fee / 100).toFixed(2) : "0.00");
      setAutoTipEnabled(data.auto_tip_enabled ?? true);
      setCouplesModeEnabled(data.couples_mode_enabled ?? false);
      setClubUserId(data.user_id ?? null);
    } catch (error) {
      console.error("Error fetching club:", error);
      toast.error("Fehler beim Laden des Clubs");
      navigate("/app", { replace: true });
    } finally {
      setIsLoading(false);
    }
  }, [clubId, navigate]);

  React.useEffect(() => {
    fetchClub();
  }, [fetchClub]);

  // Redirect non-owners: they must not access club edit
  React.useEffect(() => {
    if (!isLoading && clubUserId != null && user != null && clubUserId !== user.id) {
      toast.error("Keine Berechtigung, diesen Club zu bearbeiten");
      navigate("/app", { replace: true });
    }
  }, [isLoading, clubUserId, user, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!clubId) return;

    const nameTrimmed = name.trim();
    if (!nameTrimmed) {
      toast.error("Name ist erforderlich");
      return;
    }
    const balanceCents = Math.round(parseFloat(balance || "0") * 100);
    const startBalanceCents = Math.round(parseFloat(startBalance || "0") * 100);
    const baseFeeCents = Math.round(parseFloat(baseFee || "0") * 100);
    if (balanceCents < 0) {
      toast.error("Balance darf nicht negativ sein");
      return;
    }
    if (baseFeeCents < 0) {
      toast.error("Basisgebühr darf nicht negativ sein");
      return;
    }

    setIsSaving(true);
    try {
      const token = csrfToken || (await refreshCsrf());
      const response = await fetch(`/api/clubs/${clubId}`, {
        method: "PUT",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify({
          name: nameTrimmed,
          balance: balanceCents,
          start_balance: startBalanceCents,
          base_fee: baseFeeCents,
          auto_tip_enabled: autoTipEnabled,
          couples_mode_enabled: couplesModeEnabled,
        }),
      });

      if (!response.ok) {
        const errData = await response.json().catch(() => ({}));
        throw new Error(errData.message || "Fehler beim Speichern");
      }

      const updated = await response.json();
      await refreshClubs();
      setActiveClub(updated);
      toast.success("Club gespeichert");
      navigate("/app", { replace: true });
    } catch (error) {
      console.error("Error saving club:", error);
      toast.error((error as Error).message);
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2Icon className="size-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/app")}
            aria-label="Zurück"
          >
            <ArrowLeftIcon className="size-4" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold">Club bearbeiten</h1>
            <p className="text-muted-foreground">
              Name, Startguthaben und Einstellungen für diesen Club
            </p>
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Einstellungen</CardTitle>
            <CardDescription>
              Nur der Owner darf diese Einstellungen ändern.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="club-name">Name</Label>
                <Input
                  id="club-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Club-Name"
                  required
                  disabled={isSaving}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="club-balance">Aktuelle Balance (€)</Label>
                <Input
                  id="club-balance"
                  type="number"
                  step="0.01"
                  min="0"
                  value={balance}
                  onChange={(e) => setBalance(e.target.value)}
                  onBlur={() => setBalance(formatMoneyInput(balance))}
                  placeholder="0.00"
                  disabled={isSaving}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="club-start-balance">Startguthaben (€)</Label>
                <Input
                  id="club-start-balance"
                  type="number"
                  step="0.01"
                  min="0"
                  value={startBalance}
                  onChange={(e) => setStartBalance(e.target.value)}
                  onBlur={() => setStartBalance(formatMoneyInput(startBalance))}
                  placeholder="0.00"
                  disabled={isSaving}
                />
                <p className="text-xs text-muted-foreground">
                  Basis für die Neuberechnung aus Transaktionen.
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="club-base-fee">Basisgebühr (€)</Label>
                <Input
                  id="club-base-fee"
                  type="number"
                  step="0.01"
                  min="0"
                  value={baseFee}
                  onChange={(e) => setBaseFee(e.target.value)}
                  onBlur={() => setBaseFee(formatMoneyInput(baseFee))}
                  placeholder="0.00"
                  disabled={isSaving}
                />
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="club-auto-tip"
                  checked={autoTipEnabled}
                  onCheckedChange={(checked) => setAutoTipEnabled(checked === true)}
                  disabled={isSaving}
                />
                <Label htmlFor="club-auto-tip" className="text-sm font-normal cursor-pointer">
                  Auto-Tip aktivieren
                </Label>
              </div>
              <p className="text-sm text-muted-foreground">
                Überschüssige Einzahlungen werden automatisch als Trinkgeld verbucht.
              </p>
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="club-couples-mode"
                  checked={couplesModeEnabled}
                  onCheckedChange={(checked) => setCouplesModeEnabled(checked === true)}
                  disabled={isSaving}
                />
                <Label htmlFor="club-couples-mode" className="text-sm font-normal cursor-pointer">
                  Paar-Modus aktivieren
                </Label>
              </div>
              <p className="text-sm text-muted-foreground">
                Bei Einzahlungen können mehrere Spieler ausgewählt werden; der Betrag wird gleichmäßig verteilt.
              </p>
              <div className="flex flex-wrap gap-2 pt-4">
                <Button type="submit" disabled={isSaving}>
                  {isSaving ? (
                    <Loader2Icon className="mr-2 size-4 animate-spin" />
                  ) : (
                    <SaveIcon className="mr-2 size-4" />
                  )}
                  Speichern
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate("/app")}
                  disabled={isSaving}
                >
                  Abbrechen
                </Button>
                {isOwner && (
                  <>
                    <Button
                      type="button"
                      variant="secondary"
                      disabled={isSaving || isRecalculating}
                      onClick={async () => {
                        if (!clubId) return;
                        setIsRecalculating(true);
                        try {
                          const token = csrfToken || (await refreshCsrf());
                          const response = await fetch(
                            `/api/clubs/${clubId}/recalculate-balance`,
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
                              "Fehler beim Neuberechnen der Club-Balance";
                            toast.error(message);
                            return;
                          }
                          const updated = body as { balance?: number };
                          setBalance(
                            updated.balance != null
                              ? (updated.balance / 100).toFixed(2)
                              : "0"
                          );
                          await refreshClubs();
                          setActiveClub(updated as Parameters<typeof setActiveClub>[0]);
                          toast.success("Club-Balance wurde neu berechnet.");
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
                      type="button"
                      variant="outline"
                      disabled={isSaving || isRecalculating}
                      onClick={() => setTransferOwnerDialogOpen(true)}
                    >
                      <UserCogIcon className="mr-2 size-4" />
                      Eigentümer wechseln
                    </Button>
                    <Button
                      type="button"
                      variant="destructive"
                      disabled={isSaving || isRecalculating}
                      onClick={() => setDeleteClubDialogOpen(true)}
                    >
                      <Trash2Icon className="mr-2 size-4" />
                      Club löschen
                    </Button>
                  </>
                )}
              </div>
            </form>
          </CardContent>
        </Card>

        {clubId && (
          <>
            <TransferClubOwnerDialog
              open={transferOwnerDialogOpen}
              onOpenChange={setTransferOwnerDialogOpen}
              clubId={clubId}
              clubName={name}
              onSuccess={() => {
                fetchClub();
                refreshClubs();
              }}
            />
            <DeleteClubDialog
              open={deleteClubDialogOpen}
              onOpenChange={setDeleteClubDialogOpen}
              clubId={clubId}
              clubName={name}
              onSuccess={() => {
                refreshClubs();
                navigate("/app", { replace: true });
              }}
            />
          </>
        )}
    </div>
  );
}

export default ClubEditPage;
