import { useEffect, useState } from "react"
import { TrendingDownIcon, TrendingUpIcon } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useAuth } from "@/context/AuthContext"
import { useClub } from "@/context/ClubContext"
import type { Player } from "@/types/player"
import { formatCentsToEuro } from "@/types/player"

export function SectionCards() {
  const { user, status } = useAuth()
  const { activeClub } = useClub()
  const [player, setPlayer] = useState<Player | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [hasPlayer, setHasPlayer] = useState(false)

  useEffect(() => {
    const fetchPlayer = async () => {
      if (status !== "authenticated" || !user || !activeClub) {
        setIsLoading(false)
        setHasPlayer(false)
        return
      }

      setIsLoading(true)
      try {
        const resp = await fetch(
          `/api/clubs/${activeClub.id}/players/me`,
          {
            credentials: "include",
          }
        )

        if (resp.status === 404) {
          setHasPlayer(false)
          setPlayer(null)
        } else if (!resp.ok) {
          throw new Error("Fehler beim Laden des Spielers")
        } else {
          const data: Player = await resp.json()
          setPlayer(data)
          setHasPlayer(true)
        }
      } catch (error) {
        console.error("Error fetching player:", error)
        setHasPlayer(false)
        setPlayer(null)
      } finally {
        setIsLoading(false)
      }
    }

    fetchPlayer()
  }, [user, activeClub, status])

  return (
    <div className="*:data-[slot=card]:shadow-xs @xl/main:grid-cols-2 @5xl/main:grid-cols-4 grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card dark:*:data-[slot=card]:bg-card lg:px-6">
      <Card className="@container/card">
        <CardHeader className="relative">
          <CardDescription>Mein Guthaben</CardDescription>
          <CardTitle className="@[250px]/card:text-3xl text-2xl font-semibold tabular-nums">
            {isLoading ? (
              "Laden..."
            ) : hasPlayer && player ? (
              formatCentsToEuro(player.balance)
            ) : (
              "Kein Spieler in diesem Club"
            )}
          </CardTitle>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          {hasPlayer && player && (
            <>
              <div className="line-clamp-1 flex gap-2 font-medium">
                Spieler: {player.name}
              </div>
              <div className="text-muted-foreground">
                Guthaben für {activeClub?.name}
              </div>
            </>
          )}
          {!hasPlayer && !isLoading && (
            <div className="text-muted-foreground">
              Sie haben noch keinen Spieler in diesem Club
            </div>
          )}
        </CardFooter>
      </Card>
      <Card className="@container/card">
        <CardHeader className="relative">
          <CardDescription>New Customers</CardDescription>
          <CardTitle className="@[250px]/card:text-3xl text-2xl font-semibold tabular-nums">
            1,234
          </CardTitle>
          <div className="absolute right-4 top-4">
            <Badge variant="outline" className="flex gap-1 rounded-lg text-xs">
              <TrendingDownIcon className="size-3" />
              -20%
            </Badge>
          </div>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          <div className="line-clamp-1 flex gap-2 font-medium">
            Down 20% this period <TrendingDownIcon className="size-4" />
          </div>
          <div className="text-muted-foreground">
            Acquisition needs attention
          </div>
        </CardFooter>
      </Card>
      <Card className="@container/card">
        <CardHeader className="relative">
          <CardDescription>Active Accounts</CardDescription>
          <CardTitle className="@[250px]/card:text-3xl text-2xl font-semibold tabular-nums">
            45,678
          </CardTitle>
          <div className="absolute right-4 top-4">
            <Badge variant="outline" className="flex gap-1 rounded-lg text-xs">
              <TrendingUpIcon className="size-3" />
              +12.5%
            </Badge>
          </div>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          <div className="line-clamp-1 flex gap-2 font-medium">
            Strong user retention <TrendingUpIcon className="size-4" />
          </div>
          <div className="text-muted-foreground">Engagement exceed targets</div>
        </CardFooter>
      </Card>
      <Card className="@container/card">
        <CardHeader className="relative">
          <CardDescription>Growth Rate</CardDescription>
          <CardTitle className="@[250px]/card:text-3xl text-2xl font-semibold tabular-nums">
            4.5%
          </CardTitle>
          <div className="absolute right-4 top-4">
            <Badge variant="outline" className="flex gap-1 rounded-lg text-xs">
              <TrendingUpIcon className="size-3" />
              +4.5%
            </Badge>
          </div>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          <div className="line-clamp-1 flex gap-2 font-medium">
            Steady performance <TrendingUpIcon className="size-4" />
          </div>
          <div className="text-muted-foreground">Meets growth projections</div>
        </CardFooter>
      </Card>
    </div>
  )
}
