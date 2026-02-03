import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useNavigate } from "react-router-dom";
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

function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const navigate = useNavigate();
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    if (password !== confirmPassword) {
      setError("Die Passwörter stimmen nicht überein.");
      return;
    }
    if (password.length < 8) {
      setError("Passwort muss mindestens 8 Zeichen lang sein.");
      return;
    }
    if (!token) {
      setError("Kein gültiger Link. Bitte fordern Sie einen neuen Link an.");
      return;
    }
    setIsSubmitting(true);
    try {
      const resp = await fetch("/api/auth/reset-password", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, new_password: password }),
      });
      if (!resp.ok) {
        const data = await resp.json().catch(() => ({}));
        const message = data?.message ?? resp.statusText;
        throw new Error(message);
      }
      setSuccess(true);
      toast.success("Passwort wurde geändert. Sie können sich jetzt anmelden.");
      setTimeout(() => navigate("/login", { replace: true }), 2000);
    } catch (err) {
      const message = (err as Error).message;
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (success) {
    return (
      <div className="min-h-[80vh] flex flex-col justify-center px-4 py-12">
        <div className="max-w-md mx-auto w-full">
          <Card className="border-border/50 bg-card/50">
            <CardHeader className="space-y-1">
              <CardTitle className="text-2xl text-foreground">Passwort geändert</CardTitle>
              <CardDescription className="text-muted-foreground">
                Ihr Passwort wurde erfolgreich geändert. Sie werden zum Login weitergeleitet.
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <Button asChild className="w-full rounded-full" size="lg">
                <Link to="/login">Zum Login</Link>
              </Button>
            </CardFooter>
          </Card>
        </div>
      </div>
    );
  }

  if (!token) {
    return (
      <div className="min-h-[80vh] flex flex-col justify-center px-4 py-12">
        <div className="max-w-md mx-auto w-full">
          <Card className="border-border/50 bg-card/50">
            <CardHeader className="space-y-1">
              <CardTitle className="text-2xl text-foreground">Ungültiger Link</CardTitle>
              <CardDescription className="text-muted-foreground">
                Dieser Link ist ungültig oder abgelaufen. Bitte fordern Sie einen neuen Link
                zum Zurücksetzen des Passworts an.
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <Button asChild className="w-full rounded-full" size="lg">
                <Link to="/forgot-password">Neuen Link anfordern</Link>
              </Button>
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
            <CardTitle className="text-2xl text-foreground">Neues Passwort setzen</CardTitle>
            <CardDescription className="text-muted-foreground">
              Geben Sie Ihr neues Passwort ein (mindestens 8 Zeichen).
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
                <Label htmlFor="reset-password" className="text-foreground">
                  Neues Passwort
                </Label>
                <Input
                  id="reset-password"
                  type="password"
                  autoComplete="new-password"
                  minLength={8}
                  maxLength={128}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  className="text-foreground"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="reset-confirm" className="text-foreground">
                  Passwort wiederholen
                </Label>
                <Input
                  id="reset-confirm"
                  type="password"
                  autoComplete="new-password"
                  minLength={8}
                  maxLength={128}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
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
                {isSubmitting ? "Speichere …" : "Passwort speichern"}
              </Button>
              <p className="text-center text-sm text-muted-foreground">
                <Link to="/forgot-password" className="text-primary hover:underline font-medium">
                  Neuen Link anfordern
                </Link>
              </p>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  );
}

export default ResetPasswordPage;
