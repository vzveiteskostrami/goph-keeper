package oper

import (
	"fmt"
	"net/http"

	"github.com/vzveiteskostrami/goph-keeper/internal/chttp"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func Registration(login string, password string) {
	err := makeLocalDir("ADM")
	if err != nil {
		fmt.Println("Не удалось создать административную директорию. Ошибка:")
		fmt.Println(err.Error())
		return
	}

	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Не удалось получить ключ. Ошибка:")
		fmt.Println(err.Error())
		return
	}

	err = checkVaildRegisterFile(key)
	if err != nil {
		fmt.Println("Подмена или крушение регистрационного файла! Ошибка:")
		fmt.Println(err.Error())
		return
	}

	fmt.Print("----------------------------------\n\r")
	fmt.Print("-- Регистрация нового пользователя\n\r")
	for {
		fmt.Print("----------------------------------\n\r")
		if login == "" {
			if login = dialog.GetAnswer("Введите желаемый логин:"); login == "-" {
				fmt.Print(refuse)
				break
			}
		} else {
			fmt.Print("Логин ", login, "\n\r")
		}
		if password == "" {
			if password = dialog.GetAnswer("Введите желаемый пароль:"); password == "-" {
				fmt.Print(refuse)
				break
			}
		} else {
			fmt.Print("Пароль ", password, "\n\r")
		}

		fmt.Print("Попытка зарегистрироваться...")
		code, err := chttp.Registration(login, password)
		if err == nil {
			err = registerLocally(login, password)
			if err != nil {
				fmt.Println("\rРегистрация на сервере успешна. Локальная регистрация не удалась. Ошибка:")
				fmt.Println(err.Error())
				return
			} else {
				fmt.Println("\rВы зарегистрированы! Добро пожаловать на борт!")
				if dialog.Yn("Желаете открыть сессию прямо сейчас") {
					Authorization(login, password, 0)
				}
				break
			}
		} else {
			if code == http.StatusConflict {
				if !dialog.Yn("\rЖелаемый логин занят. Попытаться зарегистрироваться под другим логином") {
					break
				}
				login = ""
			} else {
				fmt.Println("\rВо время регистрации сервер вернул ошибку:")
				fmt.Println(err.Error())
				if !dialog.Yn("Попробовать ещё раз") {
					break
				}
			}
		}
	}
}

func Authorization(login string, password string, sessDuration int64) {
	fmt.Print("----------------------------------\n\r")
	fmt.Print("-- Авторизация\n\r")
	for {
		fmt.Print("----------------------------------\n\r")
		if login == "" {
			if login = dialog.GetAnswer("Введите логин:"); login == "-" {
				fmt.Print(refuse)
				return
			}
		} else {
			fmt.Print("Логин ", login, "\n\r")
		}
		if password == "" {
			if password = dialog.GetAnswer("Введите пароль:"); password == "-" {
				fmt.Print(refuse)
				return
			}
		} else {
			fmt.Print("Пароль ", password, "\n\r")
		}

		if sessDuration == 0 {
			ok := true
			if sessDuration, ok = dialog.GetInt("Введите желаемую продолжительность сессии (в минутах):",
				true, 1, true, 5*365*24*60); !ok {
				fmt.Print(refuse)
				break
			}
		} else {
			fmt.Print("Продолжительность сессии ", sessDuration, " минут.\n\r")
		}

		fmt.Print("Попытка авторизоваться...")
		_, err := chttp.Authorization(login, password, sessDuration)
		if err == nil {
			fmt.Println("\rВы авторизованы. Сессия валидна в течение", sessDuration, "минут.")
			//fixCurrentUser(login, time.Now().Add(time.Duration(sessDuration)*time.Minute))
			break
		} else {
			fmt.Println("\rВо время авторизации сервер вернул ошибку:")
			fmt.Println(err.Error())
			if !dialog.Yn("Попробовать ещё раз") {
				break
			}
		}
	}
}

func Syncronize() {
	fmt.Print("----------------------------------\n\r")
	fmt.Print("-- Синхронизация\n\r")
	fmt.Print("----------------------------------\n\r")
	_, err := chttp.Syncronize()
	if err != nil {
		fmt.Println("\rВо время синхронизации сервер вернул ошибку:")
		fmt.Println(err.Error())
	}
}

/*
func Session() {
	resp, err := http.Get(serverName + serverPort + "readcookie")
	if err != nil {
		log.Fatalln(err)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SessionID" {
			log.Println("Cookie: ", cookie.Value)
		}
	}
}
*/
