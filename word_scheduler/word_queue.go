// Copyright (c) 2022 Levi Gruspe
// License: GNU AGPLv3 or later

// Returns words that the user hasn't learned.
package word_scheduler

import (
	"database/sql"

	"github.com/lggruspe/polycloze/database"
)

// NOTE Does not close Rows.
func getNRows(rows *sql.Rows, n int, pred func(word string) bool) ([]string, error) {
	var words []string
	for rows.Next() && len(words) < n {
		var word string
		if err := rows.Scan(&word); err != nil {
			return nil, err
		}
		if pred(word) {
			words = append(words, word)
		}
	}
	return words, nil
}

func getWordsAboveDifficultyWith[T database.Querier](q T, n, preferredDifficulty int, pred func(word string) bool) ([]string, error) {
	query := `
select word from word where frequency_class >= ? and word not in
(select item from review)
order by id asc
`
	rows, err := q.Query(query, preferredDifficulty, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getNRows(rows, n, pred)
}

func getWordsBelowDifficultyWith[T database.Querier](q T, n, preferredDifficulty int, pred func(word string) bool) ([]string, error) {
	query := `
select word from word where frequency_class < ? and word not in
(select item from review)
order by id desc
`
	rows, err := q.Query(query, preferredDifficulty, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return getNRows(rows, n, pred)
}

// Gets up to n new words from db.
// Pass a negative n if you don't want a word limit.
// Uses preferredDifficulty as minimum word frequency class.
// If there are not enough words in query result, will also include words below
// the preferredDifficulty.
// Only words that satisfy the predicate are included in the result.
func GetNewWordsWith[T database.Querier](q T, n, preferredDifficulty int, pred func(word string) bool) ([]string, error) {
	words, err := getWordsAboveDifficultyWith(q, n, preferredDifficulty, pred)
	if err != nil {
		return nil, err
	}
	if preferredDifficulty <= 0 || len(words) >= n {
		return words, nil
	}

	more, err := getWordsBelowDifficultyWith(q, n-len(words), preferredDifficulty, pred)
	if err != nil {
		return nil, err
	}
	words = append(words, more...)
	return words, nil
}
