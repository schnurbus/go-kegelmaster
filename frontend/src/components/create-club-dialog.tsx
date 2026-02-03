"use client"

import * as React from "react"
import { toast } from "sonner"
import { useAuth } from "@/context/AuthContext"
import { useClub } from "@/context/ClubContext"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Checkbox } from "@/components/ui/checkbox"

type Club = {
  id: string
  name: string
  balance: number
  base_fee: number
  auto_tip_enabled: boolean
  couples_mode_enabled: boolean
  user_id: string
  created_at: string
  updated_at: string
}

type CreateClubDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onClubCreated: (club: Club) => void
}

export function CreateClubDialog({
  open,
  onOpenChange,
  onClubCreated,
}: CreateClubDialogProps) {
  const { csrfToken, refreshCsrf } = useAuth()
  const { refreshClubs } = useClub()
  const [name, setName] = React.useState("")
  const [balance, setBalance] = React.useState("")
  const [baseFee, setBaseFee] = React.useState("")
  const [autoTipEnabled, setAutoTipEnabled] = React.useState(true)
  const [couplesModeEnabled, setCouplesModeEnabled] = React.useState(false)
  const [isSubmitting, setIsSubmitting] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)

    try {
      const token = csrfToken || (await refreshCsrf())
      const balanceCents = Math.round(parseFloat(balance || "0") * 100)
      const baseFeeCents = Math.round(parseFloat(baseFee || "0") * 100)

      if (name.trim() === "") {
        setError("Name ist erforderlich")
        setIsSubmitting(false)
        return
      }

      if (balanceCents < 0) {
        setError("Balance darf nicht negativ sein")
        setIsSubmitting(false)
        return
      }

      if (baseFeeCents < 0) {
        setError("BaseFee darf nicht negativ sein")
        setIsSubmitting(false)
        return
      }

      const resp = await fetch("/api/clubs", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify({
          name: name.trim(),
          balance: balanceCents,
          base_fee: baseFeeCents,
          auto_tip_enabled: autoTipEnabled,
          couples_mode_enabled: couplesModeEnabled,
        }),
      })

      if (!resp.ok) {
        const errorText = await resp.text()
        try {
          const errorData = JSON.parse(errorText)
          throw new Error(errorData.message || errorData.error || "Fehler beim Erstellen des Clubs")
        } catch {
          throw new Error(errorText || "Fehler beim Erstellen des Clubs")
        }
      }

      const newClub: Club = await resp.json()
      await refreshClubs()
      onClubCreated(newClub)
      setName("")
      setBalance("")
      setBaseFee("")
      setAutoTipEnabled(true)
      setCouplesModeEnabled(false)
      onOpenChange(false)
      toast.success("Club erstellt")
    } catch (err) {
      const message = (err as Error).message
      setError(message)
      toast.error(message)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Neuen Club erstellen</DialogTitle>
          <DialogDescription>
            Erstelle einen neuen Club. Du wirst automatisch als Owner gesetzt.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Club Name"
                required
                disabled={isSubmitting}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="balance">Startguthaben (€)</Label>
              <Input
                id="balance"
                type="number"
                step="0.01"
                min="0"
                value={balance}
                onChange={(e) => setBalance(e.target.value)}
                placeholder="0.00"
                disabled={isSubmitting}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="baseFee">Basisgebühr (€)</Label>
              <Input
                id="baseFee"
                type="number"
                step="0.01"
                min="0"
                value={baseFee}
                onChange={(e) => setBaseFee(e.target.value)}
                placeholder="0.00"
                disabled={isSubmitting}
              />
            </div>
            <div className="flex items-center space-x-2">
              <Checkbox
                id="autoTipEnabled"
                checked={autoTipEnabled}
                onCheckedChange={(checked) => setAutoTipEnabled(checked === true)}
                disabled={isSubmitting}
              />
              <Label
                htmlFor="autoTipEnabled"
                className="text-sm font-normal cursor-pointer"
              >
                Auto-Tip aktivieren
              </Label>
            </div>
            <p className="text-sm text-muted-foreground -mt-2 ml-6">
              Überschüssige Einzahlungen werden automatisch als Trinkgeld verbucht.
              Spieler können kein positives Guthaben haben.
            </p>
            <div className="flex items-center space-x-2">
              <Checkbox
                id="couplesModeEnabled"
                checked={couplesModeEnabled}
                onCheckedChange={(checked) => setCouplesModeEnabled(checked === true)}
                disabled={isSubmitting}
              />
              <Label
                htmlFor="couplesModeEnabled"
                className="text-sm font-normal cursor-pointer"
              >
                Paar-Modus aktivieren
              </Label>
            </div>
            <p className="text-sm text-muted-foreground -mt-2 ml-6">
              Bei Einzahlungen können mehrere Spieler ausgewählt werden; der Betrag wird gleichmäßig verteilt.
            </p>
            {error && (
              <div className="text-sm text-destructive">{error}</div>
            )}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Abbrechen
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Erstelle..." : "Erstellen"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

