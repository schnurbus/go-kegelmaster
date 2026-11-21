import * as React from "react";
import { toast } from "sonner";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

import type { Role } from "@/types/role";
import { csrfTokenFromCookie } from "@/lib/csrf";

type DeleteRoleDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  role: Role | null;
  clubId: string;
  onSuccess: () => void;
};

export function DeleteRoleDialog({
  open,
  onOpenChange,
  role,
  clubId,
  onSuccess,
}: DeleteRoleDialogProps) {
  const [isDeleting, setIsDeleting] = React.useState(false);
  // player_count is only known from the role list; the server rejects the
  // deletion with 409 in any case.
  const playerCount = role?.player_count ?? 0;
  const inUse = playerCount > 0;

  const handleDelete = async () => {
    if (!role) return;

    setIsDeleting(true);

    try {
      const response = await fetch(
        `/api/clubs/${clubId}/roles/${role.id}`,
        {
          method: "DELETE",
          headers: {
            "X-CSRF-Token": csrfTokenFromCookie(),
          },
          credentials: "include",
        }
      );

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Löschen");
      }

      toast.success("Rolle erfolgreich gelöscht");
      onSuccess();
      onOpenChange(false);
    } catch (error) {
      console.error("Error deleting role:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Löschen"
      );
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {inUse ? "Rolle kann nicht gelöscht werden" : "Rolle wirklich löschen?"}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {inUse
              ? `Die Rolle "${role?.name}" hat noch ${playerCount} Spieler. Jeder Spieler braucht eine Rolle. Ordnen Sie ${playerCount === 1 ? "ihn" : "diese"} zuerst einer anderen Rolle zu.`
              : `Möchten Sie die Rolle "${role?.name}" wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.`}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>
            Abbrechen
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={isDeleting || inUse}
            className="bg-red-600 hover:bg-red-700"
          >
            {isDeleting ? "Lösche..." : "Löschen"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

