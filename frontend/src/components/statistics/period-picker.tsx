import { CalendarRangeIcon } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PERIOD_PRESETS, type PeriodPreset } from "@/lib/statistics-period";

interface PeriodPickerProps {
  preset: PeriodPreset;
  customFrom: string;
  customTo: string;
  onPresetChange: (preset: PeriodPreset) => void;
  onCustomChange: (from: string, to: string) => void;
}

export function PeriodPicker({
  preset,
  customFrom,
  customTo,
  onPresetChange,
  onCustomChange,
}: PeriodPickerProps) {
  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-end">
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="stats-period" className="text-xs text-muted-foreground">
          Zeitraum
        </Label>
        <Select value={preset} onValueChange={(v) => onPresetChange(v as PeriodPreset)}>
          <SelectTrigger id="stats-period" className="w-full sm:w-52" aria-label="Zeitraum wählen">
            <span className="flex items-center gap-2">
              <CalendarRangeIcon className="size-4 opacity-60" />
              <SelectValue />
            </span>
          </SelectTrigger>
          <SelectContent>
            {PERIOD_PRESETS.map((p) => (
              <SelectItem key={p.value} value={p.value}>
                {p.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      {preset === "custom" && (
        <div className="flex gap-2">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="stats-from" className="text-xs text-muted-foreground">
              Von
            </Label>
            <Input
              id="stats-from"
              type="date"
              value={customFrom}
              max={customTo || undefined}
              onChange={(e) => onCustomChange(e.target.value, customTo)}
              className="w-40"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="stats-to" className="text-xs text-muted-foreground">
              Bis
            </Label>
            <Input
              id="stats-to"
              type="date"
              value={customTo}
              min={customFrom || undefined}
              onChange={(e) => onCustomChange(customFrom, e.target.value)}
              className="w-40"
            />
          </div>
        </div>
      )}
    </div>
  );
}
