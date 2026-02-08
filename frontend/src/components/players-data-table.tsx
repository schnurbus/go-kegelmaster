import * as React from "react";
import { useNavigate } from "react-router-dom";
import {
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import {
  AlertTriangleIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronsLeftIcon,
  ChevronsRightIcon,
  EditIcon,
  EyeIcon,
  MoreVerticalIcon,
  PlusIcon,
  TrashIcon,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";

import type { Player, Role } from "@/types/player";
import { formatCentsToEuro } from "@/types/player";

const PLAYERS_STORAGE_KEY_PREFIX = "kegelmaster_players_";

type PlayersStorage = {
  pageIndex?: number;
  pageSize?: number;
  sorting?: SortingState;
  roleFilter?: string;
  balanceFilter?: string;
  hideInactive?: boolean;
};

function loadPlayersStorage(clubId: string): Partial<PlayersStorage> {
  try {
    const raw = localStorage.getItem(PLAYERS_STORAGE_KEY_PREFIX + clubId);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as Partial<PlayersStorage>;
    return {
      pageIndex: typeof parsed.pageIndex === "number" && parsed.pageIndex >= 0 ? parsed.pageIndex : undefined,
      pageSize: typeof parsed.pageSize === "number" && parsed.pageSize >= 1 ? parsed.pageSize : undefined,
      sorting: Array.isArray(parsed.sorting) ? parsed.sorting : undefined,
      roleFilter: typeof parsed.roleFilter === "string" ? parsed.roleFilter : undefined,
      balanceFilter: typeof parsed.balanceFilter === "string" ? parsed.balanceFilter : undefined,
      hideInactive: typeof parsed.hideInactive === "boolean" ? parsed.hideInactive : undefined,
    };
  } catch {
    return {};
  }
}

function savePlayersStorage(clubId: string, data: PlayersStorage) {
  try {
    localStorage.setItem(PLAYERS_STORAGE_KEY_PREFIX + clubId, JSON.stringify(data));
  } catch {
    // ignore
  }
}

type PlayersDataTableProps = {
  players: Player[];
  roles: Role[];
  onView: (player: Player) => void;
  onEdit: (player: Player) => void;
  onDelete: (player: Player) => void;
  onCreate: () => void;
  isLoading?: boolean;
  clubId?: string;
  canCreate?: boolean;
  canUpdate?: boolean;
  canDelete?: boolean;
};

export function PlayersDataTable({
  players,
  roles,
  onView,
  onEdit,
  onDelete,
  onCreate,
  isLoading = false,
  clubId,
  canCreate = true,
  canUpdate = true,
  canDelete = true,
}: PlayersDataTableProps) {
  const navigate = useNavigate();
  const [sorting, setSorting] = React.useState<SortingState>([
    { id: "name", desc: false },
  ]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [pagination, setPagination] = React.useState({
    pageIndex: 0,
    pageSize: 10,
  });
  const [roleFilter, setRoleFilter] = React.useState<string>("all");
  const [balanceFilter, setBalanceFilter] = React.useState<string>("all");
  const [hideInactive, setHideInactive] = React.useState<boolean>(false);

  // Skip the first save after loading from storage so we don't overwrite with initial state
  const skipNextSaveRef = React.useRef(false);

  React.useEffect(() => {
    if (!clubId) return;
    const stored = loadPlayersStorage(clubId);
    if (stored.sorting !== undefined) setSorting(stored.sorting);
    if (stored.pageIndex !== undefined) setPagination((p) => ({ ...p, pageIndex: stored.pageIndex! }));
    if (stored.pageSize !== undefined) setPagination((p) => ({ ...p, pageSize: stored.pageSize! }));
    if (stored.roleFilter !== undefined) setRoleFilter(stored.roleFilter);
    if (stored.balanceFilter !== undefined) setBalanceFilter(stored.balanceFilter);
    if (stored.hideInactive !== undefined) {
      setHideInactive(stored.hideInactive);
      skipNextSaveRef.current = true;
    }
  }, [clubId]);

  React.useEffect(() => {
    if (!clubId) return;
    if (skipNextSaveRef.current) {
      skipNextSaveRef.current = false;
      return;
    }
    savePlayersStorage(clubId, {
      pageIndex: pagination.pageIndex,
      pageSize: pagination.pageSize,
      sorting,
      roleFilter,
      balanceFilter,
      hideInactive,
    });
  }, [clubId, pagination.pageIndex, pagination.pageSize, sorting, roleFilter, balanceFilter, hideInactive]);

  // Helper function to get role name
  const getRoleName = (roleId: string | null) => {
    if (!roleId) return "Keine Rolle zugewiesen";
    const role = roles.find((r) => r.id === roleId);
    return role?.name || "Rolle nicht gefunden";
  };

  const columns: ColumnDef<Player>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue("name")}</div>
      ),
    },
    {
      accessorKey: "role_id",
      header: "Rolle",
      cell: ({ row }) => {
        const roleId = row.original.role_id;
        const roleName = getRoleName(roleId);
        const inactive = row.original.inactive;
        return (
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant={roleId ? "outline" : "secondary"}>{roleName}</Badge>
            {inactive && (
              <Badge variant="secondary" title="Zahlt keine Grundgebühr">
                Inaktiv
              </Badge>
            )}
          </div>
        );
      },
      filterFn: (row, _id, value) => {
        if (value === "all") return true;
        if (value === "none") return row.original.role_id === null;
        return row.original.role_id === value;
      },
    },
    {
      accessorKey: "balance",
      header: () => <div className="text-right">Balance</div>,
      cell: ({ row }) => {
        const balance = row.getValue("balance") as number;
        const formatted = formatCentsToEuro(balance);
        return (
          <div
            className={`text-right font-medium flex items-center justify-end gap-1 ${
              balance < 0 ? "text-red-500" : balance > 0 ? "text-green-500" : ""
            }`}
          >
            {formatted}
            {balance > 0 && (
              <span title="Positives Guthaben - sollte mit Auto-Tip nicht vorkommen">
                <AlertTriangleIcon className="size-4 text-yellow-500" />
              </span>
            )}
          </div>
        );
      },
      filterFn: (row, _id, value) => {
        if (value === "all") return true;
        const balance = row.getValue("balance") as number;
        if (value === "positive") return balance > 0;
        if (value === "negative") return balance < 0;
        if (value === "zero") return balance === 0;
        return true;
      },
    },
    {
      id: "actions",
      cell: ({ row }) => {
        const player = row.original;
        const hasAnyAction = canUpdate || canDelete;
        if (!hasAnyAction) return null;
        return (
          <div className="flex justify-end" onClick={(e) => e.stopPropagation()}>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  className="flex size-8 p-0 data-[state=open]:bg-muted"
                >
                  <MoreVerticalIcon className="size-4" />
                  <span className="sr-only">Menü öffnen</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem
                  onClick={() => navigate(`/app/players/${player.id}`)}
                >
                  <EyeIcon className="mr-2 size-4" />
                  Details
                </DropdownMenuItem>
                {canUpdate && (
                  <DropdownMenuItem onClick={() => onEdit(player)}>
                    <EditIcon className="mr-2 size-4" />
                    Bearbeiten
                  </DropdownMenuItem>
                )}
                {canUpdate && canDelete && <DropdownMenuSeparator />}
                {canDelete && (
                  <DropdownMenuItem
                    className="text-red-600"
                    onClick={() => onDelete(player)}
                  >
                    <TrashIcon className="mr-2 size-4" />
                    Löschen
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        );
      },
    },
  ];

  // Apply filters
  const filteredData = React.useMemo(() => {
    let filtered = [...players];

    // Role filter
    if (roleFilter !== "all") {
      if (roleFilter === "none") {
        filtered = filtered.filter((p) => p.role_id === null);
      } else {
        filtered = filtered.filter((p) => p.role_id === roleFilter);
      }
    }

    // Balance filter
    if (balanceFilter !== "all") {
      if (balanceFilter === "positive") {
        filtered = filtered.filter((p) => p.balance > 0);
      } else if (balanceFilter === "negative") {
        filtered = filtered.filter((p) => p.balance < 0);
      } else if (balanceFilter === "zero") {
        filtered = filtered.filter((p) => p.balance === 0);
      }
    }

    // Hide inactive
    if (hideInactive) {
      filtered = filtered.filter((p) => !p.inactive);
    }

    return filtered;
  }, [players, roleFilter, balanceFilter, hideInactive]);

  const table = useReactTable({
    data: filteredData,
    columns,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      pagination,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  return (
    <div className="flex flex-col gap-4">
      {/* Filters and Actions */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div className="flex flex-1 flex-col gap-4 md:flex-row md:items-center">
          <Input
            placeholder="Nach Name suchen..."
            value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
            onChange={(event) =>
              table.getColumn("name")?.setFilterValue(event.target.value)
            }
            className="max-w-sm"
          />
          <div className="flex gap-2">
            <Select value={roleFilter} onValueChange={setRoleFilter}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Rolle filtern" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Alle Rollen</SelectItem>
                {roles.map((role) => (
                  <SelectItem key={role.id} value={role.id}>
                    {role.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={balanceFilter} onValueChange={setBalanceFilter}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Balance filtern" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Alle Balances</SelectItem>
                <SelectItem value="positive">Positiv</SelectItem>
                <SelectItem value="negative">Negativ</SelectItem>
                <SelectItem value="zero">Null</SelectItem>
              </SelectContent>
            </Select>
            <div className="flex items-center gap-2">
              <Checkbox
                id="hide-inactive"
                checked={hideInactive}
                onCheckedChange={(checked) => setHideInactive(checked === true)}
              />
              <Label htmlFor="hide-inactive" className="cursor-pointer text-sm font-normal">
                Inaktive ausblenden
              </Label>
            </div>
          </div>
        </div>
        {canCreate && (
          <Button onClick={onCreate}>
            <PlusIcon className="mr-2 size-4" />
            Player hinzufügen
          </Button>
        )}
      </div>

      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  Laden...
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  className="cursor-pointer"
                  onClick={() => onView(row.original)}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  Keine Player gefunden.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between px-2">
        <div className="flex-1 text-sm text-muted-foreground">
          {filteredData.length} Player
          {filteredData.length !== players.length &&
            ` (${players.length} gesamt)`}
        </div>
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-2">
            <Label htmlFor="rows-per-page" className="text-sm">
              Zeilen pro Seite
            </Label>
            <Select
              value={`${table.getState().pagination.pageSize}`}
              onValueChange={(value) => {
                table.setPageSize(Number(value));
              }}
            >
              <SelectTrigger className="w-[70px]" id="rows-per-page">
                <SelectValue placeholder={table.getState().pagination.pageSize} />
              </SelectTrigger>
              <SelectContent side="top">
                {[10, 20, 30, 50].map((pageSize) => (
                  <SelectItem key={pageSize} value={`${pageSize}`}>
                    {pageSize}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex items-center gap-2">
            <div className="text-sm">
              Seite {table.getState().pagination.pageIndex + 1} von{" "}
              {table.getPageCount()}
            </div>
            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="icon"
                className="size-8"
                onClick={() => table.setPageIndex(0)}
                disabled={!table.getCanPreviousPage()}
              >
                <ChevronsLeftIcon className="size-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                className="size-8"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
              >
                <ChevronLeftIcon className="size-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                className="size-8"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
              >
                <ChevronRightIcon className="size-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                className="size-8"
                onClick={() => table.setPageIndex(table.getPageCount() - 1)}
                disabled={!table.getCanNextPage()}
              >
                <ChevronsRightIcon className="size-4" />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

