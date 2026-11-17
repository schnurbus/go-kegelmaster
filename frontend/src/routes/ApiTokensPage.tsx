import * as React from "react";
import { CopyIcon, KeyRoundIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";

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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { ApiToken, CreatedApiToken } from "@/types/api-token";

const EXPIRY_OPTIONS = [
  { value: "30", label: "30 Tage" },
  { value: "90", label: "90 Tage" },
  { value: "365", label: "1 Jahr" },
  { value: "never", label: "Unbegrenzt" },
];

function csrfToken(): string {
  return (
    document.cookie
      .split("; ")
      .find((row) => row.startsWith("csrf_token="))
      ?.split("=")[1] || ""
  );
}

function formatDate(value: string | null): string {
  if (!value) return "–";
  return new Date(value).toLocaleString("de-DE", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

function isExpired(token: ApiToken): boolean {
  return token.expires_at !== null && new Date(token.expires_at) <= new Date();
}

function expiryLabel(token: ApiToken): React.ReactNode {
  if (isExpired(token)) return <Badge variant="destructive">Abgelaufen</Badge>;
  return token.expires_at ? formatDate(token.expires_at) : "Nie";
}

function mcpUrl(): string {
  return `${window.location.origin}/mcp`;
}

function mcpCommand(token: string): string {
  return `claude mcp add --transport http kegelmaster ${mcpUrl()} --header "Authorization: Bearer ${token}"`;
}

async function copyText(text: string, success: string) {
  try {
    await navigator.clipboard.writeText(text);
    toast.success(success);
  } catch {
    toast.error("Kopieren nicht möglich – bitte manuell markieren");
  }
}

function ApiTokensPage() {
  const [tokens, setTokens] = React.useState<ApiToken[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isCreateOpen, setIsCreateOpen] = React.useState(false);
  const [name, setName] = React.useState("");
  const [expiry, setExpiry] = React.useState("90");
  const [isCreating, setIsCreating] = React.useState(false);
  const [created, setCreated] = React.useState<CreatedApiToken | null>(null);
  const [toDelete, setToDelete] = React.useState<ApiToken | null>(null);
  const [isDeleting, setIsDeleting] = React.useState(false);

  const fetchTokens = React.useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await fetch("/api/auth/tokens", { credentials: "include" });
      if (!response.ok) throw new Error();
      setTokens(await response.json());
    } catch {
      toast.error("Fehler beim Laden der API-Tokens");
      setTokens([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  React.useEffect(() => {
    fetchTokens();
  }, [fetchTokens]);

  const openCreate = () => {
    setName("");
    setExpiry("90");
    setCreated(null);
    setIsCreateOpen(true);
  };

  const handleCreate = async (event: React.FormEvent) => {
    event.preventDefault();
    setIsCreating(true);
    try {
      const response = await fetch("/api/auth/tokens", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken(),
        },
        credentials: "include",
        body: JSON.stringify({
          name: name.trim(),
          expires_in_days: expiry === "never" ? null : Number(expiry),
        }),
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Erstellen");
      }
      setCreated(await response.json());
      fetchTokens();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Fehler beim Erstellen");
    } finally {
      setIsCreating(false);
    }
  };

  const copyToken = () => {
    if (created) copyText(created.token, "Token kopiert");
  };

  const handleDelete = async () => {
    if (!toDelete) return;
    setIsDeleting(true);
    try {
      const response = await fetch(`/api/auth/tokens/${toDelete.id}`, {
        method: "DELETE",
        headers: { "X-CSRF-Token": csrfToken() },
        credentials: "include",
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Widerrufen");
      }
      toast.success("API-Token widerrufen");
      setToDelete(null);
      fetchTokens();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Fehler beim Widerrufen");
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div className="flex flex-col gap-6 px-4 py-4 md:px-6 md:py-6">
      <Card>
        <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1.5">
            <CardTitle>Persönliche API-Tokens</CardTitle>
            <CardDescription className="max-w-2xl">
              Mit einem API-Token können Skripte oder KI-Agenten die
              Kegelmaster-API in Ihrem Namen nutzen – mit genau Ihren Rechten.
              Senden Sie das Token im Header{" "}
              <code className="rounded bg-muted px-1 py-0.5 text-xs">
                Authorization: Bearer &lt;token&gt;
              </code>
              . Die API-Dokumentation finden Sie unter{" "}
              <a className="underline" href="/api/docs" target="_blank" rel="noreferrer">
                /api/docs
              </a>
              .
            </CardDescription>
          </div>
          <Button onClick={openCreate} className="w-full shrink-0 sm:w-auto">
            <PlusIcon />
            Token erstellen
          </Button>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-muted-foreground">Lädt …</p>
          ) : tokens.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-8 text-center text-sm text-muted-foreground">
              <KeyRoundIcon className="size-8" />
              Noch keine API-Tokens vorhanden.
            </div>
          ) : (
            <>
            <ul className="divide-y md:hidden">
              {tokens.map((token) => (
                <li key={token.id} className="flex items-start gap-2 py-3 first:pt-0 last:pb-0">
                  <div className="min-w-0 flex-1 space-y-1 text-sm">
                    <p className="break-words font-medium">{token.name}</p>
                    <code className="text-xs text-muted-foreground">{token.token_prefix}…</code>
                    <dl className="grid grid-cols-[auto_1fr] gap-x-3 text-xs text-muted-foreground">
                      <dt>Erstellt</dt>
                      <dd>{formatDate(token.created_at)}</dd>
                      <dt>Zuletzt genutzt</dt>
                      <dd>{formatDate(token.last_used_at)}</dd>
                      <dt>Läuft ab</dt>
                      <dd>{expiryLabel(token)}</dd>
                    </dl>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={`Token ${token.name} widerrufen`}
                    onClick={() => setToDelete(token)}
                  >
                    <Trash2Icon />
                  </Button>
                </li>
              ))}
            </ul>
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Token</TableHead>
                    <TableHead>Erstellt</TableHead>
                    <TableHead>Zuletzt genutzt</TableHead>
                    <TableHead>Läuft ab</TableHead>
                    <TableHead className="w-12" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tokens.map((token) => (
                    <TableRow key={token.id}>
                      <TableCell className="max-w-64 truncate font-medium" title={token.name}>
                        {token.name}
                      </TableCell>
                      <TableCell>
                        <code className="text-xs">{token.token_prefix}…</code>
                      </TableCell>
                      <TableCell>{formatDate(token.created_at)}</TableCell>
                      <TableCell>{formatDate(token.last_used_at)}</TableCell>
                      <TableCell>{expiryLabel(token)}</TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="icon"
                          aria-label={`Token ${token.name} widerrufen`}
                          onClick={() => setToDelete(token)}
                        >
                          <Trash2Icon />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
            </>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>KI-Agenten anbinden (MCP)</CardTitle>
          <CardDescription className="max-w-2xl">
            Kegelmaster bietet einen MCP-Server. Damit kann ein KI-Assistent wie
            Claude Salden abfragen, Spieltage als Entwurf erfassen, nach Ihrer
            Bestätigung veröffentlichen und Einzahlungen buchen.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <div className="space-y-1">
            <Label htmlFor="mcp-url">Server-Adresse</Label>
            <div className="flex gap-2">
              <Input id="mcp-url" readOnly value={mcpUrl()} className="font-mono text-xs" onFocus={(e) => e.target.select()} />
              <Button type="button" variant="outline" size="icon" aria-label="Adresse kopieren" onClick={() => copyText(mcpUrl(), "Adresse kopiert")}>
                <CopyIcon />
              </Button>
            </div>
          </div>
          <p className="text-muted-foreground">
            Im KI-Client als HTTP-Server eintragen und den Header{" "}
            <code className="rounded bg-muted px-1 py-0.5 text-xs">Authorization: Bearer &lt;token&gt;</code>{" "}
            setzen. Für Claude Code:
          </p>
          <pre className="overflow-x-auto rounded-md bg-muted p-3 text-xs">
            <code>{mcpCommand("<token>")}</code>
          </pre>
        </CardContent>
      </Card>

      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent>
          {created ? (
            <>
              <DialogHeader>
                <DialogTitle>Token erstellt</DialogTitle>
                <DialogDescription>
                  Kopieren Sie das Token jetzt. Es wird nur dieses eine Mal
                  angezeigt und kann später nicht wiederhergestellt werden.
                </DialogDescription>
              </DialogHeader>
              <div className="flex gap-2">
                <Input readOnly value={created.token} className="font-mono text-xs" onFocus={(e) => e.target.select()} />
                <Button type="button" variant="outline" size="icon" onClick={copyToken} aria-label="Token kopieren">
                  <CopyIcon />
                </Button>
              </div>
              <Button
                type="button"
                variant="link"
                className="h-auto justify-start p-0 text-xs"
                onClick={() => copyText(mcpCommand(created.token), "Befehl kopiert")}
              >
                Einrichtungsbefehl für Claude Code (MCP) kopieren
              </Button>
              <DialogFooter>
                <Button onClick={() => setIsCreateOpen(false)}>Fertig</Button>
              </DialogFooter>
            </>
          ) : (
            <form onSubmit={handleCreate} className="space-y-4">
              <DialogHeader>
                <DialogTitle>API-Token erstellen</DialogTitle>
                <DialogDescription>
                  Das Token erhält alle Rechte Ihres Kontos. Geben Sie es nur an
                  Programme weiter, denen Sie vertrauen.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-2">
                <Label htmlFor="api-token-name">Name</Label>
                <Input
                  id="api-token-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="z. B. Kassen-Agent"
                  maxLength={100}
                  required
                  autoFocus
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="api-token-expiry">Gültigkeit</Label>
                <Select value={expiry} onValueChange={setExpiry}>
                  <SelectTrigger id="api-token-expiry" className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {EXPIRY_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <DialogFooter>
                <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)}>
                  Abbrechen
                </Button>
                <Button type="submit" disabled={isCreating || name.trim() === ""}>
                  {isCreating ? "Erstellt …" : "Erstellen"}
                </Button>
              </DialogFooter>
            </form>
          )}
        </DialogContent>
      </Dialog>

      <AlertDialog open={toDelete !== null} onOpenChange={(open) => !open && setToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>API-Token widerrufen?</AlertDialogTitle>
            <AlertDialogDescription>
              Programme, die das Token „{toDelete?.name}“ verwenden, verlieren
              sofort den Zugriff. Diese Aktion kann nicht rückgängig gemacht
              werden.
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
              {isDeleting ? "Widerruft …" : "Widerrufen"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

export default ApiTokensPage;
