import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { AlertTriangleIcon, CheckCircle2Icon } from "lucide-react";

import { AdminLoadError, AdminLoading } from "@/components/admin/admin-ui";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { formatAdminDate, useAdminFetch } from "@/lib/admin";
import type { AdminStats } from "@/types/admin";

function StatTile({ label, value, hint }: { label: string; value: number; hint?: ReactNode }) {
  return (
    <Card className="gap-2 py-4">
      <CardHeader className="px-4">
        <CardDescription>{label}</CardDescription>
        <CardTitle className="text-2xl font-semibold tabular-nums">{value}</CardTitle>
      </CardHeader>
      {hint && <CardContent className="px-4 text-xs text-muted-foreground">{hint}</CardContent>}
    </Card>
  );
}

type Anomaly = { key: string; text: string; to?: string };

function anomaliesOf(stats: AdminStats): Anomaly[] {
  const a = stats.anomalies;
  const list: Anomaly[] = [];
  if (a.orphaned_users > 0) {
    list.push({
      key: "orphaned",
      text: `${a.orphaned_users} verwaiste Konten (ohne Klub und ohne Spieler, älter als 30 Tage)`,
      to: "/admin/users?filter=orphaned",
    });
  }
  if (a.inactive_clubs > 0) {
    list.push({
      key: "inactive",
      text: `${a.inactive_clubs} Klubs ohne Aktivität seit mehr als 180 Tagen`,
      to: "/admin/clubs?filter=inactive",
    });
  }
  if (a.api_tokens_without_expiry > 0) {
    list.push({ key: "noexpiry", text: `${a.api_tokens_without_expiry} aktive API-Tokens ohne Ablaufdatum` });
  }
  if (a.api_tokens_unused_90d > 0) {
    list.push({ key: "unused", text: `${a.api_tokens_unused_90d} aktive API-Tokens seit 90 Tagen ungenutzt` });
  }
  for (const spike of a.registration_spikes) {
    list.push({
      key: `spike-${spike.date}`,
      text: `Registrierungsspitze am ${formatAdminDate(spike.date)}: ${spike.registrations} Registrierungen`,
    });
  }
  return list;
}

function AdminDashboardPage() {
  const { data: stats, error, loading } = useAdminFetch<AdminStats>("/api/admin/stats");

  if (error) return <AdminLoadError error={error} />;
  if (loading || !stats) return <AdminLoading />;

  const anomalies = anomaliesOf(stats);

  return (
    <div className="flex flex-col gap-6">
      <section aria-label="Kennzahlen" className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
        <StatTile
          label="Benutzer"
          value={stats.users.total}
          hint={`${stats.users.instance_admins} Admin${stats.users.instance_admins === 1 ? "" : "s"} · ${stats.users.disabled} gesperrt`}
        />
        <StatTile label="Neu (7 Tage)" value={stats.users.new_7d} hint={`${stats.users.new_30d} in 30 Tagen`} />
        <StatTile label="Aktiv (30 Tage)" value={stats.users.active_30d} hint="Login in den letzten 30 Tagen" />
        <StatTile
          label="Klubs"
          value={stats.clubs.total}
          hint={`${stats.clubs.active_90d} aktiv (90 Tage)`}
        />
        <StatTile
          label="Spieler"
          value={stats.players.total}
          hint={`${stats.players.linked} mit Benutzer verknüpft`}
        />
        <StatTile label="Offene Einladungen" value={stats.open_invitations} />
        <StatTile label="API-Tokens aktiv" value={stats.api_tokens.active} />
        <StatTile label="Passkeys" value={stats.passkeys} />
      </section>

      <Card>
        <CardHeader>
          <CardTitle>Auffälligkeiten</CardTitle>
          <CardDescription>Hinweise auf verwaiste Konten, inaktive Klubs und ungewöhnliche Nutzung.</CardDescription>
        </CardHeader>
        <CardContent>
          {anomalies.length === 0 ? (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
              <CheckCircle2Icon className="size-4" />
              Keine Auffälligkeiten.
            </p>
          ) : (
            <ul className="divide-y text-sm">
              {anomalies.map((item) => (
                <li key={item.key} className="flex items-start gap-2 py-2 first:pt-0 last:pb-0">
                  <AlertTriangleIcon className="mt-0.5 size-4 shrink-0 text-amber-500" />
                  {item.to ? (
                    <Link to={item.to} className="underline-offset-4 hover:underline">
                      {item.text}
                    </Link>
                  ) : (
                    <span>{item.text}</span>
                  )}
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default AdminDashboardPage;
