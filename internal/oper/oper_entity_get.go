package oper

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

type localheader struct {
	caption string
	width   int
}

func GetEntity(owner string, name string) {
	delim := dialog.DrawHeader("Получить данные сущности", true)
	ok, err := misc.FileExists("ADM\\list")
	if err == nil {
		if !ok {
			fmt.Println("Список не создан.")
			return
		}
	}
	if err != nil {
		fmt.Println("Ошибка при проверке наличия списка:")
		fmt.Println(err)
		return
	}

	list, err := getEntityList()

	if err != nil {
		fmt.Println("Ошибка при чтении списка:")
		fmt.Println(err)
		return
	}

	if name == "" {
		name = dialog.GetAnswer("Введите имя нужной сущности:", false, false)
		if name == "-" {
			fmt.Println(otk)
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
		fmt.Println("Сущность с именем \"" + name + "\" не найдена.")
		fmt.Println(err)
		return
	}

	fmt.Println("Тип сущности:", typeToRussString(*list[n].Etype))
	fmt.Println("Создана:", *list[n].CreateDate)
	fmt.Print("Изменена:")
	if list[n].LocalDate == nil {
		fmt.Println(" ---")
	} else {
		fmt.Println(*list[n].LocalDate)
	}
	fmt.Print("Синхронизация:")
	if list[n].LocalDate == nil {
		fmt.Println(" ---")
	} else {
		fmt.Println(*list[n].LocalDate)
	}
	if list[n].Note != nil && *list[n].Note != "" {
		fmt.Println("Примечание:", *list[n].Note)
	}

	if *list[n].Etype == co.EntityLoginPassword {
		fmt.Println("Логин:", *list[n].Login)
		fmt.Println("Пароль:", *list[n].Password)
	} else if *list[n].Etype == co.EntityCard {
		fmt.Println("Номер:", *list[n].Number)
		fmt.Println("Действует до:", *list[n].Expired)
		fmt.Println("Владелец:", *list[n].Holder)
		fmt.Println("CVV:", *list[n].Cvv)
	} else if *list[n].Etype == co.EntityText {

	} else if *list[n].Etype == co.EntityBinary {
		fmt.Println("Директория:", *list[n].FilePath)
		key, err := misc.UnicKeyForExeDir()
		if err != nil {
			fmt.Println("Не удалось получить ключ. Ошибка:")
			fmt.Println(err.Error())
			return
		}
		fmt.Print("Достаём данные из архива...")
		fi, _, err := misc.ReadFromFileProtectedZIP_finfo("ADM\\"+*list[n].File, key)
		if err != nil {
			fmt.Println("\nНе удалось считать информацию. Ошибка:")
			fmt.Println(err.Error())
		} else {
			fmt.Println("\rИмя файла:", fi.FileName)
			fmt.Println("Размер:", misc.ByteCountIEC(fi.FileSize))
			fmt.Println(delim)
			y := dialog.Yn("Скачать файл из хранилища")
			if y {
				newfn, err := saveToDownload("ADM\\"+*list[n].File, key, fi.FileName)
				if err != nil {
					fmt.Println("Не удалось сохранить файл. Ошибка:")
					fmt.Println(err.Error())
				} else {
					newfn := misc.ExecPath() + "\\DOWNLOAD\\" + newfn
					fmt.Println("Файл сохранён в " + newfn)
				}
			}
		}
	}
	fmt.Println(delim)
}

func saveToDownload(file string, key string, sfile string) (string, error) {
	fn := filepath.Base(sfile)
	fn = "extracted." + time.Now().Format("20060102150405") + "." + fn
	err := misc.MakeDir("DOWNLOAD")
	if err != nil {
		return "", errors.New("Не удалось создать директорию загрузок. " + err.Error())
	}
	h, err := os.OpenFile("DOWNLOAD\\"+fn, os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}
	defer h.Close()
	_, err = misc.ReadFromFileProtectedZIP_fonly_w(file, key, h)
	if err != nil {
		return "", err
	}
	return fn, nil
}

func ShowEntityList(owner string, brief bool, etype int) {
	dialog.DrawHeader("Список хранимых сущностей", true)
	ok, err := misc.FileExists("ADM\\list")
	if err == nil {
		if !ok {
			fmt.Println("Список не создан.")
			return
		}
	}
	if err != nil {
		fmt.Println("Ошибка при проверке наличия списка:")
		fmt.Println(err)
		return
	}

	list, err := getEntityList()

	if err != nil {
		fmt.Println("Ошибка при чтении списка:")
		fmt.Println(err)
		return
	}

	header := []localheader{
		{caption: "Наименование", width: 12},
		{caption: "Тип", width: 3},
		{caption: "Дата создания", width: 16},
		{caption: "Дата изменения", width: 14},
		{caption: "Синхронизация", width: 13},
		{caption: "Файл", width: 36},
	}

	for _, en := range list {
		if *en.Owner == owner {
			n := rlen(*en.Name)
			if n > header[0].width {
				header[0].width = n
			}

			n = rlen(typeToRussString(*en.Etype))
			if n > header[1].width {
				header[1].width = n
			}

			if en.LocalDate != nil {
				header[3].width = 16
			}

			if en.ServerDate != nil {
				header[4].width = 16
			}

		}
	}

	fmt.Println("")
	for ii := 0; ii < len(header); ii++ {
		if brief && ii == 2 {
			break
		}
		fmt.Print(strings.Repeat("-", header[ii].width+3))
	}
	fmt.Println("")
	for ii := 0; ii < len(header); ii++ {
		if brief && ii == 2 {
			break
		}
		fmt.Printf("| %-*s ", header[ii].width, header[ii].caption)
	}
	fmt.Println("")
	for ii := 0; ii < len(header); ii++ {
		if brief && ii == 2 {
			break
		}
		fmt.Print(strings.Repeat("-", header[ii].width+3))
	}
	fmt.Println("")

	for _, en := range list {
		if *en.Owner == owner {
			if etype == co.EntityNotDefined || *en.Etype == etype {
				fmt.Printf("| %-*s ", header[0].width, *en.Name)
				fmt.Printf("| %-*s ", header[1].width, typeToRussString(*en.Etype))
				if !brief {
					dt := *en.CreateDate
					fmt.Printf("| %-*s ", header[2].width, dt.Format("02.01.2006 15:04"))
					if en.LocalDate != nil {
						dt := *en.LocalDate
						fmt.Printf("| %-*s ", header[3].width, dt.Format("02.01.2006 15:04"))
					} else {
						fmt.Printf("| %-*s ", header[3].width, "---")
					}
					if en.ServerDate != nil {
						dt := *en.ServerDate
						fmt.Printf("| %-*s ", header[4].width, dt.Format("02.01.2006 15:04"))
					} else {
						fmt.Printf("| %-*s ", header[4].width, "---")
					}
					fmt.Printf("| %-*s ", header[5].width, *en.File)
				}
				fmt.Println("")
			}
		}
	}
	for ii := 0; ii < len(header); ii++ {
		if brief && ii == 2 {
			break
		}
		fmt.Print(strings.Repeat("-", header[ii].width+3))
	}
	fmt.Println("")
}

func typeToRussString(etype int) string {
	if etype == co.EntityBinary {
		return "Бинарные данные"
	} else if etype == co.EntityText {
		return "Текстовые данные"
	} else if etype == co.EntityCard {
		return "Банковская карта"
	} else if etype == co.EntityLoginPassword {
		return "Логин/пароль"
	} else {
		return "Не определено"
	}
}

func rlen(s string) int {
	return len([]rune(s))
}
