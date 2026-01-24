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
import { Checkbox } from "@/components/ui/checkbox";

import type { Role, CreateRoleRequest, UpdateRoleRequest } from "@/types/role";

type RoleDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  role?: Role | null;
  clubId: string;
  onSuccess: () => void;
};

export function RoleDialog({
  open,
  onOpenChange,
  role,
  clubId,
  onSuccess,
}: RoleDialogProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [name, setName] = React.useState("");
  const [paysBaseFee, setPaysBaseFee] = React.useState(false);
  const isEdit = !!role;

  React.useEffect(() => {
    if (open && role) {
      setName(role.name);
      setPaysBaseFee(role.pays_base_fee);
    } else if (open && !role) {
      setName("");
      setPaysBaseFee(false);
    }
  }, [open, role]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error("Name ist erforderlich");
      return;
    }
    setIsSubmitting(true);

    try {
      const body: CreateRoleRequest | UpdateRoleRequest = {
        name: name.trim(),
        pays_base_fee: paysBaseFee,
      };

      const url = isEdit
        ? `/api/clubs/${clubId}/roles/${role.id}`
        : `/api/clubs/${clubId}/roles`;

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
        isEdit ? "Rolle erfolgreich aktualisiert" : "Rolle erfolgreich erstellt"
      );
      onSuccess();
      onOpenChange(false);
      setName("");
      setPaysBaseFee(false);
    } catch (error) {
      console.error("Error saving role:", error);
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
            {isEdit ? "Rolle bearbeiten" : "Neue Rolle erstellen"}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? "Ändern Sie die Daten der Rolle."
              : "Erstellen Sie eine neue Rolle für den Club."}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              placeholder="z.B. Vorstand, Mitglied, Gast"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="pays_base_fee"
              checked={paysBaseFee}
              onCheckedChange={(checked) =>
                setPaysBaseFee(checked === true)
              }
            />
            <Label
              htmlFor="pays_base_fee"
              className="text-sm font-normal cursor-pointer"
            >
              Bezahlt Grundgebühr
            </Label>
          </div>
          <p className="text-sm text-muted-foreground">
            Legt fest, ob Mitglieder mit dieser Rolle die Club-Grundgebühr zahlen müssen.
          </p>
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

