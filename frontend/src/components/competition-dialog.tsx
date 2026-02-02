import * as React from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
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

import type { Competition, ScoringType } from "@/types/competition";

const SCORING_OPTIONS: { value: ScoringType; label: string }[] = [
  { value: "winner", label: "Gewinner" },
  { value: "loser", label: "Verlierer" },
  { value: "both", label: "Beides" },
];

type CompetitionDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  competition?: Competition | null;
  clubId: string;
  onSuccess: () => void;
};

export function CompetitionDialog({
  open,
  onOpenChange,
  competition,
  clubId,
  onSuccess,
}: CompetitionDialogProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [name, setName] = React.useState("");
  const [scoringType, setScoringType] = React.useState<ScoringType>("winner");
  const [isGenderSpecific, setIsGenderSpecific] = React.useState(false);
  const [displayOrder, setDisplayOrder] = React.useState("");
  const isEdit = !!competition;

  React.useEffect(() => {
    if (open && competition) {
      setName(competition.name);
      setScoringType(competition.scoring_type);
      setIsGenderSpecific(competition.is_gender_specific);
      setDisplayOrder(competition.display_order.toString());
    } else if (open && !competition) {
      setName("");
      setScoringType("winner");
      setIsGenderSpecific(false);
      setDisplayOrder("");
    }
  }, [open, competition]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error("Name ist erforderlich");
      return;
    }

    setIsSubmitting(true);

    try {
      const body: {
        name: string;
        scoring_type: ScoringType;
        is_gender_specific: boolean;
        display_order?: number;
      } = {
        name: name.trim(),
        scoring_type: scoringType,
        is_gender_specific: isGenderSpecific,
      };

      if (!isEdit && displayOrder.trim() !== "") {
        const orderValue = parseInt(displayOrder.trim(), 10);
        if (!isNaN(orderValue) && orderValue >= 0) {
          body.display_order = orderValue;
        }
      }

      const url = isEdit
        ? `/api/clubs/${clubId}/competitions/${competition.id}`
        : `/api/clubs/${clubId}/competitions`;

      const method = isEdit ? "PUT" : "POST";
      const putBody = isEdit
        ? {
            name: body.name,
            scoring_type: body.scoring_type,
            is_gender_specific: body.is_gender_specific,
            display_order: parseInt(displayOrder.trim(), 10) || 0,
          }
        : body;

      const response = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": document.cookie
            .split("; ")
            .find((row) => row.startsWith("csrf_token="))
            ?.split("=")[1] || "",
        },
        credentials: "include",
        body: JSON.stringify(putBody),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Speichern");
      }

      toast.success(
        isEdit
          ? "Wettbewerb erfolgreich aktualisiert"
          : "Wettbewerb erfolgreich erstellt"
      );
      onSuccess();
      onOpenChange(false);
    } catch (error) {
      console.error("Error saving competition:", error);
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
            {isEdit ? "Wettbewerb bearbeiten" : "Neuen Wettbewerb erstellen"}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? "Ändern Sie die Daten des Wettbewerbs."
              : "Erstellen Sie einen neuen Wettbewerb für den Club."}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              placeholder="z.B. Höchste Wurfzahl"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="scoring_type">Wertung *</Label>
            <Select
              value={scoringType}
              onValueChange={(v) => setScoringType(v as ScoringType)}
            >
              <SelectTrigger id="scoring_type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SCORING_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="is_gender_specific"
              checked={isGenderSpecific}
              onCheckedChange={(checked) =>
                setIsGenderSpecific(checked === true)
              }
            />
            <Label
              htmlFor="is_gender_specific"
              className="text-sm font-normal cursor-pointer"
            >
              Geschlechtsspezifisch (bei Auswertung Sieger/Verlierer pro
              Geschlecht)
            </Label>
          </div>
          <div className="space-y-2">
            <Label htmlFor="display_order">
              Reihenfolge{!isEdit ? " (optional)" : ""}
            </Label>
            <Input
              id="display_order"
              type="number"
              min="0"
              placeholder={!isEdit ? "Leer lassen für automatische Position" : undefined}
              value={displayOrder}
              onChange={(e) => setDisplayOrder(e.target.value)}
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
