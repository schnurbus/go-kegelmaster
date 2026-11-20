import * as React from "react";
import { PlugZapIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { errorMessage, oauthCsrfToken } from "@/lib/oauth";
import type { OAuthGrant } from "@/types/oauth";

function formatDate(value: string | null): string {
  if (!value) return "–";
  return new Date(value).toLocaleString("de-DE", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

/** Lists the applications connected via OAuth (MCP clients) and lets the user revoke them. */
export function ConnectedAppsCard() {
  const [grants, setGrants] = React.useState<OAuthGrant[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [toRevoke, setToRevoke] = React.useState<OAuthGrant | null>(null);
  const [isRevoking, setIsRevoking] = React.useState(false);

  const fetchGrants = React.useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await fetch("/api/oauth/grants", { credentials: "include" });
      if (!response.ok) throw new Error();
      setGrants(await response.json());
    } catch {
      toast.error("Fehler beim Laden der verbundenen Anwendungen");
      setGrants([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  React.useEffect(() => {
    fetchGrants();
  }, [fetchGrants]);

  const handleRevoke = async () => {
    if (!toRevoke) return;
    setIsRevoking(true);
    try {
      const response = await fetch(`/api/oauth/grants/${toRevoke.id}`, {
        method: "DELETE",
        headers: { "X-CSRF-Token": await oauthCsrfToken() },
        credentials: "include",
      });
      if (!response.ok) {
        throw new Error(await errorMessage(response, "Fehler beim Widerrufen"));
      }
      toast.success("Zugriff widerrufen");
      setToRevoke(null);
      fetchGrants();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Fehler beim Widerrufen");
    } finally {
      setIsRevoking(false);
    }
  };

  const revokeButton = (grant: OAuthGrant) => (
    <Button
      variant="ghost"
      size="icon"
      aria-label={`Zugriff von ${grant.client_name} widerrufen`}
      onClick={() => setToRevoke(grant)}
    >
      <Trash2Icon />
    </Button>
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Verbundene Anwendungen</CardTitle>
        <CardDescription className="max-w-2xl">
          Anwendungen, denen Sie per Anmeldung (OAuth) Zugriff auf den MCP-Server
          gegeben haben. Sie handeln mit Ihren Rechten. Nach dem Widerruf verlieren
          sie sofort den Zugriff.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Lädt …</p>
        ) : grants.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-8 text-center text-sm text-muted-foreground">
            <PlugZapIcon className="size-8" />
            Keine Anwendungen verbunden.
          </div>
        ) : (
          <>
            <ul className="divide-y md:hidden">
              {grants.map((grant) => (
                <li key={grant.id} className="flex items-start gap-2 py-3 first:pt-0 last:pb-0">
                  <div className="min-w-0 flex-1 space-y-1 text-sm">
                    <p className="break-words font-medium">{grant.client_name}</p>
                    <code className="break-all text-xs text-muted-foreground">{grant.redirect_host}</code>
                    <dl className="grid grid-cols-[auto_1fr] gap-x-3 text-xs text-muted-foreground">
                      <dt>Verbunden</dt>
                      <dd>{formatDate(grant.created_at)}</dd>
                      <dt>Zuletzt genutzt</dt>
                      <dd>{formatDate(grant.last_used_at)}</dd>
                    </dl>
                  </div>
                  {revokeButton(grant)}
                </li>
              ))}
            </ul>
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Anwendung</TableHead>
                    <TableHead>Weiterleitung an</TableHead>
                    <TableHead>Verbunden</TableHead>
                    <TableHead>Zuletzt genutzt</TableHead>
                    <TableHead className="w-12" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {grants.map((grant) => (
                    <TableRow key={grant.id}>
                      <TableCell className="max-w-64 truncate font-medium" title={grant.client_name}>
                        {grant.client_name}
                      </TableCell>
                      <TableCell>
                        <code className="text-xs">{grant.redirect_host}</code>
                      </TableCell>
                      <TableCell>{formatDate(grant.created_at)}</TableCell>
                      <TableCell>{formatDate(grant.last_used_at)}</TableCell>
                      <TableCell>{revokeButton(grant)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </>
        )}
      </CardContent>

      <AlertDialog open={toRevoke !== null} onOpenChange={(open) => !open && setToRevoke(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Zugriff widerrufen?</AlertDialogTitle>
            <AlertDialogDescription>
              „{toRevoke?.client_name}“ verliert sofort den Zugriff auf Kegelmaster.
              Um die Anwendung erneut zu verbinden, müssen Sie den Zugriff dort neu
              erlauben.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRevoking}>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                handleRevoke();
              }}
              disabled={isRevoking}
            >
              {isRevoking ? "Widerruft …" : "Widerrufen"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  );
}
