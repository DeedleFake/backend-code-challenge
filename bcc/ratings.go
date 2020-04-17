package bcc

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func RateUser(db *sqlx.DB, raterID, userID int, rating float64) error {
	if (rating < 1) || (rating > 5) {
		return fmt.Errorf("invalid rating %v", rating)
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	avgStmt, err := tx.Preparex(`SELECT AVG(rating) FROM ratings WHERE user_id=$1`)
	if err != nil {
		return fmt.Errorf("prepare average: %w", err)
	}

	var before *float64
	err = avgStmt.QueryRowx(userID).Scan(&before)
	if err != nil {
		return fmt.Errorf("scan before: %w", err)
	}

	update := `UPDATE ratings SET rating=$1, updated_at=CURRENT_TIMESTAMP WHERE rater_id=$2 AND user_id=$3`
	if before == nil {
		update = `INSERT INTO ratings (rater_id, user_id, rating) VALUES ($2, $3, $1)`
	}
	_, err = tx.Exec(update, rating, raterID, userID)
	if err != nil {
		return fmt.Errorf("upsert: %w", err)
	}

	var after *float64
	err = avgStmt.QueryRowx(userID).Scan(&after)
	if err != nil {
		return fmt.Errorf("scan after: %w", err)
	}

	// TODO: Insert event for passing 4 stars.

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
