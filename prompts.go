package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func getPromptID(db *sql.DB, promptName string) (int, error) {
	var promptID int
	err := db.QueryRow("SELECT id FROM prompts WHERE name = ?", promptName).Scan(&promptID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows found, not an error in this context
			return 0, nil
		}
		// An actual error occurred
		return 0, err
	}
	return promptID, nil
}

func createNewPrompt(db *sql.DB, promptStr string) error {
	// Split the input string into name and text
	parts := strings.SplitN(promptStr, " - ", 2)
	if len(parts) != 2 {
		return errors.New("invalid prompt format")
	}
	promptName := parts[0]
	promptText := parts[1]

	// Prepare SQL statement
	statement, err := db.Prepare("INSERT INTO prompts (name, text) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing insert into prompts: %v", err)
	}
	defer statement.Close()

	// Execute SQL statement
	_, err = statement.Exec(promptName, promptText)
	if err != nil {
		return fmt.Errorf("error executing insert into prompts: %v", err)
	}

	return nil
}

func updateUserPrompt(db *sql.DB, userID, promptID int) error {
	statement, err := db.Prepare("UPDATE users SET prompt_id = ? WHERE tg_id = ?")
	fmt.Println("userID: ", userID, "promptID: ", promptID)
	if err != nil {
		return fmt.Errorf("error preparing update users: %v", err)
	}
	defer statement.Close()

	_, err = statement.Exec(promptID, userID)
	if err != nil {
		return fmt.Errorf("error executing update users: %v", err)
	}
	return nil
}

func getPrompt(db *sql.DB, tgID int) string {
	var promptText string
	err := db.QueryRow(`
		SELECT p.text 
		FROM prompts p 
		JOIN users u ON u.prompt_id = p.id 
		WHERE u.tg_id = ?`, tgID).Scan(&promptText)
	if err != nil {
		// Handle the error according to your application's needs, like logging or returning an error message
		return ""
	}
	return promptText
}

func getAllPrompts(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM prompts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		prompts = append(prompts, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return prompts, nil
}
