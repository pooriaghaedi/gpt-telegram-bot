package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func dbInit() {
	db, err := sql.Open("sqlite3", "data/users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table
	statement, err := db.Prepare(
		"CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, tg_id INTEGER, username TEXT, is_valid BOOLEAN, credits INTEGER, created_at TIMESTAMP, prompt_id INTEGER, FOREIGN KEY (prompt_id) REFERENCES prompts(id));",
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}

	// Create the prompts table
	statement, err = db.Prepare(
		"CREATE TABLE IF NOT EXISTS prompts (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, text TEXT NOT NULL);",
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func addUser(db *sql.DB, tg_id int, username string, is_valid bool, credits int) error {
	// Start a new transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Check if the user already exists
	var id int
	err = tx.QueryRow("SELECT id FROM users WHERE tg_id = ?", tg_id).Scan(&id)
	fmt.Println("USER_ID: " + string(id))

	if err != nil && err != sql.ErrNoRows {
		// An error occurred while querying for the user
		tx.Rollback()
		return err
	} else if err == sql.ErrNoRows {
		// The user does not exist, insert a new user
		_, err = tx.Exec("INSERT INTO users (tg_id, username, is_valid, credits, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)", tg_id, username, is_valid, credits)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// The user exists, update is_valid
		_, err = tx.Exec("UPDATE users SET is_valid = ? WHERE id = ?", is_valid, id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func getUsers(db *sql.DB) []User {
	rows, _ := db.Query("SELECT id, tg_id, username, is_valid, credits, created_at FROM users")
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		rows.Scan(&user.ID, &user.TelegramId, &user.Username, &user.IsValid, &user.Credits, &user.CreatedAt)
		users = append(users, user)
	}

	return users
}

func isUserValid(db *sql.DB, tg_id int) bool {
	var isValid bool
	err := db.QueryRow("SELECT is_valid FROM users WHERE tg_id = ?", tg_id).Scan(&isValid)
	if err != nil {
		return false
	}
	return isValid
}

func processSecret(db *sql.DB, receivedMessage string, tg_id int, userName string) string {
	passwd := os.Getenv("PASSWD")
	var responseMessage string
	if receivedMessage == passwd {
		err := addUser(db, tg_id, userName, true, 10)
		handleErr(err)
		responseMessage = "Access Granted, Now enjoy using ChatGPT :*"
	} else {
		err := addUser(db, tg_id, userName, false, 0)
		handleErr(err)
		responseMessage = "Enter Password:"
		fmt.Println("denied " + userName)
	}

	return responseMessage
}
