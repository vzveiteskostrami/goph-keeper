package cconfig

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

type Configc struct {
	Login           *string `json:"login,omitempty"`
	Password        *string
	Operation       *string
	ServerAddress   *string `json:"server_address,omitempty"`
	Place           *int
	SessionDuration *int64
	Brief           *bool
	EntityName      *string
	WriteLogin      *string
	WritePassword   *string
	WriteFile       *string
	EntityKind      *int
}

var cfg Configc

func ReadData() error {
	ok, err := misc.FileExists("config.json")
	if err != nil {
		return fmt.Errorf("ошибка проверки наличия config файла: " + err.Error())
	}

	if !ok {
		err = makeConfig()
		if err != nil {
			return fmt.Errorf("ошибка проверки наличия config файла: " + err.Error())
		}
	}

	content, err := os.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("ошибка чтения config файла: " + err.Error())
	}

	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return fmt.Errorf("ошибка парсинга config файла: " + err.Error())
	}

	if cfg.ServerAddress == nil || *cfg.ServerAddress == "" {
		fmt.Println()
		return fmt.Errorf("в config.json не указан server_address")
	}

	s := []rune(*cfg.ServerAddress)
	if s[len(s)-1] != []rune("/")[0] {
		*cfg.ServerAddress += "/"
	}

	cfg.Operation = flag.String("o", "", "User's operation, strictly")
	lg := flag.String("l", "Не обязательно указывать", "User's login, optional")
	cfg.Password = flag.String("p", "", "User's password, optional")
	sessBoth := flag.Bool("both", false, "Create session for both places (Server & Client), optional")
	sessLocal := flag.Bool("local", false, "Create session for client only, optional")
	cfg.SessionDuration = flag.Int64("du", 0, "Desired session duration, minutes, optional")
	cfg.Brief = flag.Bool("brief", false, "Show brief or full info anywhere, optional")
	en := flag.String("e", "", "Kind of new entity, optional. Values: pass, card, text, bin")
	cfg.EntityName = flag.String("n", "", "Name, in any cases, entity name, like example, optional")
	cfg.WriteLogin = flag.String("wl", "", "Saved login for entity \"pass\", optional")
	cfg.WritePassword = flag.String("wp", "", "Saved password for entity \"pass\", optional")
	cfg.WriteFile = flag.String("f", "", "File name, in any cases, for save to entity, as instance, optional")

	flag.Parse()

	if *lg != "Не обязательно указывать" {
		cfg.Login = lg
	} else {
		if cfg.Login == nil {
			s := ""
			cfg.Login = &s
		}
	}

	a := co.EntityNotDefined
	cfg.EntityKind = &a
	if *en != "" {
		if *en == "pass" {
			*cfg.EntityKind = co.EntityLoginPassword
		} else if *en == "card" {
			*cfg.EntityKind = co.EntityCard
		} else if *en == "txt" {
			*cfg.EntityKind = co.EntityText
		} else if *en == "bin" {
			*cfg.EntityKind = co.EntityBinary
		} else {
			fmt.Println("!!! Не опознана сущность -e " + *en + ".")
			fmt.Println("Сброшено.")
		}
	}

	i := co.SessionNotDefined
	cfg.Place = &i
	if *sessBoth {
		*cfg.Place = co.SessionBoth
	} else if *sessLocal {
		*cfg.Place = co.SessionLocal
	}

	return nil
}

func Get() Configc {
	return cfg
}

type cof struct {
	Login  *string `json:"login,omitempty"`
	Server *string `json:"server_address,omitempty"`
}

func makeConfig() (err error) {
	fmt.Println("---------------------------------------------")
	fmt.Println("-- Отсутствует файл конфигурации config.json.")
	fmt.Println("-- Попытаемся создать его интерактивно.")
	fmt.Println("---------------------------------------------")
	fmt.Println("Логин по умолчанию. Не обязательно для заполнения.")
	fmt.Println("Если не ввести, то логин придётся указывать для каждой")
	fmt.Println("сессии клиента. Если вы один пользуетесь приложением,")
	fmt.Println("имеет смысл ввести логин один раз. Если на этом компьютере")
	fmt.Println("несколько пользователей приложения, логин по умолчанию")
	fmt.Println("лучше не вводить.")
	login := dialog.GetAnswer("Ведите:", false, true)
	if login == "-" {
		err = fmt.Errorf("процесс прерван")
		return
	}

	fmt.Println("URL адрес сервера. Обязательно к заполнению.")
	fmt.Println("Без серверного хранилища невозможна регистрация")
	fmt.Println("и авторизация пользователей.")
	url := ""
	for {
		url = dialog.GetAnswer("Ведите:", false, false)
		if url == "-" {
			err = fmt.Errorf("отказ от ввода url. Процесс прерван")
			return
		}
		s := []rune(url)
		if s[len(s)-1] != []rune("/")[0] {
			url += "/"
		}
		err = checkServer(url)
		if err == nil {
			break
		} else {
			fmt.Println("Ошибка соединения с сервером:")
			fmt.Println(err)
		}
	}

	co := cof{}
	co.Server = &url
	if login != "" {
		co.Login = &login
	}

	b, _ := json.Marshal(co)
	err = os.WriteFile("config.json", b, 0644)
	if err == nil {
		fmt.Println("config.json успешно создан.")
	}
	return err
}

func checkServer(url string) error {
	req, err := http.NewRequest("GET", url+"ready", nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5000*time.Millisecond)
	defer cancel()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("не положительный ответ сервера")
	}
	return nil
}
