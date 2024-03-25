package oper

import (
	"fmt"
	"net/http"

	"github.com/vzveiteskostrami/goph-keeper/internal/chttp"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
)

func Registration(login string, password string) error {
	fmt.Print("----------------------------------\n\r")
	fmt.Print("-- Регистрация нового пользователя\n\r")
	for {
		fmt.Print("----------------------------------\n\r")
		if login == "" {
			if login = dialog.GetAnswer("Введите желаемый логин:"); login == "-" {
				fmt.Print(refuse)
				return nil
			}
		} else {
			fmt.Print("Логин ", login, "\n\r")
		}
		if password == "" {
			if password = dialog.GetAnswer("Введите желаемый пароль:"); password == "-" {
				fmt.Print(refuse)
				return nil
			}
		} else {
			fmt.Print("Пароль ", password, "\n\r")
		}

		fmt.Print("Попытка зарегистрироваться...")
		code, err := chttp.Registration(login, password)
		if err == nil {
			fmt.Println("\rВы зарегистрированы! Добро пожаловать на борт!")
			if !dialog.Yn("Желаете получить токен прямо сейчас") {
				break
			} else {
				return nil
			}
		} else {
			if code == http.StatusConflict {
				if !dialog.Yn("\rЖелаемый логин занят. Желаете зарегистрироваться под другим логином") {
					return nil
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
	return nil
}
