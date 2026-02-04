import * as React from "react";
import { useNavigate } from "react-router-dom";
import { useClub } from "@/context/ClubContext";
import { useDashboardCompetitionData } from "@/hooks/use-dashboard-competition-data";
import {
  type ColumnDef,
  type SortingState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronsLeftIcon,
  ChevronsRightIcon,
  GiftIcon,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import type { Transaction, TransactionType } from "@/types/transaction";
import {
  getTransactionTypeLabel,
  getTransactionTypeColor,
} from "@/types/transaction";
import { formatCentsToEuro } from "@/types/player";
import type { PaginatedTransactions } from "@/types/transaction";

const PAGE_SIZE_OPTIONS = [10, 25, 50] as const;
const DEFAULT_PAGE_SIZE = 10;

export function DashboardTransactionTable() {
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const { myPlayer } = useDashboardCompetitionData();

  const [transactions, setTransactions] = React.useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = React.useState(false);
  const [currentPage, setCurrentPage] = React.useState(1);
  const [totalPages, setTotalPages] = React.useState(1);
  const [total, setTotal] = React.useState(0);
  const [pageSize, setPageSize] = React.useState(DEFAULT_PAGE_SIZE);
  const [sorting, setSorting] = React.useState<SortingState>([]);

  const fetchTransactions = React.useCallback(
    async (page: number = 1) => {
      if (!activeClub || !myPlayer) return;

      setIsLoading(true);
      try {
        const url = `/api/clubs/${activeClub.id}/players/${myPlayer.id}/transactions?page=${page}&limit=${pageSize}`;
        const response = await fetch(url, { credentials: "include" });

        if (!response.ok) {
          throw new Error("Fehler beim Laden der Transaktionen");
        }

        const data: PaginatedTransactions = await response.json();
        setTransactions(data.data || []);
        setCurrentPage(data.page);
        setTotalPages(data.total_pages);
        setTotal(data.total);
      } catch (error) {
        console.error("Error fetching transactions:", error);
        toast.error("Fehler beim Laden der Transaktionen");
        setTransactions([]);
      } finally {
        setIsLoading(false);
      }
    },
    [activeClub, myPlayer, pageSize]
  );

  React.useEffect(() => {
    if (activeClub && myPlayer) {
      setCurrentPage(1);
      fetchTransactions(1);
    }
  }, [activeClub, myPlayer?.id, pageSize, fetchTransactions]);

  const columns: ColumnDef<Transaction>[] = [
    {
      accessorKey: "transaction_date",
      header: "Datum",
      cell: ({ row }) => {
        const dateStr = (row.getValue("transaction_date") ??
          row.original.created_at) as string;
        const date = new Date(dateStr);
        return (
          <div className="font-medium">
            {date.toLocaleDateString("de-DE", {
              day: "2-digit",
              month: "2-digit",
              year: "numeric",
            })}
          </div>
        );
      },
    },
    {
      accessorKey: "transaction_type",
      header: "Typ",
      cell: ({ row }) => {
        const type = row.getValue("transaction_type") as TransactionType;
        const label = getTransactionTypeLabel(type);
        const color = getTransactionTypeColor(type);
        const isAutoTip =
          type === "tip" && row.original.description?.includes("Auto-Tip");

        return (
          <div className="flex items-center gap-2">
            <Badge variant={color as React.ComponentProps<typeof Badge>["variant"]}>{label}</Badge>
            {isAutoTip && (
              <span title="Auto-Tip">
                <GiftIcon className="size-4 text-blue-500" />
              </span>
            )}
          </div>
        );
      },
    },
    {
      accessorKey: "amount",
      header: () => <div className="text-right">Betrag</div>,
      cell: ({ row }) => {
        const amount = row.getValue("amount") as number;
        const formatted = formatCentsToEuro(amount);
        return (
          <div
            className={`text-right font-medium ${
              amount < 0 ? "text-red-500" : amount > 0 ? "text-green-500" : ""
            }`}
          >
            {formatted}
          </div>
        );
      },
    },
    {
      accessorKey: "description",
      header: "Beschreibung",
      cell: ({ row }) => {
        const description = row.getValue("description") as string;
        return (
          <div className="max-w-md truncate text-muted-foreground">
            {description}
          </div>
        );
      },
    },
    {
      id: "source",
      header: "Quelle",
      cell: ({ row }) => {
        const gameDayId = row.original.game_day_id;
        if (!gameDayId) {
          return <span className="text-muted-foreground">-</span>;
        }
        return (
          <Button
            variant="link"
            className="h-auto p-0 text-primary"
            onClick={() => navigate(`/app/gamedays/${gameDayId}`)}
          >
            Spieltag
          </Button>
        );
      },
    },
  ];

  const table = useReactTable({
    data: transactions,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    manualPagination: true,
    pageCount: totalPages,
  });

  // Kein eigener Spieler für den Club → Tabelle nicht rendern
  if (!activeClub || !myPlayer) {
    return null;
  }

  return (
    <section className="px-4 lg:px-6">
      <h2 className="mb-4 text-lg font-semibold">Meine Transaktionen</h2>
      <div className="flex flex-col gap-4">
        <div className="overflow-hidden rounded-lg border">
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
                  <TableCell
                    colSpan={columns.length}
                    className="h-24 text-center"
                  >
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
                    Keine Transaktionen gefunden.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <div className="text-sm text-muted-foreground">
              Gesamt: {total} Transaktionen
            </div>
            <div className="flex items-center gap-2">
              <Label htmlFor="dashboard-rows-per-page" className="whitespace-nowrap text-sm">
                Zeilen pro Seite
              </Label>
              <Select
                value={String(pageSize)}
                onValueChange={(v) =>
                  setPageSize(Number(v) as (typeof PAGE_SIZE_OPTIONS)[number])
                }
              >
                <SelectTrigger id="dashboard-rows-per-page" className="w-[70px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent side="top">
                  {PAGE_SIZE_OPTIONS.map((size) => (
                    <SelectItem key={size} value={String(size)}>
                      {size}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex items-center gap-8">
            <div className="text-center text-sm font-medium">
              Seite {currentPage} von {totalPages}
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                className="hidden size-8 p-0 lg:flex"
                onClick={() => fetchTransactions(1)}
                disabled={currentPage === 1}
              >
                <span className="sr-only">Zur ersten Seite</span>
                <ChevronsLeftIcon />
              </Button>
              <Button
                variant="outline"
                className="size-8"
                size="icon"
                onClick={() => fetchTransactions(currentPage - 1)}
                disabled={currentPage === 1}
              >
                <span className="sr-only">Vorherige Seite</span>
                <ChevronLeftIcon />
              </Button>
              <Button
                variant="outline"
                className="size-8"
                size="icon"
                onClick={() => fetchTransactions(currentPage + 1)}
                disabled={currentPage === totalPages}
              >
                <span className="sr-only">Nächste Seite</span>
                <ChevronRightIcon />
              </Button>
              <Button
                variant="outline"
                className="hidden size-8 lg:flex"
                size="icon"
                onClick={() => fetchTransactions(totalPages)}
                disabled={currentPage === totalPages}
              >
                <span className="sr-only">Zur letzten Seite</span>
                <ChevronsRightIcon />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
