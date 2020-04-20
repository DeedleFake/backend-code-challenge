package bcc

import (
	"errors"
	"fmt"
	"math"

	"github.com/jmoiron/sqlx"
)

// RateUser adds a rating to the ratings table.
func RateUser(db *sqlx.DB, raterID, userID uint64, rating float64) (err error) {
	if raterID == userID {
		return errors.New("not allowed to rate self")
	}
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

	var newRowID uint64
	err = tx.QueryRowx(`
		INSERT INTO ratings (rater_id, user_id, rating)
		VALUES ($1, $2, $3)
		RETURNING id
	`, raterID, userID, rating).Scan(&newRowID)
	if err != nil {
		return fmt.Errorf("scan new row: %w", err)
	}

	after, err := GetRating(tx, userID)
	if err != nil {
		return fmt.Errorf("after: %w", err)
	}

	if b, a := math.Floor(before), math.Floor(after); b != a {
		_, err = tx.Exec(`
			INSERT INTO rating_events (rating_id, rating_before, rating_after)
			VALUES ($1, $2, $3)
		`, newRowID, before, after)
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// GetRating gets the rating of a given user.
func GetRating(db sqlx.Queryer, userID uint64) (float64, error) {
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
