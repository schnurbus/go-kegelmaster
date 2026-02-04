import { useEffect, useState } from "react"

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
    <>
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
          <CardDescription>Club-Guthaben</CardDescription>
          <CardTitle className="@[250px]/card:text-3xl text-2xl font-semibold tabular-nums">
            {activeClub
              ? formatCentsToEuro(activeClub.balance)
              : "Kein Club ausgewählt"}
          </CardTitle>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1 text-sm">
          {activeClub ? (
            <div className="text-muted-foreground">
              Guthaben von {activeClub.name}
            </div>
          ) : (
            <div className="text-muted-foreground">
              Bitte wählen Sie einen Club.
            </div>
          )}
        </CardFooter>
      </Card>
    </>
  )
}
