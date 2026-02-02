package gamedayimport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrEmptyFile    = errors.New("CSV-Datei ist leer")
	ErrInvalidHeader = errors.New("Header muss mindestens drei Spalten haben: Datum, Name und mindestens eine Wert-Spalte")
	ErrInvalidDate   = errors.New("Ungültiges Datumsformat (erwartet DD.MM.YYYY)")
	ErrInvalidValue  = errors.New("Ungültiger Zahlenwert")
)

const dateLayout = "02.01.2006"

// ParseFile reads and parses a semicolon-separated CSV file.
// First row is header: column 0 = Datum, 1 = Name, rest = penalty or competition column names.
// Data rows: date (DD.MM.YYYY), player name, then integer values per column.
// Empty cells are treated as 0.
func ParseFile(path string) (*ParsedCSV, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("CSV-Datei öffnen: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV lesen: %w", err)
	}

	if len(records) == 0 {
		return nil, ErrEmptyFile
	}

	header := records[0]
	for i, h := range header {
		// Strip UTF-8 BOM from first cell if present (e.g. Excel export)
		if i == 0 {
			h = strings.TrimPrefix(h, "\ufeff")
		}
		header[i] = strings.TrimSpace(h)
	}

	if len(header) < 3 {
		return nil, ErrInvalidHeader
	}

	if strings.TrimSpace(strings.ToLower(header[0])) != "datum" || strings.TrimSpace(strings.ToLower(header[1])) != "name" {
		return nil, fmt.Errorf("%w: erste Spalte muss 'Datum', zweite 'Name' heißen", ErrInvalidHeader)
	}

	rows := make([]CSVRow, 0, len(records)-1)
	valueColumns := header[2:]

	for i := 1; i < len(records); i++ {
		rec := records[i]
		// Pad with empty strings if row has fewer columns than header
		for len(rec) < len(header) {
			rec = append(rec, "")
		}

		dateStr := strings.TrimSpace(rec[0])
		if dateStr == "" {
			return nil, fmt.Errorf("Zeile %d: leeres Datum", i+1)
		}
		date, err := time.Parse(dateLayout, dateStr)
		if err != nil {
			return nil, fmt.Errorf("Zeile %d: %w (z.B. 30.01.2026)", i+1, ErrInvalidDate)
		}

		playerName := strings.TrimSpace(rec[1])
		if playerName == "" {
			return nil, fmt.Errorf("Zeile %d: leerer Spielername", i+1)
		}

		values := make(map[string]int, len(valueColumns))
		for j, colName := range valueColumns {
			raw := ""
			if j+2 < len(rec) {
				raw = strings.TrimSpace(rec[j+2])
			}
			if raw == "" {
				values[colName] = 0
				continue
			}
			v, err := strconv.Atoi(raw)
			if err != nil {
				return nil, fmt.Errorf("Zeile %d, Spalte %s: %w (Wert: %q)", i+1, colName, ErrInvalidValue, raw)
			}
			if v < 0 {
				return nil, fmt.Errorf("Zeile %d, Spalte %s: negative Werte sind nicht erlaubt", i+1, colName)
			}
			values[colName] = v
		}

		rows = append(rows, CSVRow{
			Date:       date,
			PlayerName: playerName,
			Values:     values,
		})
	}

	return &ParsedCSV{Header: header, Rows: rows}, nil
}
