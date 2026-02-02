package gamedayimport

import "time"

// ParsedCSV is the result of parsing a semicolon-separated CSV file.
// Header row: first column = Datum, second = Name, rest = penalty or competition column names.
type ParsedCSV struct {
	Header []string  // raw header row (trimmed)
	Rows   []CSVRow  // data rows
}

// CSVRow represents one data row: date, player name, and values per column name.
type CSVRow struct {
	Date       time.Time           // parsed date (DD.MM.YYYY)
	PlayerName string              // player name
	Values     map[string]float64  // column header name -> value (count or competition score; decimal allowed for penalty when type allows)
}

// ColumnKind indicates whether a CSV column is a penalty type or a competition.
type ColumnKind int

const (
	ColumnKindPenalty     ColumnKind = 1
	ColumnKindCompetition ColumnKind = 2
)

// ColumnMapping describes how a CSV column (by header name) maps to a penalty type or competition.
type ColumnMapping struct {
	ColumnName string
	Kind       ColumnKind
	// For penalty columns (snapshot for game_day_fees)
	PenaltyTypeID          string
	PenaltyTypeName        string
	PenaltyTypeDescription string
	PenaltyTypePrice       int
	AllowsDecimalQuantity  bool
	// For competition columns
	CompetitionID string
}

// ResolvedInput holds the result of resolving parsed CSV against club data.
// Used for both dry-run output and apply.
type ResolvedInput struct {
	ClubID        string
	ClubName     string
	ColumnMapping []ColumnMapping // one per column from header index 2 onward, in order
	ColumnNames  []string         // header names for value columns (same order as ColumnMapping)
	Rows         []CSVRow         // parsed rows with resolved dates
}
