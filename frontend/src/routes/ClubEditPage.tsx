import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
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
import { ArrowLeftIcon, Loader2Icon, SaveIcon } from "lucide-react";

function ClubEditPage() {
  const { clubId } = useParams<{ clubId: string }>();
  const navigate = useNavigate();
  const { setActiveClub, refreshClubs } = useClub();
  const { csrfToken, refreshCsrf } = useAuth();
  const [name, setName] = React.useState("");
  const [balance, setBalance] = React.useState("");
  const [baseFee, setBaseFee] = React.useState("");
  const [autoTipEnabled, setAutoTipEnabled] = React.useState(true);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isSaving, setIsSaving] = React.useState(false);

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
        throw new Error("Fehler beim Laden des Clubs");
      }
      const data = await response.json();
      setName(data.name ?? "");
      setBalance(data.balance != null ? (data.balance / 100).toFixed(2) : "0");
      setBaseFee(data.base_fee != null ? (data.base_fee / 100).toFixed(2) : "0");
      setAutoTipEnabled(data.auto_tip_enabled ?? true);
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!clubId) return;

    const nameTrimmed = name.trim();
    if (!nameTrimmed) {
      toast.error("Name ist erforderlich");
      return;
    }
    const balanceCents = Math.round(parseFloat(balance || "0") * 100);
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
          base_fee: baseFeeCents,
          auto_tip_enabled: autoTipEnabled,
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
      <AppLayout title="Club bearbeiten">
        <div className="flex items-center justify-center py-12">
          <Loader2Icon className="size-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout title="Club bearbeiten">
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
                <Label htmlFor="club-balance">Startguthaben (€)</Label>
                <Input
                  id="club-balance"
                  type="number"
                  step="0.01"
                  min="0"
                  value={balance}
                  onChange={(e) => setBalance(e.target.value)}
                  placeholder="0.00"
                  disabled={isSaving}
                />
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
              <div className="flex gap-2 pt-4">
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
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

export default ClubEditPage;
