package main

import (
	"fmt"
	"log"
	"os"
)

func initial() {
	err := os.Mkdir(os.ExpandEnv("$HOME/.go-chatgpt"), 0700)
	if err != nil {
        fmt.Println(err)
        return
    }
	fmt.Println("Enter your ChatGPT API key:")

	var creds string
	fmt.Scanln(&creds)
	c1 := []byte(creds)
	err = os.WriteFile(os.ExpandEnv("$HOME/.go-chatgpt/credentials"), c1, 400)
	if err != nil {
	 log.Fatal(err)
	}
}