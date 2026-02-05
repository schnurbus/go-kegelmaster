import * as React from "react";
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
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronsLeftIcon,
  ChevronsRightIcon,
  EditIcon,
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

import type { Competition, ScoringType } from "@/types/competition";

const SCORING_LABELS: Record<ScoringType, string> = {
  winner: "Gewinner",
  loser: "Verlierer",
  both: "Beides",
};

const COMPETITIONS_STORAGE_KEY_PREFIX = "kegelmaster_competitions_";

type CompetitionsStorage = {
  pageIndex?: number;
  pageSize?: number;
  sorting?: SortingState;
};

function loadCompetitionsStorage(clubId: string): Partial<CompetitionsStorage> {
  try {
    const raw = localStorage.getItem(COMPETITIONS_STORAGE_KEY_PREFIX + clubId);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as Partial<CompetitionsStorage>;
    return {
      pageIndex: typeof parsed.pageIndex === "number" && parsed.pageIndex >= 0 ? parsed.pageIndex : undefined,
      pageSize: typeof parsed.pageSize === "number" && parsed.pageSize >= 1 ? parsed.pageSize : undefined,
      sorting: Array.isArray(parsed.sorting) ? parsed.sorting : undefined,
    };
  } catch {
    return {};
  }
}

function saveCompetitionsStorage(clubId: string, data: CompetitionsStorage) {
  try {
    localStorage.setItem(COMPETITIONS_STORAGE_KEY_PREFIX + clubId, JSON.stringify(data));
  } catch {
    // ignore
  }
}

type CompetitionsDataTableProps = {
  competitions: Competition[];
  onEdit: (competition: Competition) => void;
  onDelete: (competition: Competition) => void;
  onCreate: () => void;
  isLoading?: boolean;
  clubId?: string;
  canCreate?: boolean;
  canUpdate?: boolean;
  canDelete?: boolean;
};

export function CompetitionsDataTable({
  competitions,
  onEdit,
  onDelete,
  onCreate,
  isLoading = false,
  clubId,
  canCreate = true,
  canUpdate = true,
  canDelete = true,
}: CompetitionsDataTableProps) {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [pagination, setPagination] = React.useState({
    pageIndex: 0,
    pageSize: 10,
  });

  React.useEffect(() => {
    if (!clubId) return;
    const stored = loadCompetitionsStorage(clubId);
    if (stored.sorting !== undefined) setSorting(stored.sorting);
    if (stored.pageIndex !== undefined) setPagination((p) => ({ ...p, pageIndex: stored.pageIndex! }));
    if (stored.pageSize !== undefined) setPagination((p) => ({ ...p, pageSize: stored.pageSize! }));
  }, [clubId]);

  React.useEffect(() => {
    if (!clubId) return;
    saveCompetitionsStorage(clubId, {
      pageIndex: pagination.pageIndex,
      pageSize: pagination.pageSize,
      sorting,
    });
  }, [clubId, pagination.pageIndex, pagination.pageSize, sorting]);

  const columns: ColumnDef<Competition>[] = [
    {
      accessorKey: "display_order",
      header: () => <div className="text-center">Reihenfolge</div>,
      cell: ({ row }) => (
        <div className="text-center">{row.getValue("display_order")}</div>
      ),
    },
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <div className="font-medium">{row.getValue("name")}</div>
      ),
    },
    {
      accessorKey: "scoring_type",
      header: "Wertung",
      cell: ({ row }) => {
        const st = row.getValue("scoring_type") as ScoringType;
        return (
          <div className="text-sm">
            {SCORING_LABELS[st] ?? st}
          </div>
        );
      },
    },
    {
      accessorKey: "is_gender_specific",
      header: "Geschlechtsspezifisch",
      cell: ({ row }) => (
        <div className="text-sm">
          {row.getValue("is_gender_specific") ? "Ja" : "Nein"}
        </div>
      ),
    },
    {
      accessorKey: "created_at",
      header: "Erstellt am",
      cell: ({ row }) => {
        const date = new Date(row.getValue("created_at"));
        return (
          <div className="text-sm text-muted-foreground">
            {date.toLocaleDateString("de-DE")}
          </div>
        );
      },
    },
    {
      id: "actions",
      cell: ({ row }) => {
        const competition = row.original;
        if (!canUpdate && !canDelete) return null;
        return (
          <div className="flex justify-end">
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
                {canUpdate && (
                  <DropdownMenuItem onClick={() => onEdit(competition)}>
                    <EditIcon className="mr-2 size-4" />
                    Bearbeiten
                  </DropdownMenuItem>
                )}
                {canUpdate && canDelete && <DropdownMenuSeparator />}
                {canDelete && (
                  <DropdownMenuItem
                    className="text-red-600"
                    onClick={() => onDelete(competition)}
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

  const table = useReactTable({
    data: competitions,
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
        </div>
        {canCreate && (
          <Button onClick={onCreate}>
            <PlusIcon className="mr-2 size-4" />
            Wettbewerb hinzufügen
          </Button>
        )}
      </div>

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
                  Keine Wettbewerbe gefunden.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between px-2">
        <div className="flex-1 text-sm text-muted-foreground">
          {competitions.length} Wettbewerb{competitions.length !== 1 ? "e" : ""}
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
