package oper

import (
	"fmt"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
)

func RenameEntity(owner string, name string, newname string) {
	if name == "" {
		name = dialog.GetAnswer("Введите имя переименовываемой сущности:", false, false)
		if name == "-" {
			fmt.Println(otk)
			return
		}
	}

	if newname == "" {
		newname = dialog.GetAnswer("Введите новое имя сущности:", false, false)
		if newname == "-" {
			fmt.Println(otk)
			return
		}
	}

	ok, err := entityExists(owner, name)
	if err != nil {
		fmt.Println("Во время проверки списка сущностей произошла ошибка:")
		fmt.Println(err)
		fmt.Println("Переименование сущности прекращено.")
		return
	}

	if !ok {
		fmt.Println("Сущности с таким именем не существует.")
		fmt.Println("Переименование сущности прекращено.")
		return
	}

	list, err := getEntityList()
	if err != nil {
		fmt.Println("Во время получения списка сущностей произошла ошибка:")
		fmt.Println(err)
		fmt.Println("Переименование сущности прекращено.")
		return
	}

	for _, en := range list {
		if *en.Owner == owner && *en.Name == newname {
			fmt.Println("Сущность с именем \"" + newname + "\" уже существует.")
			fmt.Println(err)
			fmt.Println("Переименование сущности прекращено.")
			return
		}
	}

	n := -1
	for q, en := range list {
		if *en.Owner == owner && *en.Name == name {
			n = q
			break
		}
	}

	if n == -1 {
		return
	}

	data := list[n]
	*data.Name = newname
	if data.LocalDate == nil {
		v := time.Now()
		data.LocalDate = &v
	} else {
		*data.LocalDate = time.Now()
	}
	if data.Updated == nil {
		u := true
		data.Updated = &u
	} else {
		*data.Updated = true
	}
	err = setInList(data, n)
	if err != nil {
		fmt.Println("Не удалось сохранить list. Ошибка:")
		fmt.Println(err)
		return
	}

	fmt.Println("Сущность \"" + name + "\" успешно переименована в \"" + newname + "\".")
	fmt.Println("Для переименования на сервере проведите синхронизацию.")
}
