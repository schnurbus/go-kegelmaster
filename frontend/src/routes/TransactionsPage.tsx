import * as React from "react";
import { useNavigate } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { TransactionDialog } from "@/components/transaction-dialog";
import { useClub } from "@/context/ClubContext";
import {
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronsLeftIcon,
  ChevronsRightIcon,
  GiftIcon,
  PlusIcon,
  TrashIcon,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
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

import type {
  Transaction,
  PaginatedTransactions,
  TransactionType,
} from "@/types/transaction";
import {
  getTransactionTypeLabel,
  getTransactionTypeColor,
} from "@/types/transaction";
import { formatCentsToEuro } from "@/types/player";

function TransactionsPage() {
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const [transactions, setTransactions] = React.useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isDialogOpen, setIsDialogOpen] = React.useState(false);
  const [currentPage, setCurrentPage] = React.useState(1);
  const [totalPages, setTotalPages] = React.useState(1);
  const [total, setTotal] = React.useState(0);

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [typeFilter, setTypeFilter] = React.useState<string>("all");

  const fetchTransactions = React.useCallback(
    async (page: number = 1) => {
      if (!activeClub) return;

      setIsLoading(true);
      try {
        const response = await fetch(
          `/api/clubs/${activeClub.id}/transactions?page=${page}&limit=50`,
          {
            credentials: "include",
          }
        );

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
    [activeClub]
  );

  React.useEffect(() => {
    if (activeClub) {
      fetchTransactions(1);
    }
  }, [activeClub, fetchTransactions]);

  const handleDelete = async (transaction: Transaction) => {
    if (!activeClub) return;

    // Only manual transactions can be deleted
    if (
      transaction.transaction_type === "base_fee" ||
      transaction.transaction_type === "fee"
    ) {
      toast.error("Automatische Transaktionen können nicht gelöscht werden");
      return;
    }

    if (!confirm("Transaktion wirklich löschen?")) return;

    try {
      const response = await fetch(
        `/api/clubs/${activeClub.id}/transactions/${transaction.id}`,
        {
          method: "DELETE",
          headers: {
            "X-CSRF-Token":
              document.cookie
                .split("; ")
                .find((row) => row.startsWith("csrf_token="))
                ?.split("=")[1] || "",
          },
          credentials: "include",
        }
      );

      if (!response.ok) {
        throw new Error("Fehler beim Löschen der Transaktion");
      }

      toast.success("Transaktion gelöscht");
      fetchTransactions(currentPage);
    } catch (error) {
      console.error("Error deleting transaction:", error);
      toast.error("Fehler beim Löschen der Transaktion");
    }
  };

  const columns: ColumnDef<Transaction>[] = [
    {
      accessorKey: "created_at",
      header: "Datum",
      cell: ({ row }) => {
        const date = new Date(row.getValue("created_at"));
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
      accessorKey: "player_name",
      header: "Spieler",
      cell: ({ row }) => {
        const playerName = row.getValue("player_name") as string | undefined;
        return (
          <div>{playerName || <span className="text-muted-foreground">-</span>}</div>
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
        
        // Show auto-tip indicator
        const isAutoTip =
          type === "tip" && row.original.description?.includes("Auto-Tip");

        return (
          <div className="flex items-center gap-2">
            <Badge variant={color as any}>{label}</Badge>
            {isAutoTip && (
              <span title="Auto-Tip">
                <GiftIcon className="size-4 text-blue-500" />
              </span>
            )}
          </div>
        );
      },
      filterFn: (row, _id, value) => {
        if (value === "all") return true;
        return row.original.transaction_type === value;
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
    {
      id: "actions",
      cell: ({ row }) => {
        const transaction = row.original;
        const canDelete =
          transaction.transaction_type !== "base_fee" &&
          transaction.transaction_type !== "fee";

        if (!canDelete) return null;

        return (
          <div className="flex justify-end">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  className="size-8 p-0"
                  onClick={(e) => e.stopPropagation()}
                >
                  <span className="sr-only">Aktionen öffnen</span>
                  <TrashIcon className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem
                  onClick={() => handleDelete(transaction)}
                  className="text-destructive"
                >
                  <TrashIcon className="mr-2 size-4" />
                  Löschen
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        );
      },
    },
  ];

  const table = useReactTable({
    data: transactions,
    columns,
    state: {
      sorting,
      columnVisibility,
      columnFilters,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    manualPagination: true,
    pageCount: totalPages,
  });

  // Apply type filter
  React.useEffect(() => {
    if (typeFilter === "all") {
      table.getColumn("transaction_type")?.setFilterValue(undefined);
    } else {
      table.getColumn("transaction_type")?.setFilterValue(typeFilter);
    }
  }, [typeFilter, table]);

  const handleSuccess = () => {
    fetchTransactions(currentPage);
  };

  if (!activeClub) {
    return (
      <AppLayout title="Transaktionen">
        <div className="flex flex-col items-center justify-center py-12">
          <p className="text-muted-foreground">
            Bitte wählen Sie einen Club aus, um die Transaktionen zu sehen.
          </p>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout title="Transaktionen">
      <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
        <div className="px-4 lg:px-6">
          <div className="flex flex-col gap-4">
            {/* Header with filters */}
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex flex-1 items-center gap-2">
                <Input
                  placeholder="Beschreibung durchsuchen..."
                  value={
                    (table.getColumn("description")?.getFilterValue() as string) ??
                    ""
                  }
                  onChange={(event) =>
                    table
                      .getColumn("description")
                      ?.setFilterValue(event.target.value)
                  }
                  className="max-w-sm"
                />
                <Select value={typeFilter} onValueChange={setTypeFilter}>
                  <SelectTrigger className="w-48">
                    <SelectValue placeholder="Alle Typen" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Alle Typen</SelectItem>
                    <SelectItem value="base_fee">Grundgebühr</SelectItem>
                    <SelectItem value="fee">Strafgebühr</SelectItem>
                    <SelectItem value="deposit">Einzahlung</SelectItem>
                    <SelectItem value="tip">Trinkgeld</SelectItem>
                    <SelectItem value="expense">Ausgabe</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Button onClick={() => setIsDialogOpen(true)}>
                <PlusIcon className="mr-2 size-4" />
                Transaktion erstellen
              </Button>
            </div>

            {/* Table */}
            <div className="overflow-hidden rounded-lg border">
              <Table>
                <TableHeader>
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => {
                        return (
                          <TableHead key={header.id}>
                            {header.isPlaceholder
                              ? null
                              : flexRender(
                                  header.column.columnDef.header,
                                  header.getContext()
                                )}
                          </TableHead>
                        );
                      })}
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

            {/* Pagination */}
            <div className="flex items-center justify-between px-4">
              <div className="text-sm text-muted-foreground">
                Gesamt: {total} Transaktionen
              </div>
              <div className="flex items-center gap-8">
                <div className="flex items-center justify-center text-sm font-medium">
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
        </div>
      </div>

      <TransactionDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        clubId={activeClub.id}
        onSuccess={handleSuccess}
      />
    </AppLayout>
  );
}

export default TransactionsPage;
