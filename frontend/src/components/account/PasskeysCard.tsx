import * as React from "react";
import { FingerprintIcon, PencilIcon, PlusIcon, Trash2Icon } from "lucide-react";
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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Label } from "@/components/ui/label";
import {
  type Passkey,
  PasskeyCancelledError,
  createPasskey,
  passkeysSupported,
  suggestPasskeyName,
} from "@/lib/passkey";
import { csrfTokenFromCookie } from "@/lib/csrf";

const MAX_NAME_LENGTH = 100;

function formatDate(value: string | null): string {
  if (!value) return "Noch nie";
  return new Date(value).toLocaleString("de-DE", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

async function sendJSON(method: string, url: string, body?: unknown): Promise<Response> {
  const response = await fetch(url, {
    method,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrfTokenFromCookie(),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.message || "Anfrage fehlgeschlagen");
  }
  return response;
}

export function PasskeysCard() {
  const supported = passkeysSupported();
  const [passkeys, setPasskeys] = React.useState<Passkey[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);

  const [isAddOpen, setIsAddOpen] = React.useState(false);
  const [name, setName] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [addError, setAddError] = React.useState<string | null>(null);
  const [isAdding, setIsAdding] = React.useState(false);

  const [toRename, setToRename] = React.useState<Passkey | null>(null);
  const [newName, setNewName] = React.useState("");
  const [isRenaming, setIsRenaming] = React.useState(false);

  const [toDelete, setToDelete] = React.useState<Passkey | null>(null);
  const [isDeleting, setIsDeleting] = React.useState(false);

  const fetchPasskeys = React.useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await fetch("/api/auth/passkeys", { credentials: "include" });
      if (!response.ok) throw new Error();
      setPasskeys(await response.json());
    } catch {
      toast.error("Fehler beim Laden der Passkeys");
      setPasskeys([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  React.useEffect(() => {
    fetchPasskeys();
  }, [fetchPasskeys]);

  const openAdd = () => {
    setName(suggestPasskeyName());
    setPassword("");
    setAddError(null);
    setIsAddOpen(true);
  };

  const handleAdd = async (event: React.FormEvent) => {
    event.preventDefault();
    setAddError(null);
    setIsAdding(true);
    try {
      const begin = await sendJSON("POST", "/api/auth/passkeys/register/begin", {
        current_password: password,
      });
      const credential = await createPasskey(await begin.json());
      await sendJSON("POST", "/api/auth/passkeys/register/finish", {
        name: name.trim(),
        credential,
      });
      toast.success("Passkey hinzugefügt");
      setIsAddOpen(false);
      fetchPasskeys();
    } catch (error) {
      setAddError(
        error instanceof PasskeyCancelledError
          ? "Der Vorgang wurde abgebrochen. Bitte erneut versuchen."
          : error instanceof Error
            ? error.message
            : "Passkey konnte nicht hinzugefügt werden",
      );
    } finally {
      setPassword("");
      setIsAdding(false);
    }
  };

  const openRename = (passkey: Passkey) => {
    setNewName(passkey.name);
    setToRename(passkey);
  };

  const handleRename = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!toRename) return;
    setIsRenaming(true);
    try {
      await sendJSON("PATCH", `/api/auth/passkeys/${toRename.id}`, { name: newName.trim() });
      toast.success("Passkey umbenannt");
      setToRename(null);
      fetchPasskeys();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Fehler beim Umbenennen");
    } finally {
      setIsRenaming(false);
    }
  };

  const handleDelete = async () => {
    if (!toDelete) return;
    setIsDeleting(true);
    try {
      await sendJSON("DELETE", `/api/auth/passkeys/${toDelete.id}`);
      toast.success("Passkey entfernt");
      setToDelete(null);
      fetchPasskeys();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Fehler beim Entfernen");
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Card>
      <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1.5">
          <CardTitle>Passkeys</CardTitle>
          <CardDescription className="max-w-2xl">
            Mit einem Passkey melden Sie sich per Fingerabdruck, Gesichtserkennung
            oder Geräte-PIN an – ohne Passwort. Ihr Passwort bleibt zusätzlich
            gültig. Beim Zurücksetzen des Passworts über „Passwort vergessen“
            werden alle Passkeys entfernt.
          </CardDescription>
        </div>
        {supported && (
          <Button onClick={openAdd} className="w-full shrink-0 sm:w-auto">
            <PlusIcon />
            Passkey hinzufügen
          </Button>
        )}
      </CardHeader>
      <CardContent className="space-y-3">
        {!supported && (
          <p className="text-sm text-muted-foreground">
            Dieser Browser unterstützt keine Passkeys. Bitte verwenden Sie eine
            aktuelle Version von Chrome, Edge, Firefox oder Safari.
          </p>
        )}
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Lädt …</p>
        ) : passkeys.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-6 text-center text-sm text-muted-foreground">
            <FingerprintIcon className="size-8" />
            Noch keine Passkeys eingerichtet.
          </div>
        ) : (
          <ul className="divide-y" aria-label="Passkeys">
            {passkeys.map((passkey) => (
              <li key={passkey.id} className="flex items-start gap-2 py-3 first:pt-0 last:pb-0">
                <div className="min-w-0 flex-1 space-y-1 text-sm">
                  <p className="flex flex-wrap items-center gap-2">
                    <span className="break-words font-medium">{passkey.name}</span>
                    <Badge variant="secondary">
                      {passkey.synced ? "Synchronisiert" : "Nur dieses Gerät"}
                    </Badge>
                  </p>
                  <dl className="grid grid-cols-[auto_1fr] gap-x-3 text-xs text-muted-foreground">
                    <dt>Hinzugefügt</dt>
                    <dd>{formatDate(passkey.created_at)}</dd>
                    <dt>Zuletzt genutzt</dt>
                    <dd>{formatDate(passkey.last_used_at)}</dd>
                  </dl>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Passkey ${passkey.name} umbenennen`}
                  onClick={() => openRename(passkey)}
                >
                  <PencilIcon />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Passkey ${passkey.name} entfernen`}
                  onClick={() => setToDelete(passkey)}
                >
                  <Trash2Icon />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </CardContent>

      <Dialog open={isAddOpen} onOpenChange={(open) => !isAdding && setIsAddOpen(open)}>
        <DialogContent>
          <form onSubmit={handleAdd} className="space-y-4">
            <DialogHeader>
              <DialogTitle>Passkey hinzufügen</DialogTitle>
              <DialogDescription>
                Bestätigen Sie zuerst Ihr Passwort. Danach fragt Ihr Browser bzw.
                Ihr Gerät nach Fingerabdruck, Gesichtserkennung oder PIN.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              <Label htmlFor="passkey-name">Name</Label>
              <Input
                id="passkey-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="z. B. iPhone"
                maxLength={MAX_NAME_LENGTH}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="passkey-password">Aktuelles Passwort</Label>
              <PasswordInput
                id="passkey-password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                maxLength={128}
                required
                autoFocus
              />
            </div>
            {addError && (
              <p role="alert" className="text-sm text-destructive">
                {addError}
              </p>
            )}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsAddOpen(false)}
                disabled={isAdding}
              >
                Abbrechen
              </Button>
              <Button type="submit" disabled={isAdding || password === ""}>
                {isAdding ? "Warte auf Passkey …" : "Weiter"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={toRename !== null} onOpenChange={(open) => !open && setToRename(null)}>
        <DialogContent>
          <form onSubmit={handleRename} className="space-y-4">
            <DialogHeader>
              <DialogTitle>Passkey umbenennen</DialogTitle>
              <DialogDescription>Der Name hilft Ihnen, Ihre Geräte zu unterscheiden.</DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              <Label htmlFor="passkey-new-name">Name</Label>
              <Input
                id="passkey-new-name"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                maxLength={MAX_NAME_LENGTH}
                required
                autoFocus
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setToRename(null)}>
                Abbrechen
              </Button>
              <Button type="submit" disabled={isRenaming || newName.trim() === ""}>
                {isRenaming ? "Speichert …" : "Speichern"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <AlertDialog open={toDelete !== null} onOpenChange={(open) => !open && setToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Passkey entfernen?</AlertDialogTitle>
            <AlertDialogDescription>
              „{toDelete?.name}“ kann danach nicht mehr zur Anmeldung verwendet
              werden. Den Eintrag in Ihrem Passwort-Manager bzw. auf Ihrem Gerät
              können Sie anschließend dort löschen.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                handleDelete();
              }}
              disabled={isDeleting}
            >
              {isDeleting ? "Entfernt …" : "Entfernen"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  );
}
