package main

import (
	"bufio"
	"fmt"
	"os"
)

const (
	ClientChoice      = "client"
	ServerChoice      = "server"
	QuitChoice        = "quit"
	QuitShorterChoice = "q"
)

var AvailableChoices = map[string]struct{}{
	ClientChoice:      {},
	ServerChoice:      {},
	QuitChoice:        {},
	QuitShorterChoice: {},
}

func CheckLegalChoice(userInput string) bool {
	if _, ok := AvailableChoices[userInput]; ok {
		return true
	}
	return false
}

func HandleDifferentChoice(choice string) error {
	var err error
	switch choice {
	case QuitChoice:
	case QuitShorterChoice:
		fmt.Println("Program Exit")
		os.Exit(0)
	case ClientChoice:
		err = StartClient()
		if err != nil {
			return fmt.Errorf("error starting client: %w", err)
		} else {
			return nil
		}
	case ServerChoice:
		err = StartServer()
		if err != nil {
			return fmt.Errorf("error starting server: %w", err)
		} else {
			return nil
		}
	}
	return nil
}

func main() {
	// 获取用户输入, 看用户选择的是客户端还是服务器端
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input (1.client) (2.server) (3.q or quit)")
	if scanner.Scan() {
		line := scanner.Text()
		if !CheckLegalChoice(line) {
			fmt.Println(fmt.Sprintf("User Input: %s, error", line))
		}
		err := HandleDifferentChoice(line)
		if err != nil {
			fmt.Printf("handle different choice error: %v", err)
		}
	}
}
