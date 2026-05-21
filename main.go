package main

import (
	"fmt"
	"os"
	//"encoding/json"
)

type session struct {
	name   string
	access bool
}
type main_json struct {
	Name     string
	Cnt_pass int
	Passwds  []pass_data
}
type pass_data struct {
	Url string
	Log string
	Pas string
}

func clearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}

func start(cur_session *session) {
	fmt.Print("\033[?1049h")
	fmt.Print("\033[H")
	fmt.Printf("PassStorage\nCreated by Kast\n")
	waitcommand(cur_session)
}

func help(cur_session *session) {
	fmt.Printf("U can use this commands:\n")
	fmt.Printf("'help' - Вывод данной справки\n")
	fmt.Printf("'exit' - Выход из программы\n")
	waitcommand(cur_session)
}

func incorrect(cur_session *session) {
	fmt.Printf("You input incorrect command ('help' to see list)\n")
	waitcommand(cur_session)
}

func login(cur_session *session) {
	clearScreen()
	if cur_session.name != "" {
		fmt.Printf("You allready login\n")
		waitcommand(cur_session)
	}
	fmt.Printf("Hello, please input your username:\n")
	var username string
	fmt.Scanln(&username)
	var json_name string = username + ".json"
	data, err := os.ReadFile(json_name)
	if err != nil {
		clearScreen()
		fmt.Printf("Account with username: '%s' doesnt exist\n", username)
		waitcommand(cur_session)
	} else {
		clearScreen()
		fmt.Printf("You succesfully logged as %s\n", username)
		fmt.Println(data) // Допилить
		cur_session.name = username
		waitcommand(cur_session)
	}
}

func register(cur_session *session) {
	clearScreen()
	fmt.Println("Please input your username")
	var username string
	fmt.Scanln(&username)
	var json_name string = username + ".json"
	data, err := os.ReadFile(json_name)
	if err != nil {
		file, err := os.Create(json_name)
		if err != nil {
			panic(err)
		}
		file.Close()
		fmt.Printf("You registered as '%s' and logged\n", username)
		cur_session.name = username
		waitcommand(cur_session)
	} else {
		clearScreen()
		fmt.Printf("You already registered\n")
		waitcommand(cur_session)
	}
	fmt.Println(data)

}

func get(cur_session *session) {
	if cur_session.name == "" {
		fmt.Printf("You doesnt logged, use 'login'\n")
		waitcommand(cur_session)
	}
	var json_name string = cur_session.name + ".json"
	data, err := os.ReadFile(json_name)
	if err != nil {
		panic(err)
	}

}

func waitcommand(cur_session *session) {
	if cur_session.name == "" {
		fmt.Printf("You doesnt logged, use 'login'\n")
	} else {
		fmt.Printf("You logged as %s\n", cur_session.name)
	}
	fmt.Printf("Please input command (help - to see all commands)\n")
	var command string
	fmt.Printf("> ")
	fmt.Scanln(&command)
	switch command {
	case "help":
		clearScreen()
		help(cur_session)
	case "exit":
		fmt.Print("\033[?1049l") // Возврат буфера консоли
		os.Exit(0)
	case "login":
		login(cur_session)
	case "register":
		register(cur_session)
	case "get":
		get(cur_session)
	default:
		clearScreen()
		incorrect(cur_session)
	}
}

func main() {
	cur_session := new(session)
	cur_session.access = false
	defer fmt.Print("\033[?1049l")
	start(cur_session)
}
