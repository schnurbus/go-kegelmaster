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
import { useAuth } from "@/context/AuthContext";

type DeleteClubDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clubId: string;
  clubName: string;
  onSuccess: () => void;
};

export function DeleteClubDialog({
  open,
  onOpenChange,
  clubId,
  clubName,
  onSuccess,
}: DeleteClubDialogProps) {
  const { csrfToken, refreshCsrf } = useAuth();
  const [password, setPassword] = React.useState("");
  const [isSubmitting, setIsSubmitting] = React.useState(false);

  React.useEffect(() => {
    if (!open) {
      setPassword("");
    }
  }, [open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!password) {
      toast.error("Passwort zur Bestätigung ist erforderlich");
      return;
    }

    setIsSubmitting(true);
    try {
      const token = csrfToken ?? (await refreshCsrf());
      const response = await fetch(`/api/clubs/${clubId}`, {
        method: "DELETE",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify({ password }),
      });

      if (response.status === 204) {
        toast.success("Club wurde gelöscht");
        onSuccess();
        onOpenChange(false);
        return;
      }

      const data = await response.json().catch(() => ({}));
      throw new Error(
        (data as { message?: string }).message ?? "Fehler beim Löschen"
      );
    } catch (error) {
      console.error("Delete club:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Löschen"
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Club löschen</DialogTitle>
          <DialogDescription>
            Möchten Sie den Club „{clubName}“ wirklich löschen? Alle
            zugehörigen Daten (Spieler, Spieltage, Transaktionen usw.) werden
            unwiderruflich gelöscht. Geben Sie zur Bestätigung Ihr Passwort ein.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="delete-club-password">
              Ihr Passwort zur Bestätigung
            </Label>
            <Input
              id="delete-club-password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isSubmitting}
              required
            />
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
            <Button
              type="submit"
              variant="destructive"
              disabled={isSubmitting}
            >
              {isSubmitting ? "Lösche…" : "Club löschen"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
