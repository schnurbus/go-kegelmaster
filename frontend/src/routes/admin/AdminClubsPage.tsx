import { Link } from "react-router-dom";

import {
  AdminEmpty,
  AdminLoadError,
  AdminLoading,
  AdminPagination,
  AdminSearchInput,
  SortableHead,
} from "@/components/admin/admin-ui";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatAdminDate, useAdminFetch, useAdminListParams } from "@/lib/admin";
import type { AdminClubListItem, AdminPage } from "@/types/admin";

function AdminClubsPage() {
  const list = useAdminListParams("/api/admin/clubs", "created_at", "desc");
  const { data, error, loading } = useAdminFetch<AdminPage<AdminClubListItem>>(list.apiUrl);

  return (
    <Card>
      <CardHeader className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <AdminSearchInput
          value={list.search}
          onSearch={list.setSearch}
          label="Klubs suchen"
          placeholder="Name suchen …"
        />
        <Select value={list.filter || "all"} onValueChange={(v) => list.setFilter(v === "all" ? "" : v)}>
          <SelectTrigger className="w-full sm:w-52" aria-label="Filter">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Alle Klubs</SelectItem>
            <SelectItem value="inactive">Inaktiv (180 Tage)</SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent>
        {error ? (
          <AdminLoadError error={error} />
        ) : loading || !data ? (
          <AdminLoading />
        ) : data.data.length === 0 ? (
          <AdminEmpty>Keine Klubs gefunden.</AdminEmpty>
        ) : (
          <>
            <ul className="divide-y md:hidden">
              {data.data.map((club) => (
                <li key={club.id} className="space-y-1 py-3 text-sm first:pt-0">
                  <Link to={`/admin/clubs/${club.id}`} className="break-words font-medium underline-offset-4 hover:underline">
                    {club.name}
                  </Link>
                  <dl className="grid grid-cols-[auto_1fr] gap-x-3 text-xs text-muted-foreground">
                    <dt>Owner</dt>
                    <dd className="break-all">
                      <Link to={`/admin/users/${club.owner.id}`} className="underline-offset-4 hover:underline">
                        {club.owner.email}
                      </Link>
                    </dd>
                    <dt>Spieler</dt>
                    <dd>
                      {club.player_count} gesamt / {club.linked_player_count} verknüpft
                    </dd>
                    <dt>Angelegt</dt>
                    <dd>{formatAdminDate(club.created_at)}</dd>
                    <dt>Letzte Aktivität</dt>
                    <dd>{formatAdminDate(club.last_activity_at)}</dd>
                    <dt>Spieltage</dt>
                    <dd>{club.game_day_count}</dd>
                  </dl>
                </li>
              ))}
            </ul>
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <SortableHead label="Name" field="name" sort={list.sort} order={list.order} onSort={list.toggleSort} />
                    <TableHead>Owner</TableHead>
                    <TableHead className="text-right">Spieler (gesamt / verknüpft)</TableHead>
                    <SortableHead label="Angelegt am" field="created_at" sort={list.sort} order={list.order} onSort={list.toggleSort} />
                    <SortableHead
                      label="Letzte Aktivität"
                      field="last_activity_at"
                      sort={list.sort}
                      order={list.order}
                      onSort={list.toggleSort}
                    />
                    <TableHead className="text-right">Spieltage</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.data.map((club) => (
                    <TableRow key={club.id}>
                      <TableCell className="font-medium">
                        <Link to={`/admin/clubs/${club.id}`} className="underline-offset-4 hover:underline">
                          {club.name}
                        </Link>
                      </TableCell>
                      <TableCell>
                        <Link to={`/admin/users/${club.owner.id}`} className="underline-offset-4 hover:underline">
                          {club.owner.email}
                        </Link>
                      </TableCell>
                      <TableCell className="text-right tabular-nums">
                        {club.player_count} / {club.linked_player_count}
                      </TableCell>
                      <TableCell>{formatAdminDate(club.created_at)}</TableCell>
                      <TableCell>{formatAdminDate(club.last_activity_at)}</TableCell>
                      <TableCell className="text-right tabular-nums">{club.game_day_count}</TableCell>
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

export default AdminClubsPage;
