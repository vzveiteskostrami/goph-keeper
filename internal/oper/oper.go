package oper

import (
	"fmt"
	"os"

	"github.com/vzveiteskostrami/goph-keeper/internal/chttp"
)

const refuse string = "Отказ от выполнения операции"

func Registration(login string, password string) error {
	fmt.Print("Регистрация нового пользователя")
	if login == "" {
		if login = getAnswer("Введите желаемый логин:"); login == "-" {
			fmt.Print(refuse)
			return nil
		}
	} else {
		fmt.Print("\n\rЛогин ", login)
	}
	if password == "" {
		if password = getAnswer("Введите желаемый пароль:"); password == "-" {
			fmt.Print(refuse)
			return nil
		}
	} else {
		fmt.Print("Пароль ", password)
	}

	for {
		err := chttp.Registration(login, password)
		if err == nil {
			break
		} else {
			fmt.Println("Во время регистрации произошла ошибка.")
			if !yn("Попробовать ещё раз") {
				break
			}
		}
	}

	return nil
}

func getAnswer(prompt string) string {
	var answer string
	for {
		fmt.Print("\n\r" + prompt)
		fmt.Fscan(os.Stdin, &answer)
		if answer != "" {
			break
		}
	}
	return answer
}

func yn(prompt string) bool {
	var answer string
	for {
		fmt.Print("\n\r" + prompt + " (y/n)?")
		fmt.Fscan(os.Stdin, &answer)
		if answer == "y" || answer == "n" {
			break
		}
		answer = ""
	}
	return answer == "y"
}
