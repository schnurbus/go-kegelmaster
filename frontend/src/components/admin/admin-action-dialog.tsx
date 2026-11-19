import { useId, useState, type ReactNode } from "react";
import { toast } from "sonner";

import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type AdminActionInput = {
  label: string;
  placeholder?: string;
  type?: "text" | "email";
  /**
   * Confirmation by typing: the action stays disabled until the value matches
   * (e.g. the e-mail address or club name of the object to delete).
   */
  mustMatch?: string;
  /** Compare mustMatch case-insensitively (e-mail addresses). */
  ignoreCase?: boolean;
};

/**
 * Confirmation dialog for an admin intervention. Optionally asks for a value
 * (confirmation text or e.g. the new owner's e-mail). Errors of onConfirm are
 * shown as toast; the dialog stays open then.
 */
export function AdminActionDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  destructive = false,
  input,
  successMessage,
  onConfirm,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: ReactNode;
  confirmLabel: string;
  destructive?: boolean;
  input?: AdminActionInput;
  successMessage: string;
  onConfirm: (value: string) => Promise<void>;
}) {
  const inputId = useId();
  const [value, setValue] = useState("");
  const [pending, setPending] = useState(false);

  const trimmed = value.trim();
  const matches =
    !input ||
    (input.mustMatch === undefined
      ? trimmed !== ""
      : input.ignoreCase
        ? trimmed.toLowerCase() === input.mustMatch.trim().toLowerCase()
        : trimmed === input.mustMatch.trim());

  const handleOpenChange = (next: boolean) => {
    if (pending) return;
    if (!next) setValue("");
    onOpenChange(next);
  };

  const handleConfirm = async () => {
    if (!matches || pending) return;
    setPending(true);
    try {
      await onConfirm(trimmed);
      toast.success(successMessage);
      setValue("");
      onOpenChange(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Aktion fehlgeschlagen");
    } finally {
      setPending(false);
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent>
        <form
          className="grid gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            void handleConfirm();
          }}
        >
          <AlertDialogHeader>
            <AlertDialogTitle>{title}</AlertDialogTitle>
            <AlertDialogDescription asChild>
              <div className="space-y-2">{description}</div>
            </AlertDialogDescription>
          </AlertDialogHeader>
          {input && (
            <div className="grid gap-2">
              <Label htmlFor={inputId}>{input.label}</Label>
              <Input
                id={inputId}
                type={input.type ?? "text"}
                value={value}
                placeholder={input.placeholder}
                autoComplete="off"
                onChange={(e) => setValue(e.target.value)}
                disabled={pending}
              />
            </div>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel type="button" disabled={pending}>
              Abbrechen
            </AlertDialogCancel>
            <Button type="submit" variant={destructive ? "destructive" : "default"} disabled={!matches || pending}>
              {pending ? "Bitte warten …" : confirmLabel}
            </Button>
          </AlertDialogFooter>
        </form>
      </AlertDialogContent>
    </AlertDialog>
  );
}
