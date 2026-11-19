import { useState } from "react";
import { NavLink, Outlet } from "react-router-dom";
import { DownloadIcon } from "lucide-react";

import { AdminReauthDialog } from "@/components/admin/admin-reauth-dialog";
import { AdminNotFound } from "@/components/admin/admin-ui";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/context/AuthContext";
import { MEMBERSHIPS_CSV_URL } from "@/lib/admin";
import { cn } from "@/lib/utils";

const NAV = [
  { to: "/admin", label: "Übersicht", end: true },
  { to: "/admin/users", label: "Benutzer", end: false },
  { to: "/admin/clubs", label: "Klubs", end: false },
];

/**
 * Operator area (instance admins only). Independent of the selected club: nothing
 * here reads ClubContext. Non-admins get the same "not found" page the API implies
 * (it answers 404 for them), without any admin request being sent.
 */
function AdminLayout() {
  const { user } = useAuth();
  // Bumped after a re-authentication: remounts the page, which loads its data again.
  const [reauthCount, setReauthCount] = useState(0);

  if (!user?.is_instance_admin) {
    return <AdminNotFound />;
  }

  return (
    <div className="flex flex-col gap-4 px-4 py-4 md:px-6 md:py-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <nav aria-label="Administration" className="flex gap-1 overflow-x-auto rounded-lg bg-muted p-1">
          {NAV.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                cn(
                  "whitespace-nowrap rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground",
                )
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <Button asChild variant="outline" size="sm" className="w-full sm:w-auto">
          <a href={MEMBERSHIPS_CSV_URL} download>
            <DownloadIcon />
            Verknüpfungen als CSV
          </a>
        </Button>
      </div>
      <Outlet key={reauthCount} />
      <AdminReauthDialog onConfirmed={() => setReauthCount((n) => n + 1)} />
    </div>
  );
}

export default AdminLayout;
