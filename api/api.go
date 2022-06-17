package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/lggruspe/polycloze/flashcards"
	"github.com/lggruspe/polycloze/text"
	"github.com/lggruspe/polycloze/word_scheduler"
)

type Items struct {
	Items []flashcards.Item `json:"items"`
}

func getN(r *http.Request) int {
	n := 10
	q := r.URL.Query()
	if i, err := strconv.Atoi(q.Get("n")); err == nil && i > 0 {
		n = i
	}
	return n
}

// Returns predicate to pass to item generator.
func excludeWords(r *http.Request) func(string) bool {
	exclude := make(map[string]bool)
	for _, word := range r.URL.Query()["x"] {
		exclude[text.Casefold(word)] = true
	}
	return func(word string) bool {
		_, found := exclude[text.Casefold(word)]
		return !found
	}
}

func generateFlashcards(ig *flashcards.ItemGenerator, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	words, err := ig.GenerateWordsWith(getN(r), excludeWords(r))
	if err != nil {
		log.Fatal(err)
	}

	items := ig.GenerateItems(words)
	bytes, err := json.Marshal(Items{Items: items})
	if err != nil {
		log.Fatal(err)
	}
	w.Write(bytes)
}

func success() []byte {
	return []byte("{\"success\": true}")
}

func handleReviewUpdate(ig *flashcards.ItemGenerator, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		panic("expected json body in POST request")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic("something went wrong")
	}

	var reviews Reviews
	if err := json.Unmarshal(body, &reviews); err != nil {
		panic("parsing error")
	}

	session, err := ig.Session()
	if err != nil {
		panic("something went wrong")
	}
	defer session.Close()

	for _, review := range reviews.Reviews {
		word_scheduler.UpdateWord(session, review.Word, review.Correct)
	}
	w.Write(success())
}

// Middleware
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
		next.ServeHTTP(w, r)
	})
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	l1 := chi.URLParam(r, "l1")
	l2 := chi.URLParam(r, "l2")

	ig, err := loadLanguagePair(l1, l2)
	if err != nil {
		log.Fatal(err)
	}

	switch r.Method {
	case "POST":
		handleReviewUpdate(ig, w, r)
	case "GET":
		generateFlashcards(ig, w, r)
	}
}

func Router(config Config) (chi.Router, error) {
	r := chi.NewRouter()
	if config.AllowCORS {
		r.Use(cors)
	}
	r.Use(middleware.Logger)
	r.HandleFunc("/options", languageOptions)
	r.HandleFunc("/{l1}/{l2}", createHandler)
	return r, nil
}
