import * as React from "react";
import { useSearchParams } from "react-router-dom";
import { PlugZapIcon, ShieldAlertIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/context/AuthContext";
import { AUTHORIZE_PARAMS, errorMessage, oauthCsrfToken } from "@/lib/oauth";
import type { OAuthConsentInfo } from "@/types/oauth";

/**
 * Consent page of the OAuth authorization server (KEG-12). /oauth/authorize
 * validates the request and forwards here; RequireAuth sends users who are
 * not logged in to the login first and back afterwards.
 */
function OAuthConsentPage() {
  const [searchParams] = useSearchParams();
  const { user } = useAuth();
  const [info, setInfo] = React.useState<OAuthConsentInfo | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [submitting, setSubmitting] = React.useState<"approve" | "deny" | null>(null);

  const params = React.useMemo(() => {
    const out: Record<string, string> = {};
    for (const key of AUTHORIZE_PARAMS) {
      const value = searchParams.get(key);
      if (value !== null) out[key] = value;
    }
    return out;
  }, [searchParams]);

  React.useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        const resp = await fetch(`/api/oauth/authorize?${new URLSearchParams(params).toString()}`, {
          credentials: "include",
        });
        if (!resp.ok) {
          throw new Error(await errorMessage(resp, "Die Anfrage ist ungültig."));
        }
        const data: OAuthConsentInfo = await resp.json();
        if (!cancelled) setInfo(data);
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : "Die Anfrage ist ungültig.");
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, [params]);

  const decide = async (approve: boolean) => {
    setSubmitting(approve ? "approve" : "deny");
    try {
      const resp = await fetch("/api/oauth/authorize", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": await oauthCsrfToken(),
        },
        body: JSON.stringify({ ...params, approve }),
      });
      if (!resp.ok) {
        throw new Error(await errorMessage(resp, "Die Entscheidung konnte nicht gespeichert werden."));
      }
      const data: { redirect_to: string } = await resp.json();
      // Only the registered redirect_uri (checked by the server) is used here.
      window.location.assign(data.redirect_to);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Die Entscheidung konnte nicht gespeichert werden.");
      setSubmitting(null);
    }
  };

  return (
    <div className="flex justify-center px-4 py-10">
      <Card className="w-full max-w-md">
        {error ? (
          <>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <ShieldAlertIcon className="size-5 text-destructive" />
                Verbindung nicht möglich
              </CardTitle>
              <CardDescription>{error}</CardDescription>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              Bitte starten Sie die Verbindung in Ihrer Anwendung erneut.
            </CardContent>
          </>
        ) : !info ? (
          <CardHeader className="space-y-3">
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-24 w-full" />
          </CardHeader>
        ) : (
          <>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <PlugZapIcon className="size-5" />
                Zugriff erlauben?
              </CardTitle>
              <CardDescription>
                <span className="font-medium text-foreground break-words">„{info.client_name}“</span> möchte
                über den MCP-Server auf Kegelmaster zugreifen.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
                <dt className="text-muted-foreground">Angemeldet als</dt>
                <dd className="break-all">{user?.email}</dd>
                <dt className="text-muted-foreground">Weiterleitung an</dt>
                <dd className="break-all font-mono text-xs leading-5">{info.redirect_host}</dd>
              </dl>
              <p className="text-muted-foreground">
                Die Anwendung kann mit <strong>Ihren Rechten</strong> in allen Ihren Klubs
                Daten lesen und – soweit Ihre Rolle es erlaubt – Spieltage erfassen und
                Buchungen anlegen. Erlauben Sie den Zugriff nur, wenn Sie die Verbindung gerade
                selbst gestartet hast und der Weiterleitungs-Host zur Anwendung passt.
              </p>
              <p className="text-muted-foreground">
                Sie können die Verbindung jederzeit unter „API-Tokens“ → „Verbundene
                Anwendungen“ widerrufen.
              </p>
            </CardContent>
            <CardFooter className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <Button
                variant="outline"
                className="w-full sm:w-auto"
                disabled={submitting !== null}
                onClick={() => decide(false)}
              >
                {submitting === "deny" ? "Lehnt ab …" : "Ablehnen"}
              </Button>
              <Button className="w-full sm:w-auto" disabled={submitting !== null} onClick={() => decide(true)}>
                {submitting === "approve" ? "Erlaubt …" : "Zulassen"}
              </Button>
            </CardFooter>
          </>
        )}
      </Card>
    </div>
  );
}

export default OAuthConsentPage;
