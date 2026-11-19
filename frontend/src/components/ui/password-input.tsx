import * as React from "react"
import { Eye, EyeOff } from "lucide-react"

import { cn } from "@/lib/utils"
import { Input } from "@/components/ui/input"

type PasswordInputProps = Omit<React.ComponentProps<"input">, "type">

/**
 * Passwortfeld mit Augen-Button rechts im Feld, um die Eingabe anzuzeigen
 * bzw. wieder zu verbergen. Standardmäßig maskiert.
 *
 * Der Button-Name kommt bewusst aus einem sr-only-Text statt aus `aria-label`,
 * damit `getByLabel("Passwort")` (Playwright) weiterhin eindeutig das Feld findet.
 */
const PasswordInput = React.forwardRef<HTMLInputElement, PasswordInputProps>(
  function PasswordInput({ className, disabled, ...props }, ref) {
    const [visible, setVisible] = React.useState(false)
    const label = visible ? "Passwort verbergen" : "Passwort anzeigen"
    const Icon = visible ? EyeOff : Eye

    return (
      <div className="relative">
        <Input
          ref={ref}
          type={visible ? "text" : "password"}
          disabled={disabled}
          className={cn("pr-10", className)}
          {...props}
        />
        <button
          type="button"
          onClick={() => setVisible((v) => !v)}
          disabled={disabled}
          aria-pressed={visible}
          title={label}
          className="text-muted-foreground hover:text-foreground focus-visible:ring-ring/50 absolute inset-y-0 right-0 flex w-9 items-center justify-center rounded-r-md outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50"
        >
          <Icon className="size-4" aria-hidden="true" />
          <span className="sr-only">{label}</span>
        </button>
      </div>
    )
  }
)

export { PasswordInput }
