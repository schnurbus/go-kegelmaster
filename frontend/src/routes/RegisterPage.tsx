import type { FormEvent } from "react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useRegistrationMode } from "@/hooks/use-public-config";
import type { InvitationResponse } from "@/types/invitation";
import { useAuth } from "../context/AuthContext";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { Label } from "@/components/ui/label";

const initialState = { email: "", password: "" };

function RegisterPage() {
  const [form, setForm] = useState(initialState);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { register } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const inviteToken = searchParams.get("invite_token");
  const registrationMode = useRegistrationMode();
  // E-Mail aus der Einladung; bei invite_only muss sie exakt passen.
  const [invitedEmail, setInvitedEmail] = useState<string | null>(null);
  const [invitationInvalid, setInvitationInvalid] = useState(false);

  useEffect(() => {
    if (!inviteToken) return;
    let cancelled = false;
    fetch(`/api/invitations/${encodeURIComponent(inviteToken)}`, { credentials: "include" })
      .then((res) => (res.ok ? (res.json() as Promise<InvitationResponse>) : Promise.reject(res.status)))
      .then((data) => {
        if (cancelled) return;
        setInvitedEmail(data.email);
        setForm((prev) => (prev.email ? prev : { ...prev, email: data.email }));
      })
      .catch(() => {
        if (!cancelled) setInvitationInvalid(true);
      });
    return () => {
      cancelled = true;
    };
  }, [inviteToken]);

  const lockEmail = registrationMode === "invite_only" && invitedEmail !== null;
  let blockedHint: string | null = null;
  if (registrationMode === "closed") {
    blockedHint = "Auf dieser Instanz können keine neuen Accounts angelegt werden.";
  } else if (registrationMode === "invite_only" && !inviteToken) {
    blockedHint =
      "Die Registrierung ist nur mit einer Einladung möglich. Bitte nutze den Link aus deiner Einladungs-E-Mail.";
  } else if (registrationMode === "invite_only" && invitationInvalid) {
    blockedHint =
      "Die Einladung ist ungültig, abgelaufen oder wurde bereits angenommen. Bitte fordere eine neue Einladung an.";
  }

  const acceptInvitation = async (token: string) => {
    try {
      toast.loading("Einladung wird akzeptiert...", { id: "accept-invite-after-register" });

      const csrfToken = document.cookie
        .split("; ")
        .find((row) => row.startsWith("csrf_token="))
        ?.split("=")[1] || "";

      if (!csrfToken) {
        throw new Error("CSRF-Token fehlt. Bitte Seite neu laden.");
      }

      const response = await fetch(`/api/invitations/${token}/accept`, {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrfToken,
        },
        credentials: "include",
      });

      const responseData = await response.json().catch(() => ({}));

      if (!response.ok) {
        const errorMessage = responseData.message || "Fehler beim Akzeptieren der Einladung";

        if (response.status === 404) {
          throw new Error("Einladung nicht gefunden");
        }
        if (response.status === 410) {
          throw new Error("Einladung ist abgelaufen oder wurde bereits akzeptiert");
        }

        throw new Error(errorMessage);
      }

      toast.success("Einladung erfolgreich akzeptiert!", {
        id: "accept-invite-after-register",
        duration: 4000,
      });
      return responseData.player_id;
    } catch (error) {
      console.error("Error accepting invitation:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Akzeptieren der Einladung",
        { id: "accept-invite-after-register", duration: 6000 },
      );
      throw error;
    }
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    try {
      await register(inviteToken ? { ...form, invite_token: inviteToken } : form);
      toast.success("Account erstellt");
      if (inviteToken) {
        try {
          const playerId = await acceptInvitation(inviteToken);
          navigate(`/app/players/${playerId}`, { replace: true });
        } catch (inviteError) {
          console.error("Failed to accept invitation:", inviteError);
          toast.warning(
            "Registrierung erfolgreich, aber Einladung konnte nicht akzeptiert werden. " +
              "Sie können die Einladung später erneut versuchen.",
            { duration: 8000 },
          );
          navigate("/app", { replace: true });
        }
      } else {
        navigate("/app", { replace: true });
      }
    } catch (err) {
      const message = (err as Error).message;
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (blockedHint) {
    return (
      <div className="min-h-[80vh] flex flex-col justify-center px-4 py-12">
        <div className="max-w-md mx-auto w-full">
          <Card className="border-border/50 bg-card/50">
            <CardHeader className="space-y-1">
              <CardTitle className="text-2xl text-foreground">Registrierung nicht möglich</CardTitle>
              <CardDescription className="text-muted-foreground" role="status">
                {blockedHint}
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <p className="text-center text-sm text-muted-foreground w-full">
                Bereits registriert?{" "}
                <Link
                  to={inviteToken ? `/login?invite_token=${encodeURIComponent(inviteToken)}` : "/login"}
                  className="text-primary hover:underline font-medium"
                >
                  Zum Login
                </Link>
              </p>
            </CardFooter>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-[80vh] flex flex-col justify-center px-4 py-12">
      <div className="max-w-md mx-auto w-full">
        <Card className="border-border/50 bg-card/50">
          <CardHeader className="space-y-1">
            <CardTitle className="text-2xl text-foreground">Registrieren</CardTitle>
            <CardDescription className="text-muted-foreground">
              Erstelle einen Account, um Kegelabende zu verwalten.
            </CardDescription>
          </CardHeader>
          <form onSubmit={handleSubmit}>
            <CardContent className="space-y-4">
              {error && (
                <div
                  className="rounded-lg border border-destructive/50 bg-destructive/10 px-3 py-2 text-sm text-destructive"
                  role="alert"
                >
                  {error}
                </div>
              )}
              <div className="space-y-2">
                <Label htmlFor="register-email" className="text-foreground">
                  E-Mail
                </Label>
                <Input
                  id="register-email"
                  type="email"
                  autoComplete="email"
                  value={form.email}
                  onChange={(e) => setForm((prev) => ({ ...prev, email: e.target.value }))}
                  readOnly={lockEmail}
                  required
                  className="text-foreground"
                />
                {lockEmail && (
                  <p className="text-xs text-muted-foreground">
                    Die Registrierung ist nur mit der eingeladenen E-Mail-Adresse möglich.
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="register-password" className="text-foreground">
                  Passwort
                </Label>
                <PasswordInput
                  id="register-password"
                  autoComplete="new-password"
                  minLength={8}
                  maxLength={128}
                  value={form.password}
                  onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))}
                  required
                  className="text-foreground"
                />
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button
                type="submit"
                className="w-full rounded-full"
                size="lg"
                disabled={isSubmitting}
              >
                {isSubmitting ? "Registriere …" : "Account erstellen"}
              </Button>
              <p className="text-center text-sm text-muted-foreground">
                Bereits registriert?{" "}
                <Link to="/login" className="text-primary hover:underline font-medium">
                  Zum Login
                </Link>
              </p>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  );
}

export default RegisterPage;
