import { Outlet, useMatches } from "react-router-dom";
import { AppFooter } from "@/components/AppFooter";
import { AppSidebar } from "@/components/app-sidebar";
import { SiteHeader } from "@/components/site-header";
import {
  SidebarInset,
  SidebarProvider,
} from "@/components/ui/sidebar";
import { PageTitleProvider, usePageTitle } from "@/context/PageTitleContext";

function DashboardLayoutContent() {
  const { title } = usePageTitle();

  return (
    <>
      <AppSidebar variant="inset" />
      <SidebarInset>
        <SiteHeader title={title ?? "Dashboard"} />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <Outlet />
          </div>
        </div>
        <AppFooter className="app-footer border-t border-border px-4 py-2 text-center text-sm text-muted-foreground" />
      </SidebarInset>
    </>
  );
}

export function DashboardLayout() {
  const matches = useMatches();
  const routeTitle =
    (matches.find((m) => m.handle && typeof (m.handle as { title?: string }).title === "string")?.handle as
      | { title: string }
      | undefined)?.title ?? "Dashboard";

  return (
    <PageTitleProvider routeTitle={routeTitle}>
      <SidebarProvider
        style={
          {
            "--sidebar-width": "calc(var(--spacing) * 72)",
            "--header-height": "calc(var(--spacing) * 12)",
          } as React.CSSProperties
        }
      >
        <DashboardLayoutContent />
      </SidebarProvider>
    </PageTitleProvider>
  );
}
