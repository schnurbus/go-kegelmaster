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

import type { Competition } from "@/types/competition";

type DeleteCompetitionDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  competition: Competition | null;
  clubId: string;
  onSuccess: () => void;
};

export function DeleteCompetitionDialog({
  open,
  onOpenChange,
  competition,
  clubId,
  onSuccess,
}: DeleteCompetitionDialogProps) {
  const [isDeleting, setIsDeleting] = React.useState(false);

  const handleDelete = async () => {
    if (!competition) return;

    setIsDeleting(true);

    try {
      const response = await fetch(
        `/api/clubs/${clubId}/competitions/${competition.id}`,
        {
          method: "DELETE",
          headers: {
            "X-CSRF-Token": document.cookie
              .split("; ")
              .find((row) => row.startsWith("csrf_token="))
              ?.split("=")[1] || "",
          },
          credentials: "include",
        }
      );

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Löschen");
      }

      toast.success("Wettbewerb erfolgreich gelöscht");
      onSuccess();
      onOpenChange(false);
    } catch (error) {
      console.error("Error deleting competition:", error);
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
          <AlertDialogTitle>Wettbewerb wirklich löschen?</AlertDialogTitle>
          <AlertDialogDescription>
            Möchten Sie den Wettbewerb "{competition?.name}" wirklich löschen?
            Diese Aktion kann nicht rückgängig gemacht werden.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>
            Abbrechen
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={isDeleting}
            className="bg-red-600 hover:bg-red-700"
          >
            {isDeleting ? "Lösche..." : "Löschen"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
