package main
import (
	"fmt"
	"os"
	//"encoding/json"
)

type session struct{
	name string
	access bool
}


func clearScreen() {
    fmt.Print("\033[2J")  
    fmt.Print("\033[H") 
}

func start(cur_session *session){
	fmt.Print("\033[?1049h")
	fmt.Print("\033[H") 
	fmt.Printf("PassStorage\nCreated by Kast\n")
	waitcommand(cur_session)
}

func help(cur_session *session){
	fmt.Printf("U can use this commands:\n")
	fmt.Printf("'help' - Вывод данной справки\n")
	fmt.Printf("'exit' - Выход из программы\n")
	waitcommand(cur_session)
}

func incorrect(cur_session *session){
	fmt.Printf("You input incorrect command ('help' to see list)\n")
	waitcommand(cur_session)
}

func login(cur_session *session){
	if cur_session.name != ""{
		clearScreen()
		fmt.Printf("You allready login\n")
		waitcommand(cur_session)
	}
	fmt.Printf("Hello, please input your username:\n")
	var username string
	fmt.Scanln(&username)
}

func waitcommand(cur_session *session){
	fmt.Printf("Please input command (help - to see all commands)\n")
	var command string
	fmt.Printf("> ")
	fmt.Scanln(&command)
	if command == "help"{
		clearScreen()
		help(cur_session)
	}else if command == "exit"{
		fmt.Print("\033[?1049l")
		os.Exit(0)
	}else if command == "login"{
		login(cur_session)
	}else if command == "new"{

	}else{
		clearScreen()
		incorrect(cur_session)
	}
}

func main(){
	cur_session := new(session)
	defer fmt.Print("\033[?1049l")
	start(cur_session)
}