import type { FormEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { KeyRoundIcon } from "lucide-react";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router-dom";
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
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  PasskeyCancelledError,
  conditionalMediationAvailable,
  passkeysSupported,
} from "@/lib/passkey";

const initialState = { email: "", password: "" };

function LoginPage() {
  const [form, setForm] = useState(initialState);
  const [rememberMe, setRememberMe] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isPasskeyPending, setIsPasskeyPending] = useState(false);
  // Bumped to restart the autofill request after a failed sign-in attempt.
  const [autofillRun, setAutofillRun] = useState(0);
  const autofillAbort = useRef<AbortController | null>(null);
  const rememberMeRef = useRef(rememberMe);
  rememberMeRef.current = rememberMe;
  const { login, loginWithPasskey } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const inviteToken = searchParams.get("invite_token");
  const redirectTo = (location.state as { from?: string })?.from ?? "/app";

  const acceptInvitation = async (token: string) => {
    try {
      toast.loading("Einladung wird akzeptiert...", { id: "accept-invite-after-login" });

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

      const data = await response.json();
      toast.success("Einladung erfolgreich akzeptiert!", {
        id: "accept-invite-after-login",
        duration: 4000,
      });
      return data.player_id;
    } catch (error) {
      console.error("Error accepting invitation:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Akzeptieren der Einladung",
        { id: "accept-invite-after-login", duration: 6000 },
      );
      throw error;
    }
  };

  const continueAfterLogin = async () => {
    toast.success("Anmeldung erfolgreich");
    if (inviteToken) {
      try {
        const playerId = await acceptInvitation(inviteToken);
        navigate(`/app/players/${playerId}`, { replace: true });
      } catch (inviteError) {
        console.error("Failed to accept invitation:", inviteError);
        toast.warning(
          "Anmeldung erfolgreich, aber Einladung konnte nicht akzeptiert werden. " +
            "Sie können die Einladung später erneut versuchen.",
          { duration: 8000 },
        );
        navigate(redirectTo, { replace: true });
      }
    } else {
      navigate(redirectTo, { replace: true });
    }
  };

  const showError = (err: unknown) => {
    const message = err instanceof Error ? err.message : "Anmeldung fehlgeschlagen";
    setError(message);
    toast.error(message);
  };

  // Offer passkeys in the autofill of the e-mail field where the browser supports it.
  useEffect(() => {
    let cancelled = false;
    const controller = new AbortController();
    autofillAbort.current = controller;
    (async () => {
      if (!(await conditionalMediationAvailable()) || cancelled) return;
      try {
        await loginWithPasskey({
          conditional: true,
          signal: controller.signal,
          remember_me: () => rememberMeRef.current,
        });
        await continueAfterLogin();
      } catch (err) {
        // No automatic retry: a persistent error (e.g. rate limit) would loop.
        // The passkey button and the next password attempt restart autofill.
        if (cancelled || err instanceof PasskeyCancelledError) return;
        showError(err);
      }
    })();
    return () => {
      cancelled = true;
      controller.abort();
    };
    // Re-run only when restarted explicitly; the handlers read fresh state via refs.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autofillRun]);

  const stopAutofill = () => {
    autofillAbort.current?.abort();
  };

  const handlePasskeyLogin = async () => {
    stopAutofill();
    setError(null);
    setIsPasskeyPending(true);
    try {
      await loginWithPasskey({ remember_me: rememberMe });
      await continueAfterLogin();
    } catch (err) {
      if (!(err instanceof PasskeyCancelledError)) {
        showError(err);
      }
      setAutofillRun((n) => n + 1);
    } finally {
      setIsPasskeyPending(false);
    }
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    stopAutofill();
    setError(null);
    setIsSubmitting(true);
    try {
      await login({ ...form, remember_me: rememberMe });
      await continueAfterLogin();
    } catch (err) {
      showError(err);
      setAutofillRun((n) => n + 1);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-[80vh] flex flex-col justify-center px-4 py-12">
      <div className="max-w-md mx-auto w-full">
        <Card className="border-border/50 bg-card/50">
          <CardHeader className="space-y-1">
            <CardTitle className="text-2xl text-foreground">Anmelden</CardTitle>
            <CardDescription className="text-muted-foreground">
              Willkommen zurück! Bitte melde dich mit deinem Account an.
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
                <Label htmlFor="login-email" className="text-foreground">
                  E-Mail
                </Label>
                <Input
                  id="login-email"
                  type="email"
                  autoComplete="username webauthn"
                  value={form.email}
                  onChange={(e) => setForm((prev) => ({ ...prev, email: e.target.value }))}
                  required
                  className="text-foreground"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="login-password" className="text-foreground">
                  Passwort
                </Label>
                <Input
                  id="login-password"
                  type="password"
                  autoComplete="current-password"
                  minLength={8}
                  maxLength={128}
                  value={form.password}
                  onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))}
                  required
                  className="text-foreground"
                />
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="login-remember"
                  checked={rememberMe}
                  onCheckedChange={(checked) => setRememberMe(checked === true)}
                />
                <Label
                  htmlFor="login-remember"
                  className="text-sm font-normal text-muted-foreground cursor-pointer"
                >
                  Angemeldet bleiben (30 Tage)
                </Label>
              </div>
              <div className="text-sm">
                <Link
                  to="/forgot-password"
                  className="text-primary hover:underline font-medium"
                >
                  Passwort vergessen?
                </Link>
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button
                type="submit"
                className="w-full rounded-full"
                size="lg"
                disabled={isSubmitting}
              >
                {isSubmitting ? "Melde an …" : "Login"}
              </Button>
              {passkeysSupported() && (
                <>
                  <div className="flex w-full items-center gap-3 text-xs text-muted-foreground">
                    <span className="h-px flex-1 bg-border" />
                    oder
                    <span className="h-px flex-1 bg-border" />
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    className="w-full rounded-full"
                    size="lg"
                    onClick={handlePasskeyLogin}
                    disabled={isSubmitting || isPasskeyPending}
                  >
                    <KeyRoundIcon />
                    {isPasskeyPending ? "Warte auf Passkey …" : "Mit Passkey anmelden"}
                  </Button>
                </>
              )}
              <p className="text-center text-sm text-muted-foreground">
                Kein Account?{" "}
                <Link to="/register" className="text-primary hover:underline font-medium">
                  Jetzt registrieren
                </Link>
              </p>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  );
}

export default LoginPage;
