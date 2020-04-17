package bcc

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RateUser adds a rating to the ratings table.
func RateUser(db *sqlx.DB, raterID, userID int, rating float64) (err error) {
	if (rating < 1) || (rating > 5) {
		return fmt.Errorf("invalid rating %v", rating)
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	before, err := GetRating(tx, userID)
	if err != nil {
		return fmt.Errorf("before: %w", err)
	}

	_, err = tx.Exec(`INSERT INTO ratings (rater_id, user_id, rating) VALUES ($1, $2, $3)`, raterID, userID, rating)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	after, err := GetRating(tx, userID)
	if err != nil {
		return fmt.Errorf("after: %w", err)
	}

	if (before < 4) && (after >= 4) {
		// TODO: Insert event for passing 4 stars.
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// GetRating gets the rating of a given user.
func GetRating(db sqlx.Queryer, userID int) (float64, error) {
	var avg *float64
	err := db.QueryRowx(`
		SELECT
			AVG(rating)
		FROM
			(
				SELECT
					ROW_NUMBER() OVER (PARTITION BY rater_id ORDER BY rated_at DESC) AS rn,
					rating
				FROM ratings
					WHERE user_id = $1
			 ) AS r
		WHERE rn=1
	`, userID).Scan(&avg)
	if avg != nil {
		return *avg, err
	}
	return 0, err
}
