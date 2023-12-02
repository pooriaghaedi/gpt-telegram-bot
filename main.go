package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	// "strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

// func contains(slice []string, str string) bool {
// 	for _, a := range slice {
// 		if a == str {
// 			return true
// 		}
// 	}
// 	return false
// }

type User struct {
	ID         int
	TelegramId int
	Username   string
	IsValid    bool
	Credits    int
	CreatedAt  time.Time
	Prompt     string
}

func main() {

	dbInit()
	db, err := sql.Open("sqlite3", "data/users.db")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(getUsers(db))
	secret := os.Getenv("OPENAI_API_KEY")
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_KEY"))

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		switch update.Message.Command() {
		case "prompt":
			prompts, err := getAllPrompts(db) // db is your *sql.DB connection
			if err != nil {
				fmt.Println(err)
				// Optionally send an error message back to the user
				break
			}

			var rows [][]tgbotapi.KeyboardButton
			for _, prompt := range prompts {
				row := tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(prompt),
				)
				rows = append(rows, row)
			}

			// Add the "New prompt" button
			newPromptRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("New prompt"),
			)
			rows = append(rows, newPromptRow)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please choose:")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(rows...)
			// Send the message using your Telegram bot client
			// ...
			_, err = bot.Send(msg)
			if err != nil {
				handleErr(err)
			}

		default:
			// Handle other commands or add an empty default if not needed
		}

		if update.Message != nil {
			var msg tgbotapi.MessageConfig

			switch update.Message.Text {
			case "New prompt":
				// Send a message asking for the prompt in the specified format
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Please send the prompt in the format 'promptName - promptText'")
				_, err := bot.Send(msg)
				if err != nil {
					fmt.Println(msg)
				}

				// The next update from this user should be handled to capture the prompt
				// This part depends on your bot's logic to capture the next user message

			default:
				var message string
				if isUserValid(db, int(update.Message.From.ID)) {
					fmt.Println("allowed " + update.Message.From.UserName)
					prompt := getPrompt(db, int(update.Message.From.ID))
					msg := prompt + update.Message.Text
					message, _ = gpt(secret, msg)
				} else {
					message = processSecret(db, update.Message.Text, int(update.Message.From.ID), update.Message.From.UserName)
				}

				if strings.Contains(update.Message.Text, " - ") {
					// Check if the message is in the format 'promptName - promptText' and handle new prompt creation
					err := createNewPrompt(db, update.Message.Text)
					if err != nil {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Error creating prompt: "+err.Error())

					} else {
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "New prompt created successfully.")

					}
				} else {
					// Check if the message matches an existing prompt
					promptID, err := getPromptID(db, update.Message.Text)
					fmt.Println("Prompt ID:", promptID, "Error:", err)

					fmt.Println(promptID)
					if err == nil && promptID != 0 {
						updateUserPrompt(db, int(update.Message.From.ID), promptID)
						msg = tgbotapi.NewMessage(update.Message.Chat.ID, "your default prompt is changed.")
						// break
					} else {
						if strings.Contains(message, "```") {
							msg = tgbotapi.NewMessage(update.Message.Chat.ID, escapeMarkdownV2(message))
							msg.ParseMode = tgbotapi.ModeMarkdownV2
						} else {
							msg = tgbotapi.NewMessage(update.Message.Chat.ID, message)
						}
						msg.ReplyToMessageID = update.Message.MessageID
					}
				}

				// Send the final constructed message
				_, err := bot.Send(msg)
				if err != nil {
					handleErr(err)
				}
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
