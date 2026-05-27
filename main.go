package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

/*
Избавится от поля cnt_pass
проверить обраточики ошибок
заменить goto на цикл с кол-вом попыток
Заменить в записи пароля возможность выйти при сохранении пароля
реализовать logout
реализовать получение пароля
реализовать поиск пароля в получении
*/

type session struct {
	name   string
	data   main_json
	access bool
	key    []byte
}
type main_json struct {
	Name     string      `json:"Name"`
	Cnt_pass int         `json:"Count_passwords"`
	Passwds  []pass_data `json:"Passwords,omitempty"`
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

func help() {
	fmt.Printf("U can use this commands:\n")
	fmt.Printf("'help' - Вывод данной справки\n")
	fmt.Printf("'exit' - Выход из программы\n")
}

func incorrect() {
	fmt.Printf("You input incorrect command ('help' to see list)\n")
}

func hash_json(username string) string {
	hash_name := sha256.Sum256([]byte(username))
	hashString := hex.EncodeToString(hash_name[:])
	var json_name string = hashString + ".json"
	return json_name
}

func login(cur_session *session) { // Добавить обработку пароля, расшифрования
	clearScreen()
	if cur_session.name != "" {
		fmt.Printf("You allready login\n")
		return
	}
	fmt.Printf("Hello, please input your username:\n")
	var username string
	fmt.Scanln(&username)
	var json_name string = hash_json(username)
	crypto_data, err := os.ReadFile(json_name)
	if err != nil {
		clearScreen()
		fmt.Printf("Account with username: '%s' doesnt exist\n", username)
		return
	} else {
		clearScreen()
		fmt.Printf("Please, input password\n")
		master_pass, _ := term.ReadPassword(int(syscall.Stdin))
		key := argon2.IDKey(master_pass, []byte(username), 1, 64*1024, 4, 32)
		data, err := decrypt(crypto_data, key)
		var clear_slot main_json
		json.Unmarshal(data, &clear_slot)
		if err != nil {
			fmt.Printf("Incorrect password or corrupted file\n")
			return
		}
		cur_session.key = key
		cur_session.data = clear_slot
		cur_session.access = true
		fmt.Printf("You succesfully logged as %s\n%+v\n", username, clear_slot)
		cur_session.name = username
		return
	}
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// 1. Создаём AES-шифр
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. Создаём GCM-режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 3. Извлекаем IV из начала зашифрованных данных
	nonceSize := gcm.NonceSize()
	iv := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// 4. Расшифровываем (GCM проверяет тег аутентификации)
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: wrong password or corrupted data")
	}

	return plaintext, nil
}

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. Создаём GCM-режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 3. Генерируем случайный IV
	iv := make([]byte, gcm.NonceSize()) // для AES-GCM это 12 байт
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// 4. Шифруем (GCM добавляет тег аутентификации в конец)
	ciphertext := gcm.Seal(iv, iv, plaintext, nil)

	return ciphertext, nil
}

func register(cur_session *session) {
	clearScreen()
	fmt.Println("Please input your username")
	var username string
	fmt.Scanln(&username)
	var json_name string = hash_json(username)
	_, err := os.ReadFile(json_name)
	if err != nil {
		var init main_json
		init.Cnt_pass = 0
		init.Name = username
		init_js, _ := json.MarshalIndent(init, "", "    ")
	again:
		fmt.Printf("Please, input new password\n")
		master_pass, _ := term.ReadPassword(int(syscall.Stdin))
		fmt.Printf("Please, repeat new password\n")
		master_control, _ := term.ReadPassword(int(syscall.Stdin))
		if string(master_pass) != string(master_control) {
			fmt.Printf("This is different passwords, try again\n")
			goto again
		}
		key := argon2.IDKey(master_pass, []byte(username), 1, 64*1024, 4, 32)
		crypto_info, err := encrypt([]byte(init_js), key) // Обрабочик ошибок
		if err != nil {
			fmt.Printf("Error with encrypting\n")
			return
		}
		os.WriteFile(json_name, crypto_info, 0644)
		cur_session.key = key
		cur_session.data = init
		cur_session.access = true
		fmt.Printf("You registered as '%s' and logged\n", username)
		cur_session.name = username
		return
	} else {
		clearScreen()
		fmt.Printf("You already registered\n")
		return
	}
}

func get(cur_session *session) {
	if cur_session.name == "" {
		fmt.Printf("You doesnt logged, use 'login'\n")
		waitcommand(cur_session)
	}
	var json_name string = cur_session.name + ".json"
	_, err := os.ReadFile(json_name)
	if err != nil {
		panic(err)
	}

}

func update_list(cur_session *session) {
	data, _ := json.MarshalIndent(cur_session.data, "", "    ")
	crypto_info, err := encrypt([]byte(data), cur_session.key) // Обрабочик ошибок
	if err != nil {
		fmt.Printf("Error with encrypting\n")
		return
	}
	var json_name string = hash_json(cur_session.name)
	os.WriteFile(json_name, crypto_info, 0644)
}

func savepass(cur_session *session) {
	for {
		if cur_session.access == false {
			fmt.Printf("Вы не вошли в аккаунт\n")
			return
		}
		reader := bufio.NewReader(os.Stdin) // оборачивает stdin
		fmt.Printf("Please, input URL or info about service:\n")
		var temp pass_data
		temp.Url, _ = reader.ReadString('\n')
		temp.Url = strings.TrimSpace(temp.Url)
		fmt.Printf("Please, input login:\n")
		temp.Log, _ = reader.ReadString('\n')
		temp.Log = strings.TrimSpace(temp.Log)
		fmt.Printf("Please, input password:\n")
		temp.Pas, _ = reader.ReadString('\n')
		temp.Pas = strings.TrimSpace(temp.Pas)
		for {
			clearScreen()
			fmt.Printf("Your data: %+v\n If you agree press 'y', else press 'n'\n", temp)
			var word string
			fmt.Scanln(&word)
			if word == "y" {
				fmt.Printf("You succesfulle save your password\n")
				cur_session.data.Passwds = append(cur_session.data.Passwds, temp)
				cur_session.data.Cnt_pass++
				update_list(cur_session)
				return
			} else if word == "n" {
				break
			} else {
				return
			}
		}

	}
}

func waitcommand(cur_session *session) {
	for {
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
			help()
		case "exit":
			fmt.Print("\033[?1049l") // Возврат буфера консоли
			os.Exit(0)
		case "login":
			login(cur_session)
		case "register":
			register(cur_session)
		case "get":
			get(cur_session)
		case "savepass":
			savepass(cur_session)

		default:
			clearScreen()
			incorrect()
		}
	}
}

func main() {
	cur_session := new(session)
	cur_session.access = false
	defer fmt.Print("\033[?1049l")
	start(cur_session)
}
