package oper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/client/chttp"
	"github.com/vzveiteskostrami/goph-keeper/internal/client/config"
	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func Registration(login string, password string) {
	err := misc.MakeDir("ADM")
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

	delim := dialog.DrawHeader("Регистрация нового пользователя", false)
	for {
		fmt.Print(delim + "\n\r")
		if login == "" {
			if login = dialog.GetAnswer("Введите желаемый логин:", false, false); login == "-" {
				fmt.Print(refuse)
				break
			}
		} else {
			fmt.Print("Логин ", login, "\n\r")
		}
		if password == "" {
			if password = dialog.GetAnswer("Введите желаемый пароль:", true, false); password == "-" {
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
					cfg := config.Get()
					Authorization(login, password, *cfg.Place, 0)
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

func Authorization(login string, password string, place int, sessDuration int64) {
	if sessDuration < 1 || sessDuration > 5*365*24*60 {
		sessDuration = 0
	}

	delim := dialog.DrawHeader("Авторизация", false)
	for {
		fmt.Print(delim + "\n\r")
		if login == "" {
			if login = dialog.GetAnswer("Введите логин:", false, false); login == "-" {
				fmt.Print(refuse)
				return
			}
		} else {
			fmt.Print("Логин ", login, "\n\r")
		}
		if password == "" {
			if password = dialog.GetAnswer("Введите пароль:", true, false); password == "-" {
				fmt.Print(refuse)
				return
			}
		} else {
			fmt.Print("Пароль введён\n\r")
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

		if place == co.SessionNotDefined {
			rs := dialog.Menu([]string{"Выберите место авторизации:", "Локальное хранилище", "Сервер и локальное хранилище"})
			if rs == 0 {
				break
			}
			if rs == 1 {
				place = co.SessionLocal
			} else {
				place = co.SessionBoth
			}
		}

		var err error
		if place == co.SessionBoth {
			var code int
			fmt.Print("Попытка авторизоваться на сервере...")
			code, err = chttp.Authorization(login, password, sessDuration)
			if err != nil {
				fmt.Println("\rВо время авторизации сервер вернул ошибку:")
				fmt.Println(err.Error())
				if !dialog.Yn("Попробовать ещё раз") {
					break
				}
			}
			if code == http.StatusUnauthorized {
				login = ""
				password = ""
			}

			if code == http.StatusOK {
				fmt.Println("\rАвторизация на сервере успешна.           ")
			}
		}

		if err == nil && (place == co.SessionLocal || place == co.SessionBoth) {
			fmt.Print("Попытка авторизоваться локально...")
			var checked bool
			checked, err = localAuthorization(login, password, place)
			if err != nil {
				fmt.Println("\rВо время локальной авторизации произошла ошибка:")
				fmt.Println(err.Error())
				if !dialog.Yn("Попробовать ещё раз") {
					break
				}
				if checked {
					login = ""
					password = ""
				}
			} else {
				fmt.Print("\rСохранение токена...")
				if err = saveLocalToken(login, time.Now().Add(time.Duration(sessDuration)*time.Minute)); err != nil {
					fmt.Println("\rВы авторизовались, но во время сохранения сессии произошла ошибка:")
					fmt.Println(err.Error())
					if !dialog.Yn("Попробовать ещё раз") {
						break
					}
				}
			}
		}

		if err == nil {
			ss := "\rВы авторизованы "
			if place == co.SessionBoth {
				ss += "на сервере и "
			}
			ss += "в локальном хранилище. Сессия валидна в течение"
			fmt.Println(ss, sessDuration, "минут.")
			break
		}
	}
}

func SaveLastOperationDateTime() {
	err := misc.MakeDir("ADM")
	if err != nil {
		fmt.Println("Не удалось создать административную директорию. " + err.Error())
		return
	}

	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Не удалось получить ключ. Ошибка: " + err.Error())
		return
	}

	var date []authData
	da := authData{Until: time.Now()}
	date = append(date, da)
	b, err := json.Marshal(date)
	if err != nil {
		fmt.Println("Не удалось закодировать date. Ошибка:" + err.Error())
		return
	}
	err = misc.SaveToFileProtectedZIP("ADM\\local_setts", "setts", key, b)
	if err != nil {
		fmt.Println("Сохранение файла local_setts. Ошибка:" + err.Error())
	}
}

func CheckLastOperationDateTime() bool {
	ok, err := misc.FileExists("ADM\\local_setts")
	if err == nil {
		if !ok {
			fmt.Println("Нет файла local_setts. Выполните регистрацию или авторизацию.")
			return false
		}
		key, err := misc.UnicKeyForExeDir()
		if err == nil {
			var raw []byte
			raw, _, err = misc.ReadFromFileProtectedZIP("ADM\\local_setts", key)
			if err == nil {
				var ada []authData
				ada, err = getAuthData(raw)
				if err == nil {
					if time.Until(ada[0].Until) > 0 {
						fmt.Println("Ну вот нахера так делать?")
						return false
					}
				}
			}
		}
	}
	if err != nil {
		fmt.Println("Чтение файла local_setts. Ошибка:" + err.Error())
	}
	return err == nil
}
