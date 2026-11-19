import { useEffect } from "react";
import { Link, useParams } from "react-router-dom";

import {
  AdminEmpty,
  AdminFacts,
  AdminLoadError,
  AdminLoading,
  InactiveBadge,
  LinkedBadge,
  OwnerBadge,
} from "@/components/admin/admin-ui";
import { AdminClubActions } from "@/components/admin/admin-club-actions";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { usePageTitle } from "@/context/PageTitleContext";
import { formatAdminDate, formatAdminDateTime, useAdminFetch } from "@/lib/admin";
import type { AdminClubDetail } from "@/types/admin";

function AdminClubDetailPage() {
  const { id = "" } = useParams();
  const { data: club, error, loading, reload } = useAdminFetch<AdminClubDetail>(
    `/api/admin/clubs/${encodeURIComponent(id)}`,
  );
  const { setTitle } = usePageTitle();

  useEffect(() => {
    if (club) setTitle(club.name);
  }, [club, setTitle]);

  if (error) return <AdminLoadError error={error} />;
  if (loading || !club) return <AdminLoading />;

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle className="break-words">{club.name}</CardTitle>
        </CardHeader>
        <CardContent>
          <AdminFacts
            items={[
              {
                label: "Owner",
                value: (
                  <Link to={`/admin/users/${club.owner.id}`} className="break-all underline-offset-4 hover:underline">
                    {club.owner.email}
                  </Link>
                ),
              },
              { label: "Angelegt am", value: formatAdminDate(club.created_at) },
              { label: "Letzte Aktivität", value: formatAdminDate(club.last_activity_at) },
              { label: "Spieler (gesamt / verknüpft)", value: `${club.player_count} / ${club.linked_player_count}` },
              { label: "Spieltage", value: club.game_day_count },
              { label: "Transaktionen", value: club.transaction_count },
            ]}
          />
        </CardContent>
      </Card>

      <AdminClubActions club={club} onChanged={reload} />

      <Card>
        <CardHeader>
          <CardTitle>Spieler ({club.players.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {club.players.length === 0 ? (
            <AdminEmpty>Keine Spieler.</AdminEmpty>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Spieler</TableHead>
                  <TableHead>Benutzer</TableHead>
                  <TableHead>Rolle</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {club.players.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell className="font-medium">{p.name}</TableCell>
                    <TableCell>
                      {p.user ? (
                        <Link to={`/admin/users/${p.user.id}`} className="underline-offset-4 hover:underline">
                          {p.user.email}
                        </Link>
                      ) : (
                        <span className="text-muted-foreground">nicht verknüpft</span>
                      )}
                    </TableCell>
                    <TableCell>{p.role_name ?? "–"}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap items-center gap-1">
                        <LinkedBadge linked={p.user !== null} />
                        {p.user?.id === club.owner.id && <OwnerBadge />}
                        {p.inactive && <InactiveBadge />}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Offene Einladungen ({club.open_invitations.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {club.open_invitations.length === 0 ? (
              <AdminEmpty>Keine offenen Einladungen.</AdminEmpty>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Spieler</TableHead>
                    <TableHead>E-Mail</TableHead>
                    <TableHead>Läuft ab</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {club.open_invitations.map((inv) => (
                    <TableRow key={inv.id}>
                      <TableCell className="font-medium">{inv.player_name}</TableCell>
                      <TableCell>{inv.email}</TableCell>
                      <TableCell>{formatAdminDateTime(inv.expires_at)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Rollen ({club.roles.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {club.roles.length === 0 ? (
              <AdminEmpty>Keine Rollen.</AdminEmpty>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Rolle</TableHead>
                    <TableHead className="text-right">Spieler</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {club.roles.map((role) => (
                    <TableRow key={role.id}>
                      <TableCell className="font-medium">{role.name}</TableCell>
                      <TableCell className="text-right tabular-nums">{role.player_count}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default AdminClubDetailPage;
