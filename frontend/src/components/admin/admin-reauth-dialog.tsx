import { useEffect, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { KeyRoundIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { PasswordInput } from "@/components/ui/password-input";
import { useAuth } from "@/context/AuthContext";
import { onAdminReauthRequired, reauthWithPasskey, reauthWithPassword } from "@/lib/admin-reauth";
import { PasskeyCancelledError, passkeysSupported } from "@/lib/passkey";

/**
 * Asks an instance admin to confirm the login again (password or passkey) when
 * an admin request answered "reauth_required". `onConfirmed` runs after a
 * successful confirmation, so the admin pages can load their data again.
 */
export function AdminReauthDialog({ onConfirmed }: { onConfirmed: () => void }) {
  const { user, refreshCsrf } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [password, setPassword] = useState("");
  const [pending, setPending] = useState<"password" | "passkey" | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(
    () =>
      onAdminReauthRequired(() => {
        setPassword("");
        setError(null);
        setOpen(true);
      }),
    [],
  );

  const confirm = async (method: "password" | "passkey") => {
    setPending(method);
    setError(null);
    try {
      if (method === "password") {
        await reauthWithPassword(password);
      } else {
        await reauthWithPasskey();
      }
      // The server reissued the session and rotated the CSRF cookie.
      await refreshCsrf();
      setOpen(false);
      setPassword("");
      onConfirmed();
    } catch (err) {
      if (!(err instanceof PasskeyCancelledError)) {
        setError(err instanceof Error ? err.message : "Bestätigung fehlgeschlagen");
      }
    } finally {
      setPending(null);
    }
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!password || pending) return;
    void confirm("password");
  };

  const cancel = () => {
    setOpen(false);
    navigate("/app");
  };

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? setOpen(true) : cancel())}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Anmeldung bestätigen</DialogTitle>
          <DialogDescription>
            Der Admin-Bereich erfordert eine aktuelle Anmeldung. Bitte bestätigen Sie, dass Sie es sind
            {user?.email ? ` (${user.email})` : ""}.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="admin-reauth-password">Passwort</Label>
            <PasswordInput
              id="admin-reauth-password"
              autoComplete="current-password"
              autoFocus
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={pending !== null}
            />
          </div>
          {error && (
            <p role="alert" className="text-sm text-destructive">
              {error}
            </p>
          )}
          <DialogFooter className="gap-2 sm:justify-between">
            {passkeysSupported() ? (
              <Button
                type="button"
                variant="outline"
                onClick={() => void confirm("passkey")}
                disabled={pending !== null}
              >
                <KeyRoundIcon />
                {pending === "passkey" ? "Warte auf Passkey …" : "Mit Passkey bestätigen"}
              </Button>
            ) : (
              <span />
            )}
            <div className="flex flex-col-reverse gap-2 sm:flex-row">
              <Button type="button" variant="ghost" onClick={cancel} disabled={pending !== null}>
                Abbrechen
              </Button>
              <Button type="submit" disabled={!password || pending !== null}>
                {pending === "password" ? "Wird geprüft …" : "Bestätigen"}
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
