import * as React from "react"
import {
  ArrowUpCircleIcon,
  BarChartIcon,
  CalendarIcon,
  CameraIcon,
  ClipboardListIcon,
  DatabaseIcon,
  FileCodeIcon,
  FileIcon,
  FileTextIcon,
  FolderIcon,
  HelpCircleIcon,
  LayoutDashboardIcon,
  ReceiptIcon,
  SearchIcon,
  SettingsIcon,
  ShieldIcon,
  UsersIcon,
  AlertCircleIcon,
} from "lucide-react"

import { NavDocuments } from "@/components/nav-documents"
import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
import { ClubSwitcher } from "@/components/club-switcher"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

import { useAuth } from "@/context/AuthContext"

function useDashboardData() {
  const { user } = useAuth()
  
  return {
    user: {
      name: user?.email?.split("@")[0] || "User",
      email: user?.email || "",
      avatar: "",
    },
    navMain: [
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
      // {
      //   title: "Klubs",
      //   url: "#",
      //   icon: FolderIcon,
      // },
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
    ],
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
        title: "Settings",
        url: "#",
        icon: SettingsIcon,
      },
      {
        title: "Get Help",
        url: "#",
        icon: HelpCircleIcon,
      },
      {
        title: "Search",
        url: "#",
        icon: SearchIcon,
      },
    ],
    management: [
      {
        title: "Rollen",
        url: "/app/roles",
        icon: ShieldIcon,
      },
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
        <NavSecondary items={data.management} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
    </Sidebar>
  )
}
