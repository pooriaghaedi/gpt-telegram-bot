package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func gpt(secret, message string) (string, int) {
	client := openai.NewClient(secret)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "error", 1
	}

	return resp.Choices[0].Message.Content, resp.Usage.TotalTokens
	// return(resp.Usage)
}
