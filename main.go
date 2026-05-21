package main
import (
	"fmt"
	"os"
	"time"
)


func clearScreen() {
    fmt.Print("\033[2J")  
    fmt.Print("\033[H") 
}

func start(){
	fmt.Print("\033[?1049h")
	fmt.Print("\033[H") 
	fmt.Printf("PassStorage\nCreated by Kast\n")
	fmt.Printf("Please input command (help - to see all commands)\n")
	waitcommand()
}

func help(){
	fmt.Printf("U can use this commands:\n")
	fmt.Printf("'help' - Вывод данной справки\n")
	fmt.Printf("'exit' - Выход из программы\n")
	waitcommand()
}

func incorrect(){
	fmt.Printf("You input incorrect command ('help' to see list)\n")
	waitcommand()
}

func waitcommand(){
	var command string
	fmt.Scanln(&command)
	if command == "help"{
		clearScreen()
		help()
	}else if command == "exit"{
		fmt.Print("\033[?1049l")
		os.Exit(0)
	}else{
		clearScreen()
		incorrect()
	}
}

func main(){
	defer fmt.Print("\033[?1049l")
	start()
	time.Sleep(3 * time.Second)
}