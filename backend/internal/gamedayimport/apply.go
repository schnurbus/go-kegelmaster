package gamedayimport

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
)

// dateOnly returns t normalized to midnight UTC for comparison with DB date values.
func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Apply applies the resolved CSV data to the database (or prints planned actions if dryRun).
// Idempotent: get-or-create game day, get-or-create participant, upsert fees and competition values.
func Apply(ctx context.Context, resolved *ResolvedInput, playerByName map[string]string, deps *Dependencies, dryRun bool) error {
	if dryRun {
		return applyDryRun(ctx, resolved, playerByName, deps)
	}
	return applyWrite(ctx, resolved, playerByName, deps)
}

func applyDryRun(ctx context.Context, resolved *ResolvedInput, playerByName map[string]string, deps *Dependencies) error {
	datesSeen := make(map[time.Time]struct{})
	for _, row := range resolved.Rows {
		d := dateOnly(row.Date)
		datesSeen[d] = struct{}{}
	}

	for d := range datesSeen {
		exists, err := deps.GameDayRepo.CheckExistsByClubAndDate(ctx, resolved.ClubID, d)
		if err != nil {
			return fmt.Errorf("Spieltag prüfen: %w", err)
		}
		if exists {
			slog.Info("dry-run: Spieltag existiert bereits", "datum", d.Format("02.01.2006"), "club", resolved.ClubName)
		} else {
			slog.Info("dry-run: würde Spieltag anlegen", "datum", d.Format("02.01.2006"), "club", resolved.ClubName)
		}
	}

	total := len(resolved.Rows)
	slog.Info("dry-run: Verarbeitung geplant", "zeilen", total)
	for i, row := range resolved.Rows {
		slog.Info("dry-run: Zeile", "fortschritt", fmt.Sprintf("%d/%d", i+1, total), "datum", row.Date.Format("02.01.2006"), "spieler", row.PlayerName)
		for _, col := range resolved.ColumnMapping {
			if col.Kind == ColumnKindPenalty {
				v := row.Values[col.ColumnName]
				slog.Info("dry-run: würde Strafe setzen", "spieler", row.PlayerName, "strafentyp", col.PenaltyTypeName, "anzahl", v)
			} else {
				v := row.Values[col.ColumnName]
				slog.Info("dry-run: würde Wettbewerbswert setzen", "spieler", row.PlayerName, "competition_id", col.CompetitionID, "wert", v)
			}
		}
	}

	slog.Info("dry-run: keine Änderungen an der Datenbank")
	return nil
}

func applyWrite(ctx context.Context, resolved *ResolvedInput, playerByName map[string]string, deps *Dependencies) error {
	clubEntity, err := deps.ClubRepo.GetByID(ctx, resolved.ClubID)
	if err != nil {
		if err == club.ErrNotFound {
			return fmt.Errorf("Club nicht gefunden: %s", resolved.ClubID)
		}
		return fmt.Errorf("Club laden: %w", err)
	}

	// Collect unique dates and get-or-create game days; track is_draft per date for transaction creation
	gameDayByDate := make(map[time.Time]string)   // date (UTC midnight) -> game day ID
	isDraftByDate := make(map[time.Time]bool)    // date -> is_draft (no transactions when true)
	datesOrder := make([]time.Time, 0)
	datesSeen := make(map[time.Time]struct{})

	for _, row := range resolved.Rows {
		d := dateOnly(row.Date)
		if _, ok := datesSeen[d]; ok {
			continue
		}
		datesSeen[d] = struct{}{}
		datesOrder = append(datesOrder, d)
	}

	for _, d := range datesOrder {
		exists, err := deps.GameDayRepo.CheckExistsByClubAndDate(ctx, resolved.ClubID, d)
		if err != nil {
			return fmt.Errorf("Spieltag prüfen: %w", err)
		}

		var gameDayID string
		var isDraft bool
		if exists {
			gameDays, err := deps.GameDayRepo.GetByClubID(ctx, resolved.ClubID)
			if err != nil {
				return fmt.Errorf("Spieltage laden: %w", err)
			}
			for _, gd := range gameDays {
				if dateOnly(gd.Date).Equal(d) {
					gameDayID = gd.ID
					isDraft = gd.IsDraft
					break
				}
			}
			if gameDayID == "" {
				return fmt.Errorf("Spieltag für Datum %s nicht gefunden", d.Format("02.01.2006"))
			}
		} else {
			created, err := deps.GameDayRepo.Create(ctx, gameday.CreateGameDayParams{
				ClubID:  resolved.ClubID,
				Date:    d,
				Notes:   "",
				IsDraft: true, // import creates draft gamedays; no base fee transactions
			})
			if err != nil {
				return fmt.Errorf("Spieltag anlegen: %w", err)
			}
			gameDayID = created.ID
			isDraft = created.IsDraft

			// Base fee transactions only when created game day is not draft
			if !isDraft && clubEntity.BaseFee > 0 {
				players, err := deps.PlayerRepo.GetByClubID(ctx, resolved.ClubID)
				if err != nil {
					slog.Error("Spieler für Grundgebühr laden", "error", err)
				} else {
					for _, p := range players {
						if p.Inactive || p.RoleID == nil {
							continue
						}
						playerRole, err := deps.RoleRepo.GetByID(ctx, *p.RoleID)
						if err != nil || !playerRole.PaysBaseFee {
							continue
						}
						playerID := p.ID
						_, err = deps.TransactionRepo.Create(ctx, transaction.CreateTransactionParams{
							ClubID:           resolved.ClubID,
							PlayerID:         &playerID,
							TransactionType:  transaction.TransactionTypeBaseFee,
							Amount:           -clubEntity.BaseFee,
							Description:      fmt.Sprintf("Grundgebühr für %s", d.Format("02.01.2006")),
							GameDayID:        &gameDayID,
							TransactionDate:  d,
						})
						if err != nil {
							slog.Error("Grundgebühr-Transaktion anlegen", "player", p.ID, "error", err)
						}
					}
				}
			}
		}
		gameDayByDate[d] = gameDayID
		isDraftByDate[d] = isDraft
	}

	// Process each row: get-or-create participant, upsert fees and competition values
	total := len(resolved.Rows)
	slog.Info("Import starten", "zeilen", total)
	for i, row := range resolved.Rows {
		slog.Info("Import: Zeile", "fortschritt", fmt.Sprintf("%d/%d", i+1, total), "datum", row.Date.Format("02.01.2006"), "spieler", row.PlayerName)
		d := dateOnly(row.Date)
		gameDayID := gameDayByDate[d]
		rowIsDraft := isDraftByDate[d]
		playerID := playerByName[row.PlayerName]

		participant, err := deps.GameDayRepo.GetParticipantByGameDayAndPlayer(ctx, gameDayID, playerID)
		if err != nil {
			if err != gameday.ErrParticipantNotFound {
				return fmt.Errorf("Teilnehmer laden: %w", err)
			}
			participant, err = deps.GameDayRepo.AddParticipant(ctx, gameDayID, playerID)
			if err != nil {
				return fmt.Errorf("Teilnehmer anlegen: %w", err)
			}
		}

		existingFees, err := deps.GameDayRepo.GetFeesByParticipant(ctx, participant.ID)
		if err != nil {
			return fmt.Errorf("Fees laden: %w", err)
		}
		existingFeeMap := make(map[string]gameday.GameDayFee)
		for _, f := range existingFees {
			existingFeeMap[f.PenaltyTypeID] = f
		}

		for _, col := range resolved.ColumnMapping {
			if col.Kind == ColumnKindPenalty {
				countRaw := row.Values[col.ColumnName]
				if countRaw <= 0 || math.IsNaN(countRaw) {
					if !rowIsDraft {
						if existingFee, exists := existingFeeMap[col.PenaltyTypeID]; exists {
							existingTx, err := deps.TransactionRepo.GetByGameDayFee(ctx, existingFee.ID)
							if err == nil && existingTx != nil {
								_ = deps.TransactionRepo.DeleteFeeTransaction(ctx, existingTx.ID)
							}
						}
					}
					_ = deps.GameDayRepo.DeleteFee(ctx, participant.ID, col.PenaltyTypeID)
					continue
				}

				quantityScale := 1
				if col.AllowsDecimalQuantity {
					quantityScale = 100
				}
				countStored := int(math.Round(countRaw * float64(quantityScale)))
				if countStored <= 0 {
					continue
				}

				existingFee, isUpdate := existingFeeMap[col.PenaltyTypeID]
				if !rowIsDraft && isUpdate {
					existingTx, err := deps.TransactionRepo.GetByGameDayFee(ctx, existingFee.ID)
					if err == nil && existingTx != nil {
						_ = deps.TransactionRepo.DeleteFeeTransaction(ctx, existingTx.ID)
					}
				}

				createdFee, err := deps.GameDayRepo.UpsertFee(ctx, gameday.UpsertFeeParams{
					ParticipantID:          participant.ID,
					PenaltyTypeID:          col.PenaltyTypeID,
					PenaltyTypeName:        col.PenaltyTypeName,
					PenaltyTypeDescription: col.PenaltyTypeDescription,
					PenaltyTypePrice:       col.PenaltyTypePrice,
					Count:                  countStored,
					QuantityScale:          quantityScale,
				})
				if err != nil {
					return fmt.Errorf("Fee upsert: %w", err)
				}

				if !rowIsDraft {
					amount := -(col.PenaltyTypePrice * countStored / quantityScale)
					quantityDesc := fmt.Sprintf("%d", countStored)
					if quantityScale > 1 {
						quantityDesc = fmt.Sprintf("%.2f", float64(countStored)/float64(quantityScale))
					}
					_, err = deps.TransactionRepo.Create(ctx, transaction.CreateTransactionParams{
						ClubID:           resolved.ClubID,
						PlayerID:         &playerID,
						TransactionType:  transaction.TransactionTypeFee,
						Amount:           amount,
						Description:      fmt.Sprintf("%s ×%s", col.PenaltyTypeName, quantityDesc),
						GameDayFeeID:     &createdFee.ID,
						GameDayID:        &gameDayID,
						TransactionDate:  dateOnly(row.Date),
					})
					if err != nil {
						slog.Error("Fee-Transaktion anlegen", "error", err)
					}
				}
			} else {
				value := int(math.Round(row.Values[col.ColumnName]))
				_, err = deps.GameDayRepo.UpsertCompetitionValue(ctx, gameday.UpsertCompetitionValueParams{
					ParticipantID: participant.ID,
					CompetitionID: col.CompetitionID,
					Value:         value,
				})
				if err != nil {
					return fmt.Errorf("Wettbewerbswert upsert: %w", err)
				}
			}
		}
	}

	return nil
}
