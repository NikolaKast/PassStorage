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
	"time"

	"github.com/atotto/clipboard"
	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

/*
проверить обраточики ошибок
заменить goto на цикл с кол-вом попыток +
Заменить в записи пароля возможность выйти при сохранении пароля +
реализовать logout +
реализовать получение пароля +
реализовать поиск пароля в получении +
Добавить обработчики и проверки неверных символов (по Панкову)
релизовать серверное взаимодействие
реализовать вызов с консоли по команде
Вывод всех аккаунтов
отображение паролей (небезопасный режим)
Добавить ограничение по бездействию + logout
Смена мастер пароля
Атомарная запись пароля без потери данных +
Проверка на clipboard
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
	fmt.Printf("'login' - Вход в аккаунт\n")
	fmt.Printf("'logout' - Выход из аккаунта\n")
	fmt.Printf("'register' - Регистрация\n")
	fmt.Printf("'savepass' - Сохранить пароль\n")
	fmt.Printf("'findpass' - Найти пароль\n")
	fmt.Printf("'update' - Поменять данные\n")
	fmt.Printf("'delete' - Удалить аккаунт\n")
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
		var master_pass []byte // проверить надо еще
		for i := 0; i < 3; i++ {
			fmt.Printf("Please, input new password\n")
			master_pass, _ := term.ReadPassword(int(syscall.Stdin))
			fmt.Printf("Please, repeat new password\n")
			master_control, _ := term.ReadPassword(int(syscall.Stdin))
			if string(master_pass) != string(master_control) {
				fmt.Printf("This is different passwords, try again\n")
			} else if i == 2 {
				fmt.Printf("So many tried, try later\n")
				return
			} else {
				break
			}
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
	tmpFile, _ := os.CreateTemp(".", "passstorage-*.tmp")
	tmpFile.Write(crypto_info)
	tmpFile.Sync()
	tmpFile.Close()
	os.Rename(tmpFile.Name(), json_name)
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
			fmt.Printf("Your data: %+v\n If you agree press 'y', to correct press 'n', to leave press other symbols \n", temp)
			var word string
			fmt.Scanln(&word)
			if word == "y" {
				fmt.Printf("You succesfully save your password\n")
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

func logout(cur_session *session) {
	clearScreen()
	if cur_session.access == false {
		fmt.Printf("You doesent logged\n")
		return
	}
	*cur_session = session{}
	fmt.Printf("You succesfulle logout\n")
}

func find(cur_session *session) int {
	clearScreen()
	if cur_session.access == false {
		fmt.Printf("You doesnt logged\n")
		return -1
	}
	fmt.Printf("Пожалуйста, введите URL или часть названия\n")
	var search string
	fmt.Scanln(&search)
	search = strings.ToLower(search)
	var results []pass_data
	var id []int
	var found int = 0
	for i := 0; i < cur_session.data.Cnt_pass; i++ {
		if strings.Contains(strings.ToLower(cur_session.data.Passwds[i].Url), search) {
			results = append(results, cur_session.data.Passwds[i])
			id = append(id, i)
			found++
		}

	}
	if found == 0 {
		clearScreen()
		fmt.Printf("Ничего не нашлось по данному запросу\n")
		return -1
	} else {
		fmt.Printf("Нашлось столько %d совпадений\n", found)
		for i := 0; i < found; i++ {
			fmt.Printf("№%d\nURL/data = %s\nlogin = %s\n\n", i, results[i].Url, results[i].Log)
		}
		fmt.Printf("Введите номер нужного аккаунта:\n")
		var num int
		fmt.Scanln(&num)
		clearScreen()
		return id[num]
	}
}

func findpass(cur_session *session) {

	num := find(cur_session)
	if num < 0 {
		return
	}
	clipboard.WriteAll(cur_session.data.Passwds[num].Pas)
	fmt.Printf("№%d\nURL/data = %s\nlogin = %s\n\n", num, cur_session.data.Passwds[num].Url, cur_session.data.Passwds[num].Log)
	fmt.Printf("Ваш пароль скопирован в буфер\n")
	fmt.Printf("Буфер будет очищен через 30 секунд или нажатием enter\n")
	timer := time.NewTimer(30 * time.Second)
	input := make(chan string)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		input <- text
	}()
	select {
	case <-timer.C:
		clipboard.WriteAll("")

	case <-input:
		clipboard.WriteAll("")
		timer.Stop()
	}
	clearScreen()
	return
}

func update(cur_session *session) {
	num := find(cur_session)
	if num < 0 {
		return
	}
	fmt.Printf("Что вы хотите изменить?\n1 - URL или данные\n2 - login\n3 - Пароль\n")
	var chose int
	fmt.Scanln(&chose)
	reader := bufio.NewReader(os.Stdin)
	switch chose {
	case 1:
		fmt.Printf("Введите новый URL\n")
		cur_session.data.Passwds[num].Url, _ = reader.ReadString('\n')
		cur_session.data.Passwds[num].Url = strings.TrimSpace(cur_session.data.Passwds[num].Url)
	case 2:
		fmt.Printf("Введите новый логин\n")
		cur_session.data.Passwds[num].Log, _ = reader.ReadString('\n')
		cur_session.data.Passwds[num].Log = strings.TrimSpace(cur_session.data.Passwds[num].Log)
	case 3:
		fmt.Printf("Введите новый пароль\n")
		cur_session.data.Passwds[num].Pas, _ = reader.ReadString('\n')
		cur_session.data.Passwds[num].Pas = strings.TrimSpace(cur_session.data.Passwds[num].Pas)
	default:
		clearScreen()
		return
	}
	update_list(cur_session)
	clearScreen()
}

func delete(cur_session *session) {
	num := find(cur_session)
	if num < 0 {
		return
	}
	fmt.Printf("Вы точно хотите удалить данный аккаунт?\nНажмите 'y' - если да, любой символ - нет\n")
	fmt.Printf("№%d\nURL/data = %s\nlogin = %s\n\n", num, cur_session.data.Passwds[num].Url, cur_session.data.Passwds[num].Log)
	var yes string
	fmt.Scanln(&yes)
	if yes == "y" {
		cur_session.data.Passwds = append(cur_session.data.Passwds[:num], cur_session.data.Passwds[num+1:]...)
		cur_session.data.Cnt_pass--
		update_list(cur_session)
		clearScreen()
		fmt.Printf("Аккаунт успешно удалён\n")
		return
	} else {
		clearScreen()
		return
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
		case "logout":
			logout(cur_session)
		case "findpass":
			findpass(cur_session)
		case "update":
			update(cur_session)
		case "delete":
			delete(cur_session)

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
