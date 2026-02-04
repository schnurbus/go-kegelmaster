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

type TransferClubOwnerDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clubId: string;
  clubName: string;
  onSuccess: () => void;
};

export function TransferClubOwnerDialog({
  open,
  onOpenChange,
  clubId,
  clubName,
  onSuccess,
}: TransferClubOwnerDialogProps) {
  const { csrfToken, refreshCsrf } = useAuth();
  const [newOwnerEmail, setNewOwnerEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [isSubmitting, setIsSubmitting] = React.useState(false);

  React.useEffect(() => {
    if (!open) {
      setNewOwnerEmail("");
      setPassword("");
    }
  }, [open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newOwnerEmail.trim()) {
      toast.error("E-Mail des neuen Eigentümers ist erforderlich");
      return;
    }
    if (!password) {
      toast.error("Passwort zur Bestätigung ist erforderlich");
      return;
    }

    setIsSubmitting(true);
    try {
      const token = csrfToken ?? (await refreshCsrf());
      const response = await fetch(`/api/clubs/${clubId}/transfer-owner`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify({
          new_owner_email: newOwnerEmail.trim(),
          password,
        }),
      });

      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(
          (data as { message?: string }).message ?? "Fehler beim Übertragen"
        );
      }

      toast.success("Eigentümer wurde erfolgreich gewechselt");
      onSuccess();
      onOpenChange(false);
    } catch (error) {
      console.error("Transfer owner:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Übertragen"
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Eigentümer wechseln</DialogTitle>
          <DialogDescription>
            Übertragen Sie den Club „{clubName}“ an einen anderen User. Der neue
            Eigentümer muss bereits ein Konto haben. Geben Sie zur Bestätigung
            Ihr Passwort ein.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="transfer-new-owner-email">
              E-Mail des neuen Eigentümers
            </Label>
            <Input
              id="transfer-new-owner-email"
              type="email"
              autoComplete="email"
              value={newOwnerEmail}
              onChange={(e) => setNewOwnerEmail(e.target.value)}
              placeholder="email@beispiel.de"
              disabled={isSubmitting}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="transfer-password">Ihr Passwort zur Bestätigung</Label>
            <Input
              id="transfer-password"
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
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Übertrage…" : "Eigentümer übertragen"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
