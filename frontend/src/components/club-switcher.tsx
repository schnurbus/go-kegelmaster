"use client"

import * as React from "react"
import { ChevronsUpDown, Plus } from "lucide-react"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"
import { CreateClubDialog } from "./create-club-dialog"
import { useClub } from "@/context/ClubContext"
import { cn } from "@/lib/utils"

export function ClubSwitcher() {
  const { isMobile } = useSidebar()
  const {
    clubs,
    activeClub,
    setActiveClub,
    refreshClubs,
    isLoading,
  } = useClub()
  const [isDialogOpen, setIsDialogOpen] = React.useState(false)

  const handleClubCreated = () => {
    refreshClubs()
    setIsDialogOpen(false)
  }

  if (isLoading) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" disabled>
            <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
              <div className="size-4 animate-pulse bg-current opacity-50" />
            </div>
            <div className="grid flex-1 text-left text-sm leading-tight">
              <span className="truncate font-medium">Lade...</span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    )
  }

  if (!activeClub && clubs.length === 0) {
    return (
      <>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              asChild={false}
              onClick={() => setIsDialogOpen(true)}
            >
              <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                <Plus className="size-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">Neuer Club</span>
                <span className="truncate text-xs">Erstellen</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <CreateClubDialog
          open={isDialogOpen}
          onOpenChange={setIsDialogOpen}
          onClubCreated={handleClubCreated}
        />
      </>
    )
  }

  return (
    <>
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className={cn(
                  "peer/menu-button flex w-full items-center gap-2 overflow-hidden rounded-md p-2 text-left text-sm outline-hidden ring-sidebar-ring transition-[width,height,padding] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-2 active:bg-sidebar-accent active:text-sidebar-accent-foreground disabled:pointer-events-none disabled:opacity-50 data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground h-12 group-data-[collapsible=icon]:size-8! group-data-[collapsible=icon]:p-2! [&>span:last-child]:truncate [&>svg]:size-4 [&>svg]:shrink-0"
                )}
              >
                <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  {activeClub?.name.charAt(0).toUpperCase() || "C"}
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">
                    {activeClub?.name || "Kein Club"}
                  </span>
                  <span className="truncate text-xs">
                    {activeClub
                      ? `${(activeClub.balance / 100).toFixed(2)} €`
                      : "Auswählen"}
                  </span>
                </div>
                <ChevronsUpDown className="ml-auto size-4" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="min-w-56 rounded-lg"
              align="start"
              side={isMobile ? "bottom" : "right"}
              sideOffset={4}
            >
              <DropdownMenuLabel className="text-muted-foreground text-xs">
                Clubs
              </DropdownMenuLabel>
              {clubs.length === 0 ? (
                <DropdownMenuItem disabled className="text-muted-foreground">
                  Keine Clubs vorhanden
                </DropdownMenuItem>
              ) : (
                clubs.map((club) => (
                  <DropdownMenuItem
                    key={club.id}
                    onClick={() => setActiveClub(club)}
                    className="gap-2 p-2"
                  >
                    <div className="flex size-6 items-center justify-center rounded-md border bg-sidebar-primary text-sidebar-primary-foreground">
                      {club.name.charAt(0).toUpperCase()}
                    </div>
                    <div className="flex-1">
                      <div className="font-medium">{club.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {(club.balance / 100).toFixed(2)} €
                      </div>
                    </div>
                  </DropdownMenuItem>
                ))
              )}
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="gap-2 p-2"
                onClick={() => setIsDialogOpen(true)}
              >
                <div className="flex size-6 items-center justify-center rounded-md border bg-transparent">
                  <Plus className="size-4" />
                </div>
                <div className="text-muted-foreground font-medium">
                  Neuen Club erstellen
                </div>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
      <CreateClubDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        onClubCreated={handleClubCreated}
      />
    </>
  )
}

