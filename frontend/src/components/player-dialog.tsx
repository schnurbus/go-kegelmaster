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

import type { Player, Role, CreatePlayerRequest, UpdatePlayerRequest } from "@/types/player";
import { euroToCents } from "@/types/player";

type PlayerDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  player?: Player | null;
  clubId: string;
  roles: Role[];
  onSuccess: () => void;
};

export function PlayerDialog({
  open,
  onOpenChange,
  player,
  clubId,
  roles,
  onSuccess,
}: PlayerDialogProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [name, setName] = React.useState("");
  const [balance, setBalance] = React.useState("0");
  const [startBalance, setStartBalance] = React.useState("0");
  const [roleId, setRoleId] = React.useState<string | null>(null);
  const isEdit = !!player;

  React.useEffect(() => {
    if (open && player) {
      setName(player.name);
      setBalance((player.balance / 100).toString());
      setStartBalance((player.start_balance / 100).toString());
      setRoleId(player.role_id);
    } else if (open && !player) {
      setName("");
      setBalance("0");
      setStartBalance("0");
      // Pre-select first role when creating new player
      setRoleId(roles.length > 0 ? roles[0].id : null);
    }
  }, [open, player, roles]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error("Name ist erforderlich");
      return;
    }

    if (!roleId) {
      toast.error("Rolle ist erforderlich");
      return;
    }
    setIsSubmitting(true);

    try {
      const body: CreatePlayerRequest | UpdatePlayerRequest = {
        name: name.trim(),
        balance: euroToCents(parseFloat(balance) || 0),
        start_balance: euroToCents(parseFloat(startBalance) || 0),
        role_id: roleId!, // role_id is required and validated above
        user_id: null,
      };

      const url = isEdit
        ? `/api/clubs/${clubId}/players/${player.id}`
        : `/api/clubs/${clubId}/players`;

      const response = await fetch(url, {
        method: isEdit ? "PUT" : "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": document.cookie
            .split("; ")
            .find((row) => row.startsWith("csrf_token="))
            ?.split("=")[1] || "",
        },
        credentials: "include",
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Speichern");
      }

      toast.success(
        isEdit ? "Player erfolgreich aktualisiert" : "Player erfolgreich erstellt"
      );
      onSuccess();
      onOpenChange(false);
      // Reset form
      setName("");
      setBalance("0");
      setStartBalance("0");
      setRoleId(roles.length > 0 ? roles[0].id : null);
    } catch (error) {
      console.error("Error saving player:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Speichern"
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? "Player bearbeiten" : "Neuen Player erstellen"}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? "Ändern Sie die Daten des Players."
              : "Fügen Sie einen neuen Player zum Club hinzu."}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              placeholder="Max Mustermann"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="balance">Balance (€)</Label>
              <Input
                id="balance"
                type="number"
                step="0.01"
                placeholder="0.00"
                value={balance}
                onChange={(e) => setBalance(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="start_balance">Start-Balance (€)</Label>
              <Input
                id="start_balance"
                type="number"
                step="0.01"
                placeholder="0.00"
                value={startBalance}
                onChange={(e) => setStartBalance(e.target.value)}
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="role_id">
              Rolle <span className="text-red-500">*</span>
            </Label>
            <Select
              value={roleId || undefined}
              onValueChange={(value) => {
                setRoleId(value);
              }}
              required
            >
              <SelectTrigger id="role_id">
                <SelectValue placeholder="Rolle auswählen" />
              </SelectTrigger>
              <SelectContent>
                {roles.map((role) => (
                  <SelectItem key={role.id} value={role.id}>
                    {role.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-sm text-muted-foreground">
              Jeder Spieler muss einer Rolle zugeordnet sein
            </p>
          </div>
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
              {isSubmitting
                ? "Speichere..."
                : isEdit
                  ? "Speichern"
                  : "Erstellen"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

