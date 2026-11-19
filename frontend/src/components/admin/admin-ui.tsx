import { useEffect, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import {
  ArrowDownIcon,
  ArrowUpDownIcon,
  ArrowUpIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  SearchIcon,
  SearchXIcon,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { TableHead } from "@/components/ui/table";
import type { AdminApiError } from "@/lib/admin";

/** Shown for /admin* to non-admins and for unknown IDs (the API answers 404 in both cases). */
export function AdminNotFound({ message }: { message?: string }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 px-4 py-16 text-center">
      <SearchXIcon className="size-10 text-muted-foreground" />
      <h2 className="text-xl font-semibold">Seite nicht gefunden</h2>
      <p className="max-w-md text-sm text-muted-foreground">
        {message ?? "Diese Seite existiert nicht oder Sie haben keinen Zugriff darauf."}
      </p>
      <Button asChild variant="outline">
        <Link to="/app">Zum Dashboard</Link>
      </Button>
    </div>
  );
}

export function AdminLoadError({ error }: { error: AdminApiError }) {
  if (error.status === 404) return <AdminNotFound />;
  return (
    <p role="alert" className="py-8 text-center text-sm text-destructive">
      {error.message}
    </p>
  );
}

export function AdminLoading() {
  return <p className="py-8 text-center text-sm text-muted-foreground">Lädt …</p>;
}

export function AdminEmpty({ children }: { children: ReactNode }) {
  return <p className="py-6 text-center text-sm text-muted-foreground">{children}</p>;
}

export function AdminPagination({
  page,
  totalPages,
  total,
  onPageChange,
}: {
  page: number;
  totalPages: number;
  total: number;
  onPageChange: (page: number) => void;
}) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-2 pt-4 text-sm text-muted-foreground">
      <span>
        {total} {total === 1 ? "Eintrag" : "Einträge"}
      </span>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
          aria-label="Vorherige Seite"
        >
          <ChevronLeftIcon />
        </Button>
        <span>
          Seite {Math.min(page, Math.max(totalPages, 1))} von {Math.max(totalPages, 1)}
        </span>
        <Button
          variant="outline"
          size="sm"
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
          aria-label="Nächste Seite"
        >
          <ChevronRightIcon />
        </Button>
      </div>
    </div>
  );
}

/** Table header cell that toggles server-side sorting. */
export function SortableHead({
  label,
  field,
  sort,
  order,
  onSort,
  className,
}: {
  label: string;
  field: string;
  sort: string;
  order: "asc" | "desc";
  onSort: (field: string) => void;
  className?: string;
}) {
  const active = sort === field;
  const Icon = !active ? ArrowUpDownIcon : order === "asc" ? ArrowUpIcon : ArrowDownIcon;
  return (
    <TableHead
      className={className}
      aria-sort={active ? (order === "asc" ? "ascending" : "descending") : "none"}
    >
      <Button variant="ghost" size="sm" className="-ml-3 h-8" onClick={() => onSort(field)}>
        {label}
        <Icon className={active ? "" : "text-muted-foreground"} />
      </Button>
    </TableHead>
  );
}

export function AdminBadge() {
  return <Badge>Admin</Badge>;
}

export function OrphanedBadge() {
  return <Badge variant="destructive">verwaist</Badge>;
}

export function OwnerBadge() {
  return <Badge variant="secondary">Owner</Badge>;
}

export function LinkedBadge({ linked }: { linked: boolean }) {
  return linked ? (
    <Badge variant="outline">verknüpft</Badge>
  ) : (
    <Badge variant="outline" className="text-muted-foreground">
      nicht verknüpft
    </Badge>
  );
}

export function InactiveBadge() {
  return (
    <Badge variant="outline" className="text-muted-foreground">
      inaktiv
    </Badge>
  );
}

/** Label/value grid for detail pages. */
export function AdminFacts({ items }: { items: { label: string; value: ReactNode }[] }) {
  return (
    <dl className="grid grid-cols-1 gap-x-6 gap-y-3 text-sm sm:grid-cols-2 lg:grid-cols-3">
      {items.map((item) => (
        <div key={item.label} className="min-w-0">
          <dt className="text-muted-foreground">{item.label}</dt>
          <dd className="break-words font-medium">{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}

/** Search field that reports its value 300 ms after the last keystroke. */
export function AdminSearchInput({
  value,
  onSearch,
  placeholder,
  label,
}: {
  value: string;
  onSearch: (value: string) => void;
  placeholder: string;
  label: string;
}) {
  const [text, setText] = useState(value);

  useEffect(() => {
    if (text.trim() === value) return;
    const timer = setTimeout(() => onSearch(text), 300);
    return () => clearTimeout(timer);
  }, [text, value, onSearch]);

  return (
    <div className="relative w-full sm:max-w-xs">
      <SearchIcon className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        type="search"
        aria-label={label}
        placeholder={placeholder}
        value={text}
        onChange={(e) => setText(e.target.value)}
        className="pl-8"
      />
    </div>
  );
}
