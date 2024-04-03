package oper

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func SetEntity(owner string, name string, login string, password string, number string,
	expired string, holder string, cvv string, note string, file string, txt string) {

	if name == "" {
		name = dialog.GetAnswer("Введите имя изменяемой сущности:", false, false)
		if name == "-" {
			fmt.Println(otk)
			return
		}
	}

	ok, err := entityExists(owner, name)
	if err != nil {
		fmt.Println("Во время проверки списка сущностей произошла ошибка:")
		fmt.Println(err)
		fmt.Println("Изменение сущности прекращено.")
		return
	}

	if !ok {
		fmt.Println("Сущности с таким именем не существует.")
		fmt.Println("Изменение сущности прекращено.")
		return
	}

	list, err := getEntityList()
	if err != nil {
		fmt.Println("Во время получения списка сущностей произошла ошибка:")
		fmt.Println(err)
		fmt.Println("Изменение сущности прекращено.")
		return
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

	if note == "" {
		if list[n].Note != nil && *list[n].Note != "" {
			fmt.Println("Пояснение к сущности:")
			fmt.Println(*list[n].Note)
			q := dialog.Menu([]string{"Выберите действие", "Оставить как есть", "Удалить пояснение", "Ввести новое"})
			if q == 1 {
				note = *list[n].Note
			} else if q == 3 {
				note = dialog.GetAnswer("Введите пояснение к сущности, если требуется:", false, true)
				if note == "-" {
					fmt.Println(otk)
					return
				}
			}
		} else {
			note = dialog.GetAnswer("Старого пояснения нет. Введите пояснение к сущности, если требуется:", false, true)
			if note == "-" {
				fmt.Println(otk)
				return
			}
		}
	}

	if list[n].Updated == nil {
		u := true
		list[n].Updated = &u
	} else {
		*list[n].Updated = true
	}

	if *list[n].Etype == co.EntityLoginPassword {
		setEntityLoPa(list[n], n, login, password, note)
	} else if *list[n].Etype == co.EntityCard {
		setEntityCard(list[n], n, number, expired, holder, cvv, note)
	} else if *list[n].Etype == co.EntityText {
		setEntityText(list[n], n, txt, note)
	} else if *list[n].Etype == co.EntityBinary {
		setEntityBinary(list[n], n, file, note)
	}
}

func setEntityLoPa(data entityData, pos int, login string, password string, note string) {
	fmt.Println("Сохранённый логин:", *data.Login)
	if login == "" {
		login = dialog.GetAnswer(posta("Новый логин"), false, true)
		if login == "-" {
			fmt.Println(otk)
			return
		} else if login == "" {
			login = *data.Login
		}
	} else {
		fmt.Println("Новый логин:", login)
	}

	fmt.Println("Сохранённый пароль есть")
	if password == "" {
		password = dialog.GetAnswer(posta("Новый пароль"), true, true)
		if password == "-" {
			fmt.Println(otk)
			return
		} else if password == "" {
			password = *data.Password
		}
	} else {
		fmt.Println("Новый пароль указан")
	}

	cd := time.Now()
	data.LocalDate = &cd
	data.Login = &login
	data.Password = &password
	if note != "" {
		data.Note = &note
	} else {
		data.Note = nil
	}

	err := setInList(data, pos)
	if err != nil {
		fmt.Println("Ошибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("Данные успешно сохранены.")
}

func setEntityCard(data entityData, pos int, number string, expired string, holder string, cvv string, note string) {
	fmt.Println("Сохранённый номер карты:", *data.Number)
	if number == "" {
		number = dialog.GetAnswer(posta("Новый номер карты"), false, true)
		if number == "-" {
			fmt.Println(otk)
			return
		} else if number == "" {
			number = *data.Number
		}
	} else {
		fmt.Println("Новый номер карты:", number)
	}

	fmt.Println("Сохранённый срок:", *data.Expired)
	if expired == "" {
		expired = dialog.GetAnswer(posta("Новый срок NN/NN"), false, true)
		if expired == "-" {
			fmt.Println(otk)
			return
		} else if expired == "" {
			expired = *data.Expired
		}
	} else {
		fmt.Println("Новый срок:", expired)
	}

	fmt.Println("Сохранённый владелец:", *data.Holder)
	if holder == "" {
		holder = dialog.GetAnswer(posta("Новое имя владельца"), false, true)
		if holder == "-" {
			fmt.Println(otk)
			return
		} else if holder == "" {
			holder = *data.Holder
		}
	} else {
		fmt.Println("Новое имя владельца:", holder)
	}

	fmt.Println("Сохранённый CVV:", *data.Cvv)
	if cvv == "" {
		cvv = dialog.GetAnswer("Новый CVV:", false, true)
		if cvv == "-" {
			fmt.Println(otk)
			return
		} else if cvv == "" {
			cvv = *data.Cvv
		}
	} else {
		fmt.Println("Новый CVV:", cvv)
	}

	cd := time.Now()
	data.LocalDate = &cd
	data.Expired = &expired
	data.Number = &number
	data.Holder = &holder
	data.Cvv = &cvv
	if note != "" {
		data.Note = &note
	} else {
		data.Note = nil
	}

	err := setInList(data, pos)
	if err != nil {
		fmt.Println("Ошибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("Данные успешно сохранены.")
}

func setEntityBinary(data entityData, pos int, file string, note string) {
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Ошибка получения ключа:")
		fmt.Println(err)
		fmt.Println("Сохранение невозможно.")
		return
	}

	for {
		if file == "" {
			file = dialog.GetAnswer("Введите имя загружаемого файла:", false, false)
			if file == "-" {
				fmt.Println(otk)
				return
			}
		}

		ok, err := misc.FileExists(file)
		if err != nil {
			fmt.Println("Ошибка проверки наличия файла:")
			fmt.Println(err)
			fmt.Println("Сохранение невозможно.")
			return
		}

		if ok {
			break
		} else {
			fmt.Println("Файл не найден.")
			file = ""
		}
	}
	cd := time.Now()
	data.LocalDate = &cd
	if note != "" {
		data.Note = &note
	} else {
		data.Note = nil
	}

	h, err := os.OpenFile(file, os.O_RDONLY, 0777)
	if err != nil {
		fmt.Println("Ошибка открытия файла " + file + ":")
		fmt.Println(err)
		return
	}

	path := filepath.Dir(file)
	data.FilePath = &path
	savedFileName := filepath.Base(file)

	fmt.Print("Сохранение данных...")
	err = misc.SaveToFileProtectedZIP_r("ADM\\"+*data.File, key, savedFileName, h)
	h.Close()
	if err != nil {
		fmt.Println("\rОшибка при сохранении:")
		fmt.Println(err)
		return
	}

	err = setInList(data, pos)
	if err != nil {
		fmt.Println("\rОшибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("\rДанные успешно сохранены.")
}

func setEntityText(data entityData, pos int, text string, note string) {
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Ошибка получения ключа:")
		fmt.Println(err)
		fmt.Println("Сохранение невозможно.")
		return
	}

	if text == "" {
		text = dialog.GetAnswer("Введите новый сохраняемый текст:", false, false)
		if text == "-" {
			fmt.Println(otk)
			return
		}
	}

	cd := time.Now()
	data.LocalDate = &cd
	if note != "" {
		data.Note = &note
	} else {
		data.Note = nil
	}

	h := bytes.NewBuffer([]byte(text))

	fmt.Print("Сохранение данных...")
	err = misc.SaveToFileProtectedZIP_r("ADM\\"+*data.File, key, "text", h)
	if err != nil {
		fmt.Println("\rОшибка при сохранении:")
		fmt.Println(err)
		return
	}

	err = setInList(data, pos)
	if err != nil {
		fmt.Println("\rОшибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("\rДанные успешно сохранены.")
}

func posta(s string) string {
	return s + " (пусто - оставить старый):"
}
