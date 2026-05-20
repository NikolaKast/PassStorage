package main
import (
	"fmt"
	"os"
	"time"
)


func start(){
	fmt.Print("\033[?1049h")
	fmt.Print("\033[H") 
	fmt.Printf("PassStorage\nCreated by Kast\n")
	fmt.Printf("Please input command (help - to see all commands)\n")
	waitcommand()
}

func help(){
	fmt.Printf("U can use this commands:\n")
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
		help()
	}else if command == "exit"{
		fmt.Print("\033[?1049l")
		os.Exit(0)
	}else{
		incorrect()
	}
}

func main(){
	defer fmt.Print("\033[?1049l")
	start()
	time.Sleep(3 * time.Second)
}