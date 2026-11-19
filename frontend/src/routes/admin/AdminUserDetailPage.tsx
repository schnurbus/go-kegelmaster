import { useEffect } from "react";
import { Link, useParams } from "react-router-dom";

import {
  AdminBadge,
  AdminEmpty,
  AdminFacts,
  AdminLoadError,
  AdminLoading,
  DisabledBadge,
  InactiveBadge,
  OrphanedBadge,
  OwnerBadge,
} from "@/components/admin/admin-ui";
import { AdminRevokeTokenButton, AdminUserActions } from "@/components/admin/admin-user-actions";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { usePageTitle } from "@/context/PageTitleContext";
import { formatAdminDate, formatAdminDateTime, useAdminFetch } from "@/lib/admin";
import type { AdminUserDetail } from "@/types/admin";

function isExpired(expiresAt: string | null): boolean {
  return expiresAt !== null && new Date(expiresAt) <= new Date();
}

function AdminUserDetailPage() {
  const { id = "" } = useParams();
  const { data: user, error, loading, reload } = useAdminFetch<AdminUserDetail>(
    `/api/admin/users/${encodeURIComponent(id)}`,
  );
  const { setTitle } = usePageTitle();

  useEffect(() => {
    if (user) setTitle(user.email);
  }, [user, setTitle]);

  if (error) return <AdminLoadError error={error} />;
  if (loading || !user) return <AdminLoading />;

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex flex-wrap items-center gap-2 break-all">
            {user.email}
            {user.is_instance_admin && <AdminBadge />}
            {user.orphaned && <OrphanedBadge />}
            {user.disabled_at && <DisabledBadge />}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <AdminFacts
            items={[
              { label: "Registriert am", value: formatAdminDateTime(user.created_at) },
              { label: "Letzter Login", value: formatAdminDateTime(user.last_login_at) },
              { label: "Passwort geändert", value: formatAdminDateTime(user.password_changed_at) },
              ...(user.disabled_at ? [{ label: "Gesperrt seit", value: formatAdminDateTime(user.disabled_at) }] : []),
              ...(user.sessions_valid_after
                ? [{ label: "Sessions widerrufen", value: formatAdminDateTime(user.sessions_valid_after) }]
                : []),
            ]}
          />
        </CardContent>
      </Card>

      <AdminUserActions user={user} onChanged={reload} />

      <Card>
        <CardHeader>
          <CardTitle>Eigene Klubs ({user.owned_clubs.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {user.owned_clubs.length === 0 ? (
            <AdminEmpty>Keine eigenen Klubs.</AdminEmpty>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Klub</TableHead>
                  <TableHead className="text-right">Spieler</TableHead>
                  <TableHead>Angelegt am</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {user.owned_clubs.map((club) => (
                  <TableRow key={club.id}>
                    <TableCell className="font-medium">
                      <Link to={`/admin/clubs/${club.id}`} className="underline-offset-4 hover:underline">
                        {club.name}
                      </Link>
                    </TableCell>
                    <TableCell className="text-right tabular-nums">{club.player_count}</TableCell>
                    <TableCell>{formatAdminDate(club.created_at)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Spieler-Verknüpfungen ({user.players.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {user.players.length === 0 ? (
            <AdminEmpty>Mit keinem Spieler verknüpft.</AdminEmpty>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Klub</TableHead>
                  <TableHead>Spieler</TableHead>
                  <TableHead>Rolle</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {user.players.map((p) => (
                  <TableRow key={p.player_id}>
                    <TableCell className="font-medium">
                      <Link to={`/admin/clubs/${p.club_id}`} className="underline-offset-4 hover:underline">
                        {p.club_name}
                      </Link>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap items-center gap-1">
                        {p.player_name}
                        {p.is_owner && <OwnerBadge />}
                        {p.inactive && <InactiveBadge />}
                      </div>
                    </TableCell>
                    <TableCell>{p.role_name ?? "–"}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>API-Tokens ({user.api_tokens.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {user.api_tokens.length === 0 ? (
            <AdminEmpty>Keine API-Tokens.</AdminEmpty>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Präfix</TableHead>
                  <TableHead>Erstellt</TableHead>
                  <TableHead>Läuft ab</TableHead>
                  <TableHead>Zuletzt benutzt</TableHead>
                  <TableHead className="text-right">
                    <span className="sr-only">Aktionen</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {user.api_tokens.map((t) => (
                  <TableRow key={t.id}>
                    <TableCell className="font-medium">{t.name}</TableCell>
                    <TableCell>
                      <code className="text-xs text-muted-foreground">{t.token_prefix}…</code>
                    </TableCell>
                    <TableCell>{formatAdminDate(t.created_at)}</TableCell>
                    <TableCell>
                      {isExpired(t.expires_at) ? (
                        <Badge variant="destructive">Abgelaufen</Badge>
                      ) : t.expires_at ? (
                        formatAdminDate(t.expires_at)
                      ) : (
                        "Nie"
                      )}
                    </TableCell>
                    <TableCell>{formatAdminDateTime(t.last_used_at)}</TableCell>
                    <TableCell className="text-right">
                      <AdminRevokeTokenButton userId={user.id} token={t} onChanged={reload} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Passkeys ({user.passkeys.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {user.passkeys.length === 0 ? (
            <AdminEmpty>Keine Passkeys.</AdminEmpty>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Erstellt</TableHead>
                  <TableHead>Zuletzt benutzt</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {user.passkeys.map((pk) => (
                  <TableRow key={pk.id}>
                    <TableCell className="font-medium">{pk.name}</TableCell>
                    <TableCell>{formatAdminDate(pk.created_at)}</TableCell>
                    <TableCell>{formatAdminDateTime(pk.last_used_at)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default AdminUserDetailPage;
