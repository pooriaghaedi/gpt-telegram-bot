package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	// "strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

func contains(slice []string, str string) bool {
	for _, a := range slice {
		if a == str {
			return true
		}
	}
	return false
}

var users []string
var password = "securePassword"
var fileName = "data/users.txt"

var failedAttempts []string
var failedAttemptsFile = "data/failed_attempts.txt"

type User struct {
	ID         int
	TelegramId int
	Username   string
	IsValid    bool
	Credits    int
	CreatedAt  time.Time
}

func dbInit() {
	db, err := sql.Open("sqlite3", "data/users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table
	statement, err := db.Prepare(
		"CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, tg_id INTEGER, username TEXT, is_valid BOOLEAN, credits INTEGER, created_at TIMESTAMP)",
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

func main() {

	dbInit()
	db, err := sql.Open("sqlite3", "data/users.db")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(getUsers(db))
	secret := "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"
	passwd := "Test@123"
	// botAdmin := os.Getenv("BOTADMIN")
	bot, err := tgbotapi.NewBotAPI("5947642349:AAFtuy-Y38smuZSc4kyiYsKiobz76lC7TvM")

	if err != nil {
		log.Panic(err)
	}
	// replacements := map[string]string{
	// 	"_": "\\_",
	// 	"*": "\\*",
	// 	"[": "\\[",
	// 	"]": "\\]",
	// 	"(": "\\(",
	// 	")": "\\)",
	// 	"~": "\\~",
	// 	"`": "\\`",
	// 	">": "\\>",
	// 	"#": "\\#",
	// 	"+": "\\+",
	// 	"-": "\\-",
	// 	"=": "\\=",
	// 	"|": "\\|",
	// 	"{": "\\{",
	// 	"}": "\\}",
	// 	".": "\\.",
	// 	"!": "\\!",
	// }

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	const timeFormat string = "01-02-2006"

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			message := ""
			responseMessage := ""
			if isUserValid(db, int(update.Message.From.ID)) {
				fmt.Println("allowed " + update.Message.From.UserName)
				message, _ = gpt(secret, update.Message.Text)

			} else {
				if update.Message.Text == passwd {
					err := addUser(db, int(update.Message.From.ID), update.Message.From.UserName, true, 10)
					handleErr(err)
					message = "Access Granted, Now enjoy using ChatGPT :*"
				} else {
					err := addUser(db, int(update.Message.From.ID), update.Message.From.UserName, false, 0)
					handleErr(err)
					message = "Enter Password:"
					fmt.Println("denied " + update.Message.From.UserName)
				}
			}

			if strings.Contains(message, "```") {
				responseMessage = escapeMarkdownV2(message)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)
				msg.ReplyToMessageID = update.Message.MessageID
				msg.ParseMode = tgbotapi.ModeMarkdownV2
				_, err := bot.Send(msg)
				handleErr(err)
			} else {
				// If not a code block, escape other Markdown characters
				responseMessage = message
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)
				msg.ReplyToMessageID = update.Message.MessageID
				_, err := bot.Send(msg)
				handleErr(err)
			}

		}
	}
}

func escapeMarkdownV2(text string) string {
	// Characters to be escaped in MarkdownV2
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
