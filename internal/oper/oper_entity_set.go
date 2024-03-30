package oper

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

const otk string = "Отказ от ввода."

func NewEntity(owner string, etype int, name string, login string, password string, number string,
	expired string, holder string, cvv string, note string, file string) {

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

	} else if etype == co.EntityBinary {
		newEntityBinary(owner, name, file, note)
	}
}

func DeleteEntity(owner string, name string) {
	if name == "" {
		name = dialog.GetAnswer("Введите имя удаляемой сущности:", false, false)
		if name == "-" {
			fmt.Println(otk)
			return
		}
	}

	ok, err := entityExists(owner, name)
	if err != nil {
		fmt.Println("Во время проверки списка сущностей произошла ошибка:")
		fmt.Println(err)
		fmt.Println("Удаление сущности прекращёно.")
		return
	}

	if !ok {
		fmt.Println("Сущности с таким названием не существует.")
		return
	}

	list, err := getEntityList()
	if err != nil {
		fmt.Println("Во время получения списка сущностей произошла ошибка:")
		fmt.Println(err)
		fmt.Println("Удаление сущности прекращёно.")
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

	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Не удалось получить ключ. Ошибка:")
		fmt.Println(err)
		fmt.Println("Удаление сущности прекращёно.")
	}

	filename := *list[n].File
	ename := *list[n].Name
	list = append(list[:n], list[n+1:]...)

	b, err := json.Marshal(list)
	if err != nil {
		fmt.Println("Не удался marshal. Ошибка:")
		fmt.Println(err)
		fmt.Println("Удаление сущности прекращёно.")
		return
	}

	err = os.Remove("ADM\\" + filename)
	if err != nil {
		fmt.Println("Не удалось удалить файл хранения. Ошибка:")
		fmt.Println(err)
		fmt.Println("Удаление сущности прекращёно.")
		return
	}

	err = misc.SaveToFileProtectedZIP("ADM\\list", "list", key, b)
	if err != nil {
		fmt.Println("Не удалось сохранить list. Ошибка:")
		fmt.Println(err)
		fmt.Println("Но файл данных уже удалён.")
		return
	}

	fmt.Println("Сущность \"" + ename + "\" успешно удалена из локального хранилища.")
	fmt.Println("Для удаления с сервера проведите синхронизацию.")
}

func newEntityLoPa(owner string, name string, login string, password string, note string) {
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Ошибка получения ключа:")
		fmt.Println(err)
		fmt.Println("Сохранение невозможно.")
		return
	}

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
	data.CreateDate = &cd
	data.Login = &login
	data.Password = &password
	data.Name = &name
	data.Owner = &owner
	data.Note = &note

	fileName := uuid.New().String()
	data.File = &fileName

	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Неудачный marshal.")
		fmt.Println("Сохранение невозможно.")
		return
	}

	err = misc.SaveToFileProtectedZIP("ADM\\"+fileName, "lopa", key, b)
	if err != nil {
		fmt.Println("Ошибка при сохранении:")
		fmt.Println(err)
		return
	}
	err = addToList(data)
	if err != nil {
		fmt.Println("Ошибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("Данные успешно сохранены.")
}

func newEntityCard(owner string, name string, number string, expired string, holder string, cvv string, note string) {
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Ошибка получения ключа:")
		fmt.Println(err)
		fmt.Println("Сохранение невозможно.")
		return
	}

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
	data.CreateDate = &cd
	data.Expired = &expired
	data.Number = &number
	data.Holder = &holder
	data.Cvv = &cvv
	data.Name = &name
	data.Owner = &owner
	data.Note = &note

	fileName := uuid.New().String()
	data.File = &fileName

	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Неудачный marshal.")
		fmt.Println("Сохранение невозможно.")
		return
	}

	err = misc.SaveToFileProtectedZIP("ADM\\"+fileName, "lopa", key, b)
	if err != nil {
		fmt.Println("Ошибка при сохранении:")
		fmt.Println(err)
		return
	}
	err = addToList(data)
	if err != nil {
		fmt.Println("Ошибка при сохранении в список:")
		fmt.Println(err)
		return
	}
	fmt.Println("Данные успешно сохранены.")
}

func addToList(data entityData) error {
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		return errors.New("не удалось получить ключ. Ошибка: " + err.Error())
	}

	exists, err := misc.FileExists("ADM\\list")
	if err != nil {
		return errors.New("Не удалось проверить реестр. Ошибка:" + err.Error())
	}

	var list []entityData
	if exists {
		r, isHuck, err := misc.ReadFromFileProtectedZIP("ADM\\list", key)
		if isHuck || err != nil {
			// Если мы не смогли открыть с нашим паролем key, полученным по алгоритму,
			// то значит нам этот файл подложили снаружи.
			return errors.New(tinto + " " + err.Error())
		}
		err = json.Unmarshal(r, &list)
		if err != nil {
			return errors.New("не удалось преобразовать в json. Ошибка:" + err.Error())
		}
	}

	list = append(list, data)
	b, err := json.Marshal(list)
	if err != nil {
		return errors.New("Не удалось закодировать list. Ошибка:" + err.Error())
	}
	err = misc.SaveToFileProtectedZIP("ADM\\list", "list", key, b)

	return err
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
	data.Note = &note

	fileName := uuid.New().String()
	data.File = &fileName

	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Неудачный marshal.")
		fmt.Println("Сохранение невозможно.")
		return
	}

	h, err := os.OpenFile(file, os.O_RDONLY, 0777)
	if err != nil {
		fmt.Println("Ошибка открытия файла " + file + ":")
		fmt.Println(err)
		return
	}

	path := filepath.Dir(file)
	data.FilePath = &path
	fnm := filepath.Base(file)

	fmt.Print("Сохранение данных...")
	err = misc.SaveToFileProtectedZIP_f("ADM\\"+fileName, key, "bin", b, fnm, h)
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
