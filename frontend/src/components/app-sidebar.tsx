import * as React from "react"
import {
  ArrowUpCircleIcon,
  BarChartIcon,
  CalendarIcon,
  CameraIcon,
  ClipboardListIcon,
  FileCodeIcon,
  FileTextIcon,
  HelpCircleIcon,
  LayoutDashboardIcon,
  ReceiptIcon,
  SettingsIcon,
  ShieldIcon,
  UsersIcon,
  AlertCircleIcon,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
import { ClubSwitcher } from "@/components/club-switcher"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

import { Link } from "react-router-dom"
import { useAuth } from "@/context/AuthContext"
import { useClub } from "@/context/ClubContext"
import { usePermissions } from "@/hooks/use-permissions"

function useDashboardData() {
  const { user } = useAuth()
  const { activeClub } = useClub()
  const { isOwner, canList, canView } = usePermissions(activeClub?.id ?? null)
  const showRoles =
    (activeClub && user && (isOwner || canList("roles") || canView("roles"))) ?? false

  const navMain = [
    {
      title: "Dashboard",
      url: "/app",
      icon: LayoutDashboardIcon,
    },
    {
      title: "Statistiken",
      url: "#",
      icon: BarChartIcon,
    },
    {
      title: "Spieltage",
      url: "/app/gamedays",
      icon: CalendarIcon,
    },
    {
      title: "Spieler",
      url: "/app/players",
      icon: UsersIcon,
    },
    {
      title: "Transaktionen",
      url: "/app/transactions",
      icon: ReceiptIcon,
    },
  ]

  return {
    user: {
      name: user?.email?.split("@")[0] || "User",
      email: user?.email || "",
      avatar: "",
    },
    navMain,
    activeClub,
    navClouds: [
      {
        title: "Capture",
        icon: CameraIcon,
        isActive: true,
        url: "#",
        items: [
          {
            title: "Active Proposals",
            url: "#",
          },
          {
            title: "Archived",
            url: "#",
          },
        ],
      },
      {
        title: "Proposal",
        icon: FileTextIcon,
        url: "#",
        items: [
          {
            title: "Active Proposals",
            url: "#",
          },
          {
            title: "Archived",
            url: "#",
          },
        ],
      },
      {
        title: "Prompts",
        icon: FileCodeIcon,
        url: "#",
        items: [
          {
            title: "Active Proposals",
            url: "#",
          },
          {
            title: "Archived",
            url: "#",
          },
        ],
      },
    ],
    navSecondary: [
      {
        title: "Hilfe",
        url: "/app/help",
        icon: HelpCircleIcon,
      },
    ],
    management: [
      ...(showRoles ? [{ title: "Rollen", url: "/app/roles", icon: ShieldIcon }] : []),
      {
        title: "Strafentypen",
        url: "/app/penalty-types",
        icon: AlertCircleIcon,
      },
      {
        title: "Wettbewerbe",
        url: "/app/competitions",
        icon: ClipboardListIcon,
      },
      // {
      //   name: "Data Library",
      //   url: "#",
      //   icon: DatabaseIcon,
      // },
      // {
      //   name: "Reports",
      //   url: "#",
      //   icon: ClipboardListIcon,
      // },
      // {
      //   name: "Word Assistant",
      //   url: "#",
      //   icon: FileIcon,
      // },
    ],
  }
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const data = useDashboardData()
  const { user: authUser } = useAuth()

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:!p-1.5"
            >
              <a href="/app">
                <ArrowUpCircleIcon className="h-5 w-5" />
                <span className="text-base font-semibold">Kegelmaster</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <ClubSwitcher />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        {data.activeClub && authUser && data.activeClub.user_id === authUser.id && (
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton tooltip="Club bearbeiten" asChild>
                    <Link to={`/app/club/${data.activeClub.id}`}>
                      <SettingsIcon />
                      <span>Club bearbeiten</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        )}
        <NavSecondary items={data.management} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
    </Sidebar>
  )
}
