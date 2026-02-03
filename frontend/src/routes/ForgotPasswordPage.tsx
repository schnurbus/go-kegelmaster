import type { FormEvent } from "react";
import { useState } from "react";
import { Link } from "react-router-dom";
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

function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    try {
      const resp = await fetch("/api/auth/forgot-password", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim() }),
      });
      if (!resp.ok) {
        const data = await resp.json().catch(() => ({}));
        const message = data?.message ?? resp.statusText;
        throw new Error(message);
      }
      setSubmitted(true);
      toast.success("Falls ein Konto existiert, wurde eine E-Mail gesendet.");
    } catch (err) {
      const message = (err as Error).message;
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (submitted) {
    return (
      <div className="min-h-[80vh] flex flex-col justify-center px-4 py-12">
        <div className="max-w-md mx-auto w-full">
          <Card className="border-border/50 bg-card/50">
            <CardHeader className="space-y-1">
              <CardTitle className="text-2xl text-foreground">E-Mail gesendet</CardTitle>
              <CardDescription className="text-muted-foreground">
                Falls ein Konto mit dieser E-Mail-Adresse existiert, wurde ein Link zum
                Zurücksetzen des Passworts gesendet. Bitte prüfen Sie Ihr Postfach (auch den
                Spam-Ordner).
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <Button asChild variant="outline" className="w-full rounded-full" size="lg">
                <Link to="/login">Zurück zum Login</Link>
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
            <CardTitle className="text-2xl text-foreground">Passwort vergessen</CardTitle>
            <CardDescription className="text-muted-foreground">
              Geben Sie Ihre E-Mail-Adresse ein. Wir senden Ihnen einen Link zum
              Zurücksetzen des Passworts.
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
                <Label htmlFor="forgot-email" className="text-foreground">
                  E-Mail
                </Label>
                <Input
                  id="forgot-email"
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
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
                {isSubmitting ? "Sende …" : "Link senden"}
              </Button>
              <p className="text-center text-sm text-muted-foreground">
                <Link to="/login" className="text-primary hover:underline font-medium">
                  Zurück zum Login
                </Link>
              </p>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  );
}

export default ForgotPasswordPage;
