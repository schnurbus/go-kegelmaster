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
import { InfoIcon } from "lucide-react";

import type { Transaction } from "@/types/transaction";
import type { Player } from "@/types/player";
import { euroToCents, formatCentsToEuro, formatMoneyInput } from "@/types/player";

type ManualType = "deposit" | "tip" | "expense";

/** Body of PUT /api/clubs/:clubId/transactions/:id */
type UpdateTransactionRequest = {
  transaction_type: ManualType;
  player_id?: string;
  amount: number; // in cents, always positive
  description: string;
  transaction_date: string; // YYYY-MM-DD
};

type EditTransactionDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clubId: string;
  transaction: Transaction | null;
  onSuccess: () => void;
};

function toDateInput(value: string | undefined): string {
  if (value) {
    const match = /^(\d{4}-\d{2}-\d{2})/.exec(value);
    if (match) return match[1];
  }
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function csrfToken(): string {
  return (
    document.cookie
      .split("; ")
      .find((row) => row.startsWith("csrf_token="))
      ?.split("=")[1] || ""
  );
}

export function EditTransactionDialog({
  open,
  onOpenChange,
  clubId,
  transaction,
  onSuccess,
}: EditTransactionDialogProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [players, setPlayers] = React.useState<Player[]>([]);
  const [isLoadingPlayers, setIsLoadingPlayers] = React.useState(false);

  const [transactionType, setTransactionType] = React.useState<ManualType>("deposit");
  const [playerId, setPlayerId] = React.useState("");
  const [amount, setAmount] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [transactionDate, setTransactionDate] = React.useState(toDateInput(undefined));

  // Prefill the form from the transaction whenever the dialog opens.
  React.useEffect(() => {
    if (!open || !transaction) return;
    setTransactionType(transaction.transaction_type as ManualType);
    setPlayerId(transaction.player_id ?? "");
    setAmount((Math.abs(transaction.amount) / 100).toFixed(2));
    setDescription(transaction.description ?? "");
    setTransactionDate(toDateInput(transaction.transaction_date ?? transaction.created_at));
  }, [open, transaction]);

  React.useEffect(() => {
    if (!open) return;
    let cancelled = false;
    setIsLoadingPlayers(true);
    fetch(`/api/clubs/${clubId}/players`, { credentials: "include" })
      .then((res) => {
        if (!res.ok) throw new Error("Fehler beim Laden der Spieler");
        return res.json() as Promise<Player[]>;
      })
      .then((data) => {
        if (cancelled) return;
        setPlayers([...data].sort((a, b) => a.name.localeCompare(b.name, "de")));
      })
      .catch((error) => {
        console.error("Error fetching players:", error);
        if (!cancelled) toast.error("Fehler beim Laden der Spieler");
      })
      .finally(() => {
        if (!cancelled) setIsLoadingPlayers(false);
      });
    return () => {
      cancelled = true;
    };
  }, [open, clubId]);

  const needsPlayer = transactionType === "deposit" || transactionType === "tip";

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!transaction) return;

    if (!amount || parseFloat(amount) <= 0) {
      toast.error("Betrag muss größer als 0 sein");
      return;
    }
    if (needsPlayer && !playerId) {
      toast.error("Bitte wählen Sie einen Spieler aus");
      return;
    }

    const body: UpdateTransactionRequest = {
      transaction_type: transactionType,
      amount: euroToCents(parseFloat(amount)),
      description: description.trim(),
      transaction_date: transactionDate,
    };
    if (needsPlayer) {
      body.player_id = playerId;
    }

    setIsSubmitting(true);
    try {
      const response = await fetch(
        `/api/clubs/${clubId}/transactions/${transaction.id}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": csrfToken(),
          },
          credentials: "include",
          body: JSON.stringify(body),
        }
      );
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Bearbeiten der Transaktion");
      }
      toast.success("Transaktion gespeichert");
      onSuccess();
      onOpenChange(false);
    } catch (error) {
      console.error("Error updating transaction:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Bearbeiten der Transaktion"
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[90dvh] flex-col sm:max-w-[500px]">
        <DialogHeader className="shrink-0">
          <DialogTitle>Transaktion bearbeiten</DialogTitle>
          <DialogDescription>
            Korrigieren Sie eine manuelle Transaktion. Die Salden von Spieler und Klub
            werden automatisch angepasst.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex min-h-0 flex-1 flex-col">
          <div className="min-h-0 flex-1 space-y-4 overflow-y-auto">
            <div className="space-y-2">
              <Label htmlFor="edit_transaction_type">
                Transaktionstyp <span className="text-red-500">*</span>
              </Label>
              <Select
                value={transactionType}
                onValueChange={(value: ManualType) => setTransactionType(value)}
                required
              >
                <SelectTrigger id="edit_transaction_type">
                  <SelectValue placeholder="Typ auswählen" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="deposit">Einzahlung</SelectItem>
                  <SelectItem value="tip">Trinkgeld</SelectItem>
                  <SelectItem value="expense">Ausgabe</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit_transaction_date">
                Transaktionsdatum <span className="text-red-500">*</span>
              </Label>
              <Input
                id="edit_transaction_date"
                type="date"
                value={transactionDate}
                onChange={(e) => setTransactionDate(e.target.value)}
                required
              />
            </div>

            {needsPlayer && (
              <div className="space-y-2">
                <Label htmlFor="edit_player_id">
                  Spieler <span className="text-red-500">*</span>
                </Label>
                <Select
                  value={playerId}
                  onValueChange={setPlayerId}
                  required
                  disabled={isLoadingPlayers}
                >
                  <SelectTrigger id="edit_player_id">
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
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="edit_amount">
                Betrag (€) <span className="text-red-500">*</span>
              </Label>
              <Input
                id="edit_amount"
                type="number"
                step="0.01"
                min="0.01"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                onBlur={() => setAmount(formatMoneyInput(amount))}
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit_description">Beschreibung (optional)</Label>
              <Textarea
                id="edit_description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
              />
            </div>

            <div className="rounded-md border bg-muted p-3">
              <div className="flex items-start gap-2">
                <InfoIcon className="size-4 mt-0.5" />
                <p className="text-sm text-muted-foreground">
                  Der Betrag wird genau so gebucht – beim Bearbeiten gibt es keine
                  Auto-Tip-Aufteilung und keinen Paar-Modus.
                </p>
              </div>
            </div>
          </div>
          <DialogFooter className="shrink-0 gap-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Abbrechen
            </Button>
            <Button type="submit" disabled={isSubmitting || !transaction}>
              {isSubmitting ? "Speichere..." : "Speichern"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
