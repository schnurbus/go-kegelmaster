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

import type { PenaltyType } from "@/types/penalty-type";

type PenaltyTypeDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  penaltyType?: PenaltyType | null;
  clubId: string;
  onSuccess: () => void;
};

export function PenaltyTypeDialog({
  open,
  onOpenChange,
  penaltyType,
  clubId,
  onSuccess,
}: PenaltyTypeDialogProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [name, setName] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [price, setPrice] = React.useState("");
  const [displayOrder, setDisplayOrder] = React.useState("");
  const [allowsDecimalQuantity, setAllowsDecimalQuantity] = React.useState(false);
  const isEdit = !!penaltyType;

  React.useEffect(() => {
    if (open && penaltyType) {
      setName(penaltyType.name);
      setDescription(penaltyType.description || "");
      setPrice((penaltyType.price / 100).toFixed(2));
      setDisplayOrder(penaltyType.display_order.toString());
      setAllowsDecimalQuantity(penaltyType.allows_decimal_quantity ?? false);
    } else if (open && !penaltyType) {
      setName("");
      setDescription("");
      setPrice("");
      setDisplayOrder("");
      setAllowsDecimalQuantity(false);
    }
  }, [open, penaltyType]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error("Name ist erforderlich");
      return;
    }

    const priceValue = parseFloat(price.replace(",", "."));
    if (isNaN(priceValue) || priceValue < 0) {
      toast.error("Ungültiger Preis");
      return;
    }

    setIsSubmitting(true);

    try {
    const body: {
      name: string;
      description: string;
      price: number;
      display_order?: number;
      allows_decimal_quantity: boolean;
    } = {
      name: name.trim(),
      description: description.trim(),
      price: Math.round(priceValue * 100), // Convert to cents
      allows_decimal_quantity: allowsDecimalQuantity,
    };

    // Add display_order only if provided (for new entries)
    if (!isEdit && displayOrder.trim() !== "") {
      const orderValue = parseInt(displayOrder.trim(), 10);
      if (!isNaN(orderValue) && orderValue >= 0) {
        body.display_order = orderValue;
      }
    }

      const url = isEdit
        ? `/api/clubs/${clubId}/penalty-types/${penaltyType.id}`
        : `/api/clubs/${clubId}/penalty-types`;

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
        isEdit
          ? "Strafentyp erfolgreich aktualisiert"
          : "Strafentyp erfolgreich erstellt"
      );
      onSuccess();
      onOpenChange(false);
      setName("");
      setDescription("");
      setPrice("");
      setDisplayOrder("");
      setAllowsDecimalQuantity(false);
    } catch (error) {
      console.error("Error saving penalty type:", error);
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
            {isEdit ? "Strafentyp bearbeiten" : "Neuen Strafentyp erstellen"}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? "Ändern Sie die Daten des Strafentyps. Eine neue Version wird erstellt."
              : "Erstellen Sie einen neuen Strafentyp für den Club."}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              placeholder="z.B. Verspätung, Fehlwurf"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Beschreibung</Label>
            <Input
              id="description"
              placeholder="Optionale Beschreibung des Strafentyps"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="price">Preis (€) *</Label>
            <Input
              id="price"
              type="number"
              step="0.01"
              min="0"
              placeholder="0.00"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              required
            />
          </div>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="allows_decimal_quantity"
              checked={allowsDecimalQuantity}
              onCheckedChange={(checked) => setAllowsDecimalQuantity(checked === true)}
            />
            <Label htmlFor="allows_decimal_quantity" className="text-sm font-normal cursor-pointer">
              Dezimalanzahl erlauben (z.B. 2,5 Stück)
            </Label>
          </div>
          {!isEdit && (
            <div className="space-y-2">
              <Label htmlFor="display_order">Reihenfolge (optional)</Label>
              <Input
                id="display_order"
                type="number"
                min="0"
                placeholder="Leer lassen für automatische Position"
                value={displayOrder}
                onChange={(e) => setDisplayOrder(e.target.value)}
              />
              <p className="text-sm text-muted-foreground">
                Wenn leer gelassen, wird der Strafentyp automatisch am Ende
                eingefügt.
              </p>
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

