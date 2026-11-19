import { Link } from "react-router-dom";

import {
  AdminBadge,
  AdminEmpty,
  AdminLoadError,
  AdminLoading,
  AdminPagination,
  AdminSearchInput,
  OrphanedBadge,
  SortableHead,
} from "@/components/admin/admin-ui";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatAdminDate, formatAdminDateTime, useAdminFetch, useAdminListParams } from "@/lib/admin";
import type { AdminPage, AdminUserListItem } from "@/types/admin";

function UserStatus({ user }: { user: AdminUserListItem }) {
  return (
    <div className="flex flex-wrap gap-1">
      {user.is_instance_admin && <AdminBadge />}
      {user.orphaned && <OrphanedBadge />}
      {!user.is_instance_admin && !user.orphaned && <Badge variant="outline">aktiv</Badge>}
    </div>
  );
}

function AdminUsersPage() {
  const list = useAdminListParams("/api/admin/users", "created_at", "desc");
  const { data, error, loading } = useAdminFetch<AdminPage<AdminUserListItem>>(list.apiUrl);

  return (
    <Card>
      <CardHeader className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <AdminSearchInput
          value={list.search}
          onSearch={list.setSearch}
          label="Benutzer suchen"
          placeholder="E-Mail suchen …"
        />
        <Select value={list.filter || "all"} onValueChange={(v) => list.setFilter(v === "all" ? "" : v)}>
          <SelectTrigger className="w-full sm:w-44" aria-label="Filter">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Alle Benutzer</SelectItem>
            <SelectItem value="orphaned">Verwaist</SelectItem>
            <SelectItem value="admin">Admins</SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent>
        {error ? (
          <AdminLoadError error={error} />
        ) : loading || !data ? (
          <AdminLoading />
        ) : data.data.length === 0 ? (
          <AdminEmpty>Keine Benutzer gefunden.</AdminEmpty>
        ) : (
          <>
            <ul className="divide-y md:hidden">
              {data.data.map((user) => (
                <li key={user.id} className="space-y-1 py-3 text-sm first:pt-0">
                  <Link to={`/admin/users/${user.id}`} className="break-all font-medium underline-offset-4 hover:underline">
                    {user.email}
                  </Link>
                  <UserStatus user={user} />
                  <dl className="grid grid-cols-[auto_1fr] gap-x-3 text-xs text-muted-foreground">
                    <dt>Registriert</dt>
                    <dd>{formatAdminDate(user.created_at)}</dd>
                    <dt>Letzter Login</dt>
                    <dd>{formatAdminDateTime(user.last_login_at)}</dd>
                    <dt>Owner / Spieler</dt>
                    <dd>
                      {user.owned_club_count} / {user.player_club_count} Klubs
                    </dd>
                    <dt>Passkeys / Tokens</dt>
                    <dd>
                      {user.passkey_count} / {user.api_token_count}
                    </dd>
                  </dl>
                </li>
              ))}
            </ul>
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <SortableHead label="E-Mail" field="email" sort={list.sort} order={list.order} onSort={list.toggleSort} />
                    <SortableHead label="Registriert am" field="created_at" sort={list.sort} order={list.order} onSort={list.toggleSort} />
                    <SortableHead label="Letzter Login" field="last_login_at" sort={list.sort} order={list.order} onSort={list.toggleSort} />
                    <TableHead className="text-right">Owner von</TableHead>
                    <TableHead className="text-right">Spieler in</TableHead>
                    <TableHead className="text-right">Passkeys / Tokens</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.data.map((user) => (
                    <TableRow key={user.id}>
                      <TableCell className="font-medium">
                        <Link to={`/admin/users/${user.id}`} className="underline-offset-4 hover:underline">
                          {user.email}
                        </Link>
                      </TableCell>
                      <TableCell>{formatAdminDate(user.created_at)}</TableCell>
                      <TableCell>{formatAdminDateTime(user.last_login_at)}</TableCell>
                      <TableCell className="text-right tabular-nums">{user.owned_club_count}</TableCell>
                      <TableCell className="text-right tabular-nums">{user.player_club_count}</TableCell>
                      <TableCell className="text-right tabular-nums">
                        {user.passkey_count} / {user.api_token_count}
                      </TableCell>
                      <TableCell>
                        <UserStatus user={user} />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
            <AdminPagination
              page={data.page}
              totalPages={data.total_pages}
              total={data.total}
              onPageChange={list.setPage}
            />
          </>
        )}
      </CardContent>
    </Card>
  );
}

export default AdminUsersPage;
