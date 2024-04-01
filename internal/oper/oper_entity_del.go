package oper

import (
	"fmt"
	"os"

	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
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

	err = os.Remove("ADM\\" + *list[n].File)
	if err != nil {
		fmt.Println("Не удалось удалить файл хранения. Ошибка:")
		fmt.Println(err)
		fmt.Println("Удаление сущности прекращено.")
		return
	}

	err = delFromList(n)
	if err != nil {
		fmt.Println("Не удалось удалить из списка. Ошибка:")
		fmt.Println(err)
		fmt.Println("Но файл данных уже удалён.")
		return
	}

	fmt.Println("Сущность \"" + *list[n].Name + "\" успешно удалена из локального хранилища.")
	fmt.Println("Для удаления с сервера проведите синхронизацию.")
}
