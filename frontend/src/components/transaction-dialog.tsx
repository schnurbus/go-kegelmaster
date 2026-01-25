import * as React from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { AlertTriangleIcon, GiftIcon, InfoIcon } from "lucide-react";

import type { CreateTransactionRequest } from "@/types/transaction";
import type { Player } from "@/types/player";
import { euroToCents, formatCentsToEuro } from "@/types/player";

type TransactionDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clubId: string;
  onSuccess: () => void;
};

export function TransactionDialog({
  open,
  onOpenChange,
  clubId,
  onSuccess,
}: TransactionDialogProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [players, setPlayers] = React.useState<Player[]>([]);
  const [isLoadingPlayers, setIsLoadingPlayers] = React.useState(false);
  const [autoTipEnabled, setAutoTipEnabled] = React.useState(true);
  const [clubBalance, setClubBalance] = React.useState<number>(0);

  const [transactionType, setTransactionType] = React.useState<
    "deposit" | "tip" | "expense"
  >("deposit");
  const [playerId, setPlayerId] = React.useState<string>("");
  const [amount, setAmount] = React.useState("");
  const [description, setDescription] = React.useState("");

  // Fetch players and club info when dialog opens
  React.useEffect(() => {
    if (open) {
      fetchPlayers();
      fetchClubInfo();
    }
  }, [open, clubId]);

  const fetchPlayers = async () => {
    setIsLoadingPlayers(true);
    try {
      const response = await fetch(`/api/clubs/${clubId}/players`, {
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Laden der Spieler");
      }

      const data: Player[] = await response.json();
      setPlayers(data);
    } catch (error) {
      console.error("Error fetching players:", error);
      toast.error("Fehler beim Laden der Spieler");
    } finally {
      setIsLoadingPlayers(false);
    }
  };

  const fetchClubInfo = async () => {
    try {
      const response = await fetch(`/api/clubs/${clubId}`, {
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Laden der Club-Informationen");
      }

      const data = await response.json();
      setAutoTipEnabled(data.auto_tip_enabled ?? true);
      setClubBalance(data.balance ?? 0);
    } catch (error) {
      console.error("Error fetching club info:", error);
      // Default to true if we can't fetch
      setAutoTipEnabled(true);
      setClubBalance(0);
    }
  };

  // Reset form when dialog opens/closes
  React.useEffect(() => {
    if (open) {
      setTransactionType("deposit");
      setPlayerId("");
      setAmount("");
      setDescription("");
    }
  }, [open]);

  // Calculate auto-tip preview
  const getAutoTipPreview = () => {
    if (transactionType !== "deposit" || !playerId || !amount || !autoTipEnabled) {
      return null;
    }

    const player = players.find((p) => p.id === playerId);
    if (!player) return null;

    const amountCents = euroToCents(parseFloat(amount) || 0);
    const playerBalance = player.balance;

    // Player balance is negative (debt), deposit is positive
    if (playerBalance >= 0) {
      // Player has no debt, entire deposit becomes tip
      return {
        depositAmount: 0,
        tipAmount: amountCents,
        newBalance: 0,
      };
    }

    const debt = Math.abs(playerBalance);
    
    if (amountCents <= debt) {
      // Deposit doesn't cover full debt, no auto-tip
      return null;
    }

    // Deposit covers debt with excess
    const depositToBalance = debt;
    const autoTip = amountCents - debt;

    return {
      depositAmount: depositToBalance,
      tipAmount: autoTip,
      newBalance: 0,
    };
  };

  const autoTipPreview = getAutoTipPreview();
  const selectedPlayer = players.find((p) => p.id === playerId);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!amount || parseFloat(amount) <= 0) {
      toast.error("Betrag muss größer als 0 sein");
      return;
    }

    if (!description.trim()) {
      toast.error("Beschreibung ist erforderlich");
      return;
    }

    if (transactionType === "deposit" && !playerId) {
      toast.error("Bitte wählen Sie einen Spieler aus");
      return;
    }

    setIsSubmitting(true);

    try {
      const body: CreateTransactionRequest = {
        transaction_type: transactionType,
        amount: euroToCents(parseFloat(amount)),
        description: description.trim(),
        player_id: playerId || undefined,
      };

      const response = await fetch(`/api/clubs/${clubId}/transactions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token":
            document.cookie
              .split("; ")
              .find((row) => row.startsWith("csrf_token="))
              ?.split("=")[1] || "",
        },
        credentials: "include",
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Erstellen der Transaktion");
      }

      const successMessage = autoTipPreview
        ? `Transaktion erstellt: ${formatCentsToEuro(autoTipPreview.depositAmount)} Einzahlung + ${formatCentsToEuro(autoTipPreview.tipAmount)} Auto-Tip`
        : "Transaktion erfolgreich erstellt";

      toast.success(successMessage);
      onSuccess();
      onOpenChange(false);
    } catch (error) {
      console.error("Error creating transaction:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Erstellen der Transaktion"
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const getDefaultDescription = () => {
    switch (transactionType) {
      case "deposit":
        return "Einzahlung";
      case "tip":
        return "Trinkgeld";
      case "expense":
        return "Ausgabe";
      default:
        return "";
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Transaktion erstellen</DialogTitle>
          <DialogDescription>
            Erstellen Sie eine manuelle Transaktion (Einzahlung, Trinkgeld oder Ausgabe).
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="transaction_type">
              Transaktionstyp <span className="text-red-500">*</span>
            </Label>
            <Select
              value={transactionType}
              onValueChange={(value: "deposit" | "tip" | "expense") => {
                setTransactionType(value);
                setPlayerId(""); // Reset player when changing type
              }}
              required
            >
              <SelectTrigger id="transaction_type">
                <SelectValue placeholder="Typ auswählen" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="deposit">Einzahlung</SelectItem>
                <SelectItem value="tip">Trinkgeld</SelectItem>
                <SelectItem value="expense">Ausgabe</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-sm text-muted-foreground">
              Grundgebühren und Strafgebühren werden automatisch erstellt
            </p>
          </div>

          {transactionType === "deposit" && (
            <div className="space-y-2">
              <Label htmlFor="player_id">
                Spieler <span className="text-red-500">*</span>
              </Label>
              <Select
                value={playerId}
                onValueChange={setPlayerId}
                required
                disabled={isLoadingPlayers}
              >
                <SelectTrigger id="player_id">
                  <SelectValue placeholder="Spieler auswählen" />
                </SelectTrigger>
                <SelectContent>
                  {players.map((player) => (
                    <SelectItem key={player.id} value={player.id}>
                      {player.name} - Balance: {formatCentsToEuro(player.balance)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {selectedPlayer && (
                <p className="text-sm text-muted-foreground">
                  Aktuelles Guthaben:{" "}
                  <span
                    className={
                      selectedPlayer.balance < 0
                        ? "text-red-500 font-medium"
                        : selectedPlayer.balance > 0
                          ? "text-green-500 font-medium"
                          : ""
                    }
                  >
                    {formatCentsToEuro(selectedPlayer.balance)}
                  </span>
                  {selectedPlayer.balance < 0 && (
                    <span> (Schulden: {formatCentsToEuro(Math.abs(selectedPlayer.balance))})</span>
                  )}
                </p>
              )}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="amount">
              Betrag (€) <span className="text-red-500">*</span>
            </Label>
            <Input
              id="amount"
              type="number"
              step="0.01"
              min="0.01"
              placeholder="10.00"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              required
            />
            {transactionType === "expense" && amount && parseFloat(amount) > 0 && (
              <div className="space-y-2">
                <p className="text-sm text-muted-foreground">
                  Club-Guthaben: {formatCentsToEuro(clubBalance)}
                </p>
                {euroToCents(parseFloat(amount)) > clubBalance && (
                  <div className="flex items-start gap-2 p-3 rounded-md bg-yellow-50 dark:bg-yellow-950 border border-yellow-200 dark:border-yellow-800">
                    <span className="mt-0.5" title="Warnung">
                      <AlertTriangleIcon className="size-4 text-yellow-600 dark:text-yellow-500" />
                    </span>
                    <p className="text-sm text-yellow-800 dark:text-yellow-200">
                      <strong>Warnung:</strong> Diese Ausgabe würde das Club-Guthaben auf{" "}
                      {formatCentsToEuro(clubBalance - euroToCents(parseFloat(amount)))} reduzieren.
                    </p>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Auto-Tip Preview */}
          {autoTipPreview && (
            <div className="rounded-md border border-blue-500 bg-blue-50 dark:bg-blue-950 p-4">
              <div className="flex items-start gap-3">
                <GiftIcon className="size-5 text-blue-500 mt-0.5" />
                <div className="flex-1">
                  <div className="font-medium text-blue-700 dark:text-blue-300">
                    Auto-Tip Vorschau
                  </div>
                  <div className="mt-2 space-y-1 text-sm text-blue-600 dark:text-blue-400">
                    <div>
                      Einzahlung: {formatCentsToEuro(autoTipPreview.depositAmount)} →
                      Balance wird auf €0,00 gesetzt
                    </div>
                    <div className="font-medium">
                      Auto-Tip: {formatCentsToEuro(autoTipPreview.tipAmount)} → Club
                      erhält Trinkgeld
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="description">
              Beschreibung <span className="text-red-500">*</span>
            </Label>
            <Textarea
              id="description"
              placeholder={getDefaultDescription()}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              required
            />
          </div>

          {transactionType === "deposit" && autoTipEnabled && (
            <div className="rounded-md border bg-muted p-3">
              <div className="flex items-start gap-2">
                <InfoIcon className="size-4 mt-0.5" />
                <p className="text-sm text-muted-foreground">
                  Auto-Tip ist für diesen Club aktiviert. Überschüssige Einzahlungen
                  werden automatisch als Trinkgeld verbucht.
                </p>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Abbrechen
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Erstelle..." : "Erstellen"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
