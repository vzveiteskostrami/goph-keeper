package oper

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func NewEntity(owner string, etype int16, name string, login string, password string, number string,
	expired string, holder string, cvv string, note string, file string, txt string) {

	if name == "" {
		name = dialog.GetAnswer("Введите имя новой сущности:", false, false)
		if name == "-" {
			fmt.Println(otk)
			return
		}
	}

	ok, err := entityExists(owner, name)
	if err != nil {
		fmt.Println("Во время проверки списка сущностей произошла ошибка:")
		fmt.Println(err)
		fmt.Println("Ввод новой сущности прекращён.")
		return
	}

	if ok {
		fmt.Println("Сущность с таким названием уже существует.")
		fmt.Println("Ввод новой сущности прекращён.")
		return
	}

	if etype == co.EntityNotDefined {
		et := dialog.Menu([]string{"Выберите тип новой сущности:", "Пара логин/пароль", "Данные банковской карты",
			"Текстовая информация", "Бинарная информация (Файл)"})
		if et == 0 {
			fmt.Println(otk)
			return
		}
		etype = et
	}

	if note == "" {
		note = dialog.GetAnswer("Введите пояснения к сущности, если требуется:", false, true)
		if note == "-" {
			fmt.Println(otk)
			return
		}
	}

	if etype == co.EntityLoginPassword {
		newEntityLoPa(owner, name, login, password, note)
	} else if etype == co.EntityCard {
		newEntityCard(owner, name, number, expired, holder, cvv, note)
	} else if etype == co.EntityText {
		newEntityText(owner, name, txt, note)
	} else if etype == co.EntityBinary {
		newEntityBinary(owner, name, file, note)
	}
}

func newEntityLoPa(owner string, name string, login string, password string, note string) {
	if login == "" {
		login = dialog.GetAnswer("Введите сохраняемый логин:", false, false)
		if login == "-" {
			fmt.Println(otk)
			return
		}
	}

	if password == "" {
		password = dialog.GetAnswer("Введите сохраняемый пароль:", true, false)
		if password == "-" {
			fmt.Println(otk)
			return
		}
	}

	cd := time.Now()
	et := co.EntityLoginPassword

	var data entityData
	data.Etype = &et
	data.Name = &name
	data.Owner = &owner
	data.CreateDate = &cd
	data.Login = &login
	data.Password = &password
	if note != "" {
		data.Note = &note
	}

	err := addToList(data)
	if err != nil {
		fmt.Println("Ошибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("Данные успешно сохранены.")
}

func newEntityCard(owner string, name string, number string, expired string, holder string, cvv string, note string) {
	if number == "" {
		number = dialog.GetAnswer("Введите номер карты:", false, false)
		if number == "-" {
			fmt.Println(otk)
			return
		}
	}

	if expired == "" {
		expired = dialog.GetAnswer("Введите срок NN/NN:", false, false)
		if expired == "-" {
			fmt.Println(otk)
			return
		}
	}

	if holder == "" {
		holder = dialog.GetAnswer("Имя владельца:", false, false)
		if holder == "-" {
			fmt.Println(otk)
			return
		}
	}

	if cvv == "" {
		cvv = dialog.GetAnswer("CVV:", false, false)
		if cvv == "-" {
			fmt.Println(otk)
			return
		}
	}

	cd := time.Now()
	et := co.EntityCard

	var data entityData
	data.Etype = &et
	data.Name = &name
	data.Owner = &owner
	data.CreateDate = &cd
	data.Expired = &expired
	data.Number = &number
	data.Holder = &holder
	data.Cvv = &cvv
	if note != "" {
		data.Note = &note
	}

	err := addToList(data)
	if err != nil {
		fmt.Println("Ошибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("Данные успешно сохранены.")
}

func newEntityBinary(owner string, name string, file string, note string) {
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
	et := co.EntityBinary

	var data entityData
	data.Etype = &et
	data.CreateDate = &cd
	data.Name = &name
	data.Owner = &owner
	if note != "" {
		data.Note = &note
	}

	fileName := uuid.New().String()
	data.File = &fileName

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
	err = misc.SaveToFileProtectedZIP_r("ADM\\"+fileName, key, savedFileName, h)
	h.Close()
	if err != nil {
		fmt.Println("\rОшибка при сохранении:")
		fmt.Println(err)
		return
	}

	err = addToList(data)
	if err != nil {
		fmt.Println("\rОшибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("\rДанные успешно сохранены.")
}

func newEntityText(owner string, name string, text string, note string) {
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Ошибка получения ключа:")
		fmt.Println(err)
		fmt.Println("Сохранение невозможно.")
		return
	}

	if text == "" {
		text = dialog.GetAnswer("Введите сохраняемый текст:", false, false)
		if text == "-" {
			fmt.Println(otk)
			return
		}
	}

	cd := time.Now()
	et := co.EntityText

	var data entityData
	data.Etype = &et
	data.CreateDate = &cd
	data.Name = &name
	data.Owner = &owner
	if note != "" {
		data.Note = &note
	}

	fileName := uuid.New().String()
	data.File = &fileName

	h := bytes.NewBuffer([]byte(text))

	fmt.Print("Сохранение данных...")
	err = misc.SaveToFileProtectedZIP_r("ADM\\"+fileName, key, "text", h)
	if err != nil {
		fmt.Println("\rОшибка при сохранении:")
		fmt.Println(err)
		return
	}

	err = addToList(data)
	if err != nil {
		fmt.Println("\rОшибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("\rДанные успешно сохранены.")
}
