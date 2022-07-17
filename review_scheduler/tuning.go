// Auto-tuning stuff (of spacing algorithm parameters).
package review_scheduler

import (
	"database/sql"
	"time"
)

const day time.Duration = 24 * time.Hour

func isTooEasy(correct, incorrect int) bool {
	// Threshold can't be too high or the tuner will be too conservative.
	// Only uses 0.80 confidence, higher values require too many samples.

	z := -0.845 // z-score for one-sided confidence interval (80% confidence)
	lower := Wilson(correct, incorrect, z)

	// 80% likelihood that the true proportion is bounded below by `lower`.
	// > 0.9 test is too slow when incorrect > 0
	return lower > 0.875
}

func isTooHard(correct, incorrect int) bool {
	z := 2.325 // z-score for one-sided confidence interval
	upper := Wilson(correct, incorrect, z)

	// 99% confident that the true proportion is bounded above by `upper`
	return upper < 0.8
}

func tuneDifficulty(tx *sql.Tx) error {
	query := `select correct, incorrect from student`
	row := tx.QueryRow(query)

	var correct, incorrect int
	if err := row.Scan(&correct, &incorrect); err != nil {
		return err
	}

	if isTooHard(correct, incorrect) {
		return decreaseDifficulty(tx)
	} else if isTooEasy(correct, incorrect) {
		return increaseDifficulty(tx)
	}
	return nil
}

// Auto-tunes intervals and base difficulty (student.frequency_class).
func autoTune(tx *sql.Tx) error {
	if err := tuneDifficulty(tx); err != nil {
		return err
	}

	query := `select interval, correct, incorrect from interval order by interval asc`
	rows, err := tx.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var interval_s time.Duration
		var correct, incorrect int
		if err := rows.Scan(&interval_s, &correct, &incorrect); err != nil {
			return err
		}

		interval := interval_s * time.Second

		if interval <= day {
			// Don't change intervals = 0 and 1 day.
			continue
		}

		if isTooHard(correct, incorrect) {
			if err := decreaseInterval(tx, interval); err != nil {
				return err
			}
		} else if isTooEasy(correct, incorrect) {
			if err := increaseInterval(tx, interval); err != nil {
				return err
			}
		}
	}
	return nil
}

// Returns biggest interval smaller than the specified value.
func previousInterval(tx *sql.Tx, interval time.Duration) (time.Duration, error) {
	if interval <= day {
		return 0, nil
	}
	query := `select max(interval) from interval where interval < ?`
	row := tx.QueryRow(query, seconds(interval))

	var prev time.Duration
	if err := row.Scan(&prev); err != nil {
		return 0, err
	}
	// NOTE Assumes the query never returns null.
	return prev * time.Second, nil
}

// Check if interval already exists in db.
func alreadyExists(tx *sql.Tx, interval time.Duration) (bool, error) {
	query := `select * from interval where interval = ?`
	rows, err := tx.Query(query, seconds(interval))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		return true, nil
	}
	return false, nil
}

// Assumes replacement isn't already in the interval table.
func replaceInterval(tx *sql.Tx, interval, replacement time.Duration) error {
	interval_s := seconds(interval)
	replacement_s := seconds(replacement)

	query := `
update interval set interval = ?, correct = 0, incorrect = 0
where interval = ?
`
	if _, err := tx.Exec(query, replacement_s, interval_s); err != nil {
		return err
	}

	query = `
update review set
	interval = ?,
	due = datetime(unixepoch(due) + (? - interval) / 1e9, 'unixepoch')
where interval = ?
`
	_, err := tx.Exec(query, replacement_s, replacement_s, interval_s)
	return err
}

func replaceWithExistingInterval(tx *sql.Tx, interval, replacement time.Duration) error {
	interval_s := seconds(interval)
	replacement_s := seconds(interval)

	query := `delete from interval where interval = ?`
	if _, err := tx.Exec(query, interval_s); err != nil {
		return err
	}

	query = `update review set interval = ? where interval = ?`
	_, err := tx.Exec(query, replacement_s, interval_s)
	return err
}

func decreaseInterval(tx *sql.Tx, interval time.Duration) error {
	if interval <= day {
		return nil
	}

	prev, err := previousInterval(tx, interval)
	if err != nil {
		return err
	}
	mid := (prev + interval) / 2

	if exists, err := alreadyExists(tx, mid); err != nil {
		return err
	} else if exists {
		return replaceWithExistingInterval(tx, interval, mid)
	}
	return replaceInterval(tx, interval, mid)
}

// Returns the largest interval in the database.
func maxInterval(tx *sql.Tx) (time.Duration, error) {
	query := `select max(interval) from interval`
	row := tx.QueryRow(query)

	var max time.Duration
	if err := row.Scan(&max); err != nil {
		return 0, err
	}
	return max * time.Second, nil
}

// Creates record for interval if it doesn't already exist.
func insertInterval(tx *sql.Tx, interval time.Duration) error {
	query := `insert or ignore into interval (interval) values (?)`
	_, err := tx.Exec(query, seconds(interval))
	return err
}

// Inserts all needed intervals to increase the specified interval.
func insertMissingIntervals(tx *sql.Tx, interval time.Duration) error {
	max, err := maxInterval(tx)
	if err != nil {
		return err
	}

	if max > interval {
		return nil
	}

	next := 2 * max
	if next <= 0 {
		next = day
	}

	for next <= interval {
		if err := insertInterval(tx, next); err != nil {
			return err
		}
		next *= 2
	}
	// Make sure that a larger interval exists
	return insertInterval(tx, next)
}

// Returns smallest interval bigger than the specified value.
func nextInterval(tx *sql.Tx, interval time.Duration) (time.Duration, error) {
	if err := insertMissingIntervals(tx, interval); err != nil {
		return 0, err
	}

	query := `select min(interval) from interval where interval > ?`
	row := tx.QueryRow(query, seconds(interval))

	var next time.Duration
	if err := row.Scan(&next); err != nil {
		return 0, err
	}
	return next * time.Second, nil
}

func increaseInterval(tx *sql.Tx, interval time.Duration) error {
	next, err := nextInterval(tx, interval)
	if err != nil {
		return err
	}
	mid := (interval + next) / 2

	if exists, err := alreadyExists(tx, mid); err != nil {
		return err
	} else if exists {
		return replaceWithExistingInterval(tx, interval, mid)
	}
	return replaceInterval(tx, interval, mid)
}

// Updates interval and student tables.
func updateIntervalStats(tx *sql.Tx, review *Review, correct bool) error {
	var interval time.Duration = 0
	if review != nil {
		interval = review.Interval
	}

	// Update interval
	query := `update interval set correct = correct + 1 where interval = ?`
	if !correct {
		query = `update interval set incorrect = incorrect + 1 where interval = ?`
	}
	_, err := tx.Exec(query, seconds(interval))
	return err
}

func increaseDifficulty(tx *sql.Tx) error {
	query := `
update student set
	frequency_class = frequency_class + 1,
	correct = 0,
	incorrect = 0
`
	_, err := tx.Exec(query)
	return err
}

func decreaseDifficulty(tx *sql.Tx) error {
	query := `
update student set
	frequency_class = max(0, frequency_class - 1),
	correct = 0,
	incorrect = 0
`
	_, err := tx.Exec(query)
	return err
}
