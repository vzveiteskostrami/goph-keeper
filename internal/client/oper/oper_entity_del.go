package oper

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

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
		fmt.Println("Удаление сущности прекращено.")
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
		fmt.Println("Удаление сущности прекращено.")
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

	if list[n].ServerID != nil {
		err = saveDeleteList(*list[n].ServerID)
		if err != nil {
			fmt.Println("Не удалось сохранить в список удалённых. Ошибка:")
			fmt.Println(err)
			fmt.Println("Удаление сущности прекращено.")
		}
	}

	if *list[n].Etype == co.EntityText || *list[n].Etype == co.EntityBinary {
		err = os.Remove("ADM\\" + *list[n].File)
		if err != nil {
			fmt.Println("Не удалось удалить файл хранения. Ошибка:")
			fmt.Println(err)
			fmt.Println("Удаление сущности прекращено.")
			return
		}
	}

	err = delFromList(n)
	if err != nil {
		fmt.Println("Не удалось удалить из списка. Ошибка:")
		fmt.Println(err)
		fmt.Println("Но файл данных уже удалён.")
		return
	}

	if list[n].ServerID != nil {
		fmt.Println("Сущность \"" + *list[n].Name + "\" успешно удалена из локального хранилища.")
		fmt.Println("Для удаления с сервера проведите синхронизацию.")

	} else {
		fmt.Println("Сущность \"" + *list[n].Name + "\" успешно удалена.")
	}
}

func saveDeleteList(oid int64) error {
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		return err
	}

	var aed []co.Udata
	var del []byte
	if ex, _ := misc.FileExists("ADM\\delete"); ex {
		del, _, err = misc.ReadFromFileProtectedZIP("ADM\\delete", key)
		if err != nil {
			return err
		}
		err = json.Unmarshal(del, &aed)
		if err != nil {
			return err
		}
	}

	ed := co.Udata{Oid: &oid}
	aed = append(aed, ed)

	del, err = json.Marshal(aed)
	if err != nil {
		return err
	}
	err = misc.SaveToFileProtectedZIP("ADM\\delete", "del", key, del)
	return err
}

func CloseSession() {
	os.Remove("ADM\token")
	os.Remove("ADM\\local_token")
	fmt.Println("Сессия закрыта.")
}
