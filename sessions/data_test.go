// Copyright (c) 2022 Levi Gruspe
// License: GNU AGPLv3 or later

package sessions

import (
	"testing"
)

func TestGetDataNonExistentID(t *testing.T) {
	// The result should be a non-nil empty map.
	t.Parallel()
	id := "abcdefg"
	db := testDB()
	defer db.Close()

	data := getData(db, id)
	if data == nil {
		t.Fatal("expected non-nil result:", data)
	}

	if len(data) != 0 {
		t.Fatal("expected result to be an empty map:", data)
	}
}

func TestSaveDataNonExistentID(t *testing.T) {
	// Allowed, but getData should be empty.
	t.Parallel()
	id := "abcdefg"
	db := testDB()
	defer db.Close()

	s := Session{
		ID: id,
		Data: map[string]any{
			"userID":   123,
			"username": "foobar",
		},
	}
	if err := saveData(db, &s); err != nil {
		t.Fatal("expected err to be nil:", err)
	}

	data := getData(db, id)
	if len(data) != 0 {
		t.Fatal("expected the result to be an empty map:", data)
	}
}

func TestGetSaveNonEmptyData(t *testing.T) {
	// Result should contain session data.
	t.Parallel()
	id := "abcdefg"
	db := testDB()
	defer db.Close()

	s := Session{
		ID: id,
		Data: map[string]any{
			"userID":   123,
			"username": "foobar",
		},
	}

	if err := reserveID(db, id); err != nil {
		t.Fatal("expected err to be nil:", err)
	}
	if err := saveData(db, &s); err != nil {
		t.Fatal("expected err to be nil:", err)
	}

	data := getData(db, id)
	if data["username"] != s.Data["username"] {
		t.Fatal(
			"expected usernames to be equal:",
			data["username"],
			s.Data["username"],
		)
	}
	if data["userID"] != s.Data["userID"] {
		t.Fatal(
			"expected userIDs to be equal:",
			data["userID"],
			s.Data["userID"],
		)
	}
	if len(data) != 2 {
		t.Fatal("expected data to not contain extraneous entries:", data)
	}
}

func TestGetSaveEmptyData(t *testing.T) {
	// If map is empty, DB entries should be set to null.
	t.Parallel()
	id := "abcdefg"
	db := testDB()
	defer db.Close()

	// First, insert data into DB.
	s := Session{
		ID: id,
		Data: map[string]any{
			"userID":   123,
			"username": "foobar",
		},
	}

	if err := reserveID(db, id); err != nil {
		t.Fatal("expected err to be nil:", err)
	}
	if err := saveData(db, &s); err != nil {
		t.Fatal("expected err to be nil:", err)
	}
	if data := getData(db, id); len(data) != 2 {
		t.Fatal("expected result to contain two entries:", data)
	}

	// Next, save session with empty data.
	s.Data = make(map[string]any)

	if err := saveData(db, &s); err != nil {
		t.Fatal("expected err to be nil:", err)
	}

	if data := getData(db, id); len(data) != 0 {
		t.Fatal("expected result to be an empty map:", data)
	}
}
