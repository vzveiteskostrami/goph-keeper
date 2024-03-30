package main

import (
	"fmt"

	"github.com/vzveiteskostrami/goph-keeper/internal/cconfig"
	"github.com/vzveiteskostrami/goph-keeper/internal/chttp"
	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/oper"
)

func main() {
	if err := cconfig.ReadData(); err != nil {
		fmt.Println("Ошибка чтения конфигурации:", err)
		return
	}

	cfg := cconfig.Get()

	if cfg.Operation == nil || *cfg.Operation == "" {
		fmt.Println("Не указана операция \"-o=<operation>\"")
		return
	}

	// Для отлова хитрожопых, которые поменяют системное время и таким образом
	// захотят продлить время работы чужого локального токена вспять.
	// Если это регистрация или новая сессия, то пофиг. А вот если какие-то
	// операции, то надо проверить, что текущее время строго больше времени
	// последней проведённой операции.
	if !(*cfg.Operation == "registration" ||
		*cfg.Operation == "session") {
		if !oper.CheckLastOperationDateTime() {
			return
		}
	}

	// Если операция требует обращения к серверу, сначала просто проверим его
	// наличие в системе. Если его нет, то и затевать ничего не надо.
	if *cfg.Operation == "registration" ||
		(*cfg.Operation == "session" && *cfg.Place == co.SessionBoth) ||
		*cfg.Operation == "sync" {
		err := chttp.CheckServerPresent()
		if err != nil {
			fmt.Println("Сервер неработоспособен. Операция в данный момент невозможна.")
			fmt.Println("Попробуйте выполнить операцию позднее.")
			fmt.Println("При проверке сервера произошла ошибка:")
			fmt.Println(err)
			return
		}
	}

	sessionOwner := ""
	// Если операция требует наличия локальной сессии, проверим, что она открыта и
	// не истекла.
	if *cfg.Operation == "get" ||
		*cfg.Operation == "set" ||
		*cfg.Operation == "strict" ||
		*cfg.Operation == "list" ||
		*cfg.Operation == "delete" ||
		*cfg.Operation == "new" {
		var err error
		sessionOwner, err = oper.CheckLocalSession()
		if err != nil {
			fmt.Println("Неудачная проверка локальной сессии:")
			fmt.Println(err)
			return
		} else {
			fmt.Println("Владелец сессии", sessionOwner)
		}
	}

	if *cfg.Operation == "registration" {
		oper.Registration(*cfg.Login, *cfg.Password)
	} else if *cfg.Operation == "session" {
		oper.Authorization(*cfg.Login, *cfg.Password, *cfg.Place, *cfg.SessionDuration)
	} else if *cfg.Operation == "list" {
		oper.ShowEntityList(sessionOwner, *cfg.Brief, *cfg.EntityKind)
	} else if *cfg.Operation == "new" {
		oper.NewEntity(sessionOwner,
			*cfg.EntityKind,
			*cfg.EntityName,
			*cfg.WriteLogin,
			*cfg.WritePassword,
			"",
			"",
			"",
			"",
			"",
			*cfg.WriteFile)
	} else if *cfg.Operation == "delete" {
		oper.DeleteEntity(sessionOwner, *cfg.EntityName)
	} else if *cfg.Operation == "get" {
		oper.GetEntity(sessionOwner, *cfg.EntityName)
	} else if *cfg.Operation == "sync" {
		oper.Syncronize()
	} else {
		fmt.Println("Не опознана операция " + *cfg.Operation)
	}
	oper.SaveLastOperationDateTime()
}
