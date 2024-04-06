package oper

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	zp "github.com/alexmullins/zip"
	"github.com/vzveiteskostrami/goph-keeper/internal/client/chttp"
	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func Syncronize(owner string, name string, strict int16) {
	if name == "" {
		strict = co.StrictNo
	}

	var conflict []entityData
	delim := dialog.DrawHeader("Синхронизация", false)
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println(delim)
		fmt.Println("Не удалось получить ключ. Ошибка:")
		fmt.Println(err.Error())
		return
	}

	wasActivity, err := makeLocalDel(key)
	if err != nil {
		return
	}

	a := false
	b := true
	sdata, _, err := chttp.GetEntityList(co.RequestList{Full: &a, All: &b, Data: nil})
	if err != nil {
		fmt.Println(delim)
		fmt.Println("Получение серверных данных. Ошибка:")
		fmt.Println(err.Error())
		return
	}

	var ldata []entityData
	if ex, _ := misc.FileExists("ADM\\list"); ex {
		ldata, err = getEntityList()
		if err != nil {
			fmt.Println(delim)
			fmt.Println("Получение локальных данных. Ошибка:")
			fmt.Println(err.Error())
			return
		}
	}

	// Создадим два массива каждый по длинне полученных данных.
	// В них будем хранить перекрёстную ссылку на данные другого
	// набора, чтобы не решать проблему прямо по месту.
	// Это сделает текст более "плоским" и проще читаемым.
	lindex := make([]int, len(ldata))
	sindex := make([]int, len(sdata))

	// Установка ссылок в обоих массивах
	for s := range sindex {
		if name != "" {
			if *sdata[s].DataName == name {
				sindex[s] = -1
			} else {
				sindex[s] = -4
			}
		} else {
			sindex[s] = -1
		}
	}
	// lindex[l] = -2 означает, что данная сущность не принадлежит
	// текущему владельцу.
	// -3 будет означать конфликт.
	// -4 - игнорирование.
	for l, ld := range ldata {
		lindex[l] = -2
		if *ld.Owner == owner {
			if name != "" && *ld.Name != name {
				lindex[l] = -4
			} else {
				lindex[l] = -1
				if ld.ServerID != nil {
					for s, sd := range sdata {
						if *sd.Oid == *ld.ServerID {
							sindex[s] = l
							lindex[l] = s
							break
						}
					}
				}
			}
		}
	}

	// Связь между элементами, пришедшими с сервера и локальными установлена.
	// Обрабатываем серверный массив. Тут может быть или допись в локалку,
	// либо перезапись локалки, либо конфликт.
	var dload []co.Udata
	for ii := 0; ii < len(sindex); ii++ {
		// Игнорирование
		if sindex[ii] == -4 {
			continue
		}
		// Пришло с сервера, на локале нет. Или выставлено строгое чтение с сервера. Надо добавлять.
		if sindex[ii] == -1 || strict == co.StrictRead {
			ud := co.Udata{Oid: sdata[ii].Oid}
			dload = append(dload, ud)
		} else {
			// Сравним даты синхронизации.
			dif := time.Duration(sdata[ii].UpdateTime.Sub(*ldata[sindex[ii]].ServerDate)) / time.Millisecond
			//Если серверная дата синхронизации больше, чем та, когда мы с сервера забирали
			// если мы не правили, то надо просто забрать.
			// А если правили, то это конфликт.
			if dif > time.Duration(0) {
				if ldata[sindex[ii]].Updated == nil || !*ldata[sindex[ii]].Updated {
					ud := co.Udata{Oid: sdata[ii].Oid}
					dload = append(dload, ud)
				} else {
					conflict = append(conflict, entityData{Name: ldata[sindex[ii]].Name,
						ServerDate: sdata[ii].UpdateTime,
						LocalDate:  ldata[sindex[ii]].ServerDate})
					lindex[sindex[ii]] = -3
				}
			}
		}
	}

	// Если с сервера надо скачать данные - качаем. Что поделать.
	if len(dload) > 0 {
		err = loadFromServer(owner, &dload, &ldata, key)
		if err != nil {
			fmt.Println("Не удалось считать с сервера. Ошибка:")
			fmt.Println(err.Error())
			return
		}
		// Сохранение промежуточного списка. На случай, если
		// во время дальнейшей синхронизации произойдёт ошибка.
		lda, err := json.Marshal(ldata)
		if err != nil {
			fmt.Println("Не удалось закодировать list. Ошибка:")
			fmt.Println(err.Error())
			return
		}
		err = misc.SaveToFileProtectedZIP("ADM\\list", "list", key, lda)
		if err != nil {
			fmt.Println("Ошибка сохранения list:")
			fmt.Println(err.Error())
		}
		wasActivity = true
	}

	var toDelete []entityData
	var toWrite []co.Udata
	printHeader := true
	// Обрабатываем локальный массив.
	// Локальный массив возвращается сразу на всех локальных пользователей.
	// Поэтому -2 записано в строках, которыми текущий пользователь не владеет.
	// А -3 записано в конфликтных строках.
	// Так мы быстрее проигнорируем ненужное.
	for ii := 0; ii < len(lindex); ii++ {
		// -3 конфликт, -4 игнорирование. И то и другое пропускаем.
		if lindex[ii] <= -2 {
			continue
		}
		// Нет на сервере, или выставлен флаг строго записывать. Надо добавлять/заменять/удалять.
		if lindex[ii] == -1 || strict == co.StrictWrite {
			if ldata[ii].ServerID != nil && strict != co.StrictWrite {
				// Запись на сервере удалена. Надо удалить и на локале
				// Только сразу удалить нельзя, массив разъедется.
				// Поэтому сохраним в список удаляемых и выполним удаление в
				// самом конце.
				tod := entityData{ServerID: ldata[ii].ServerID}
				toDelete = append(toDelete, tod)
			} else {
				if printHeader {
					dialog.DrawHeader("Выгрузка данных на сервер", true)
					printHeader = false
				}
				if lindex[ii] == -1 && ldata[ii].ServerID != nil {
					ldata[ii].ServerID = nil
				}
				tw, err := formingDataToWriteToServer(ldata[ii], key)
				if err != nil {
					fmt.Println("Формирование записи для сервера. Ошибка:")
					fmt.Println(err.Error())
					return
				}
				toWrite = append(toWrite, tw)
			}
		} else {
			// Сравним даты синхронизации.
			dif := ldata[ii].ServerDate.Sub(*sdata[lindex[ii]].UpdateTime) / time.Millisecond
			//Если даты равны, но мы правили у себя,то надо выложить новый
			// вариант. Знак меньше равно тут на всякий случай. В реальности
			// по технологии такого быть не может. Но чисто математически
			// мы закрыли все теоретические варианты состояния даты.
			// В серверном цикле мы закрыли >, тут <=
			if dif <= time.Duration(0) {
				if ldata[ii].Updated != nil && *ldata[ii].Updated {
					if printHeader {
						dialog.DrawHeader("Выгрузка данных на сервер", true)
						printHeader = false
					}
					var tw co.Udata
					tw, err = formingDataToWriteToServer(ldata[ii], key)
					if err != nil {
						fmt.Println("Формирование изменённой записи для сервера. Ошибка:")
						fmt.Println(err.Error())
						return
					}
					toWrite = append(toWrite, tw)
				}
			}
		}
	}

	// Запись текущих данных на сервер, а затем, после получения
	// ответа от сервера, изменения в локальном списке.
	if len(toWrite) > 0 {
		fmt.Print("Запись данных на сервер...")
		written, _, err := chttp.WriteEntityList(&toWrite)
		if err != nil {
			fmt.Println("\rЗапись данных на сервер. Ошибка:")
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("\rЗапись данных на сервер успешна. Записано %d сущностей.\r\n", len(toWrite))
		for ii := 0; ii < len(lindex); ii++ {
			if lindex[ii] == -2 {
				continue
			}
			for jj := 0; jj < len(written); jj++ {
				if *written[jj].DataName == *ldata[ii].Name {
					ldata[ii].ServerDate = written[jj].UpdateTime
					ldata[ii].ServerID = written[jj].Oid
					ldata[ii].Updated = nil
					break
				}
			}
		}
		wasActivity = true
	}

	if len(toDelete) > 0 {
		dialog.DrawHeader("Удаление сущностей, удалённых на сервере", true)
		for _, dl := range toDelete {
			for ii := 0; ii < len(ldata); ii++ {
				if ldata[ii].ServerID != nil && *dl.ServerID == *ldata[ii].ServerID {
					del := true
					name := *ldata[ii].Name
					fmt.Print("Наименование \"" + *ldata[ii].Name + "\", удаление...")
					if ldata[ii].File != nil {
						err = os.Remove("ADM\\" + *ldata[ii].File)
						if err != nil {
							fmt.Println("\rНе удалось удалить файл хранения. Ошибка:")
							fmt.Println(err)
							del = false
						}
					}
					if del {
						ldata = append(ldata[:ii], ldata[ii+1:]...)
						fmt.Println("\rНаименование \"" + name + "\", сущность удалена")
					}
					break
				}
			}
		}
		fmt.Printf("Удалено %d сущностей из локального хранилища.\r\n", len(toDelete))
		wasActivity = true
	}

	lda, err := json.Marshal(ldata)
	if err != nil {
		fmt.Println("Не удалось закодировать list. Ошибка:")
		fmt.Println(err.Error())
		return
	}
	err = misc.SaveToFileProtectedZIP("ADM\\list", "list", key, lda)
	if err != nil {
		fmt.Println("Ошибка сохранения list:")
		fmt.Println(err.Error())
	} else {
		if wasActivity {
			fmt.Println("Данные синхронизированы.")
		} else {
			fmt.Println(delim)
			fmt.Println("Данные идентичны. Синхронизация не потребовалась.")
		}
		if len(conflict) > 0 {
			delim = dialog.DrawHeader("ВНИМАНИЕ!! КОНФЛИКТЫ!", true)
			for _, c := range conflict {
				fmt.Println("Сущность \""+*c.Name+"\" обновлена на сервере", c.ServerDate, "что позже последнего обновления на этом устройстве", c.LocalDate)
			}
			fmt.Println(delim)
			fmt.Println("Разберитесь с конфликтной ситуацией.")

		}
	}
}

func formingDataToWriteToServer(ldata entityData, key string) (ud co.Udata, err error) {
	if ldata.ServerID != nil {
		ud.Oid = ldata.ServerID
	}
	ud.DataName = ldata.Name
	ud.CreateTime = ldata.CreateDate
	ud.DataType = ldata.Etype

	fmt.Print("Наименование \"" + *ldata.Name + "\", подготовка...")

	var file []byte
	filename := ""
	if *ldata.Etype == co.EntityBinary || *ldata.Etype == co.EntityText {
		if *ldata.Etype == co.EntityBinary {
			fi := misc.FileInfoFromZIP{}
			fi, err = misc.ReadFromFileProtectedZIP_file_info("ADM\\" + *ldata.File)
			if err != nil {
				fmt.Print("\r")
				return
			}
			filename = fi.FileName
		} else {
			filename = "text"
		}
		file, _, err = misc.ReadFromFileProtectedZIP("ADM\\"+*ldata.File, key)
		if err != nil {
			fmt.Print("\r")
			return
		}
	}

	// Обнулим данные, которые на сервер идут через Udata или просто не
	// должны идти на сервер.
	ldata.ServerID = nil
	ldata.Name = nil
	ldata.CreateDate = nil
	etype := *ldata.Etype
	ldata.Etype = nil
	ldata.ServerDate = nil
	ldata.LocalDate = nil
	ldata.Owner = nil
	ldata.Updated = nil

	header, err := json.Marshal(ldata)
	if err != nil {
		fmt.Print("\r")
		return
	}

	raw := new(bytes.Buffer)
	zipw := zp.NewWriter(raw)

	w, err := zipw.Encrypt("header", co.ServerKey)
	if err != nil {
		fmt.Print("\r")
		return
	}
	_, err = io.Copy(w, bytes.NewReader(header))
	if err != nil {
		fmt.Print("\r")
		return
	}
	if etype == co.EntityBinary || etype == co.EntityText {
		w, err = zipw.Encrypt(filename, co.ServerKey)
		if err != nil {
			fmt.Print("\r")
			return
		}
		_, err = io.Copy(w, bytes.NewReader(file))
		if err != nil {
			fmt.Print("\r")
			return
		}
	}
	zipw.Close()

	zs := base64.StdEncoding.EncodeToString(raw.Bytes())
	ud.Data = &zs
	fmt.Println("\rНаименование \"" + *ud.DataName + "\", данные сформированы")
	return
}

func loadFromServer(owner string, dload *[]co.Udata, ldata *[]entityData, key string) error {
	dialog.DrawHeader("Загрузка данных с сервера", true)
	a := true
	b := false
	sdata, _, err := chttp.GetEntityList(co.RequestList{Full: &a, All: &b, Data: *dload})
	if err != nil {
		return err
	}

	for _, d := range sdata {
		ed := entityData{ServerID: d.Oid,
			ServerDate: d.UpdateTime,
			Name:       d.DataName,
			Owner:      &owner,
			Etype:      d.DataType,
			CreateDate: d.CreateTime,
		}

		r, err := base64.StdEncoding.DecodeString(*d.Data)
		if err != nil {
			return err
		}

		zipr, err := zp.NewReader(bytes.NewReader(r), int64(len(r)))
		if err != nil {
			return err
		}

		for n, z := range zipr.File {
			if n == 0 {
				fmt.Print("Чтение сущности...")
			} else {
				fmt.Print("Чтение файла...")
			}
			z.SetPassword(co.ServerKey)
			rr, err := z.Open()
			if err != nil {
				return err
			}
			b, _ := io.ReadAll(rr)
			rr.Close()
			if n == 0 {
				fmt.Println("\rНаименование: \"" + *ed.Name + "\"")
				e := entityData{}
				err = json.Unmarshal(b, &e)
				if err != nil {
					fmt.Println("")
					return err
				}
				ed.Note = e.Note
				if *d.DataType == co.EntityLoginPassword {
					ed.Login = e.Login
					ed.Password = e.Password
				} else if *d.DataType == co.EntityCard {
					ed.Cvv = e.Cvv
					ed.Expired = e.Expired
					ed.Holder = e.Holder
					ed.Number = e.Number
				} else if *d.DataType == co.EntityBinary {
					ed.FilePath = e.FilePath
					ed.File = e.File
				} else if *d.DataType == co.EntityText {
					ed.File = e.File
				}
				if *d.DataType == co.EntityLoginPassword || *d.DataType == co.EntityCard {
					break
				}
			} else {
				err = misc.SaveToFileProtectedZIP("ADM\\"+*ed.File, z.FileInfo().Name(), key, b)
				fmt.Println("\rФайл: " + *ed.File)
				if err != nil {
					fmt.Println("")
					return err
				}
			}
		}

		app := true
		for ii, d := range *ldata {
			if d.ServerID != nil && *d.ServerID == *ed.ServerID {
				[]entityData(*ldata)[ii] = ed
				app = false
				break
			}
		}

		if app {
			*ldata = append(*ldata, ed)
		}
	}
	fmt.Printf("С сервера считано %d сущностей.\n\r", len(sdata))

	return nil
}

func makeLocalDel(key string) (bool, error) {
	if ex, _ := misc.FileExists("ADM\\delete"); !ex {
		return false, nil
	}

	dialog.DrawHeader("Удаление данных на сервере", true)

	arr, _, err := misc.ReadFromFileProtectedZIP("ADM\\delete", key)
	if err != nil {
		return false, err
	}

	var del []co.Udata
	err = json.Unmarshal(arr, &del)

	if len(del) == 0 {
		return false, nil
	}

	fmt.Print("Удаление...")
	_, err = chttp.DeleteEntityList(&del)

	if err == nil {
		fmt.Printf("\rНа сервере удалено %d сущностей.\r\n", len(del))
		os.Remove("ADM\\delete")
	} else {
		fmt.Println("\rНе удалось выполнить удаление на сервере. Ошибка:")
		fmt.Println(err.Error())
	}
	return true, err
}
