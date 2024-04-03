package oper

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	zp "github.com/alexmullins/zip"
	"github.com/vzveiteskostrami/goph-keeper/internal/chttp"
	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/dialog"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

func Syncronize(owner string) {
	dialog.DrawHeader("Синхронизация", true)
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		fmt.Println("Не удалось получить ключ. Ошибка:")
		fmt.Println(err.Error())
		return
	}

	a := false
	b := true
	sdata, _, err := chttp.GetList(co.RequestList{Full: &a, All: &b, Data: nil})
	if err != nil {
		fmt.Println("Получение серверных данных. Ошибка:")
		fmt.Println(err.Error())
		return
	}

	ldata, err := getEntityList()
	if err != nil {
		fmt.Println("Получение локальных данных. Ошибка:")
		fmt.Println(err.Error())
		return
	}

	// Создадим два массива каждый по длинне полученных данных.
	// В них будем хранить перекрёстную ссылку на данные другого
	// набора, чтобы не решать проблему прямо по месту.
	// Это сделает текст более "плоским" и проще читаемым.
	lindex := make([]int, len(ldata))
	sindex := make([]int, len(sdata))

	// Установка ссылок в обоих массивах
	for s := range sindex {
		sindex[s] = -1
	}
	// lindex[l] = -2 означает, что данная сущность не принадлежит
	// текущему владельцу.
	for l, ld := range ldata {
		lindex[l] = -2
		if *ld.Owner == owner {
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

	// Связь между элементами, пришедшими с сервера и локальными установлена.
	// Обрабатываем серверный массив. Тут может быть или допись в локалку,
	// либо перезапись локалки, либо конфликт.
	for ii := 0; ii < len(sindex); ii++ {
		// Пришло с сервера, на локале нет. Надо добавлять.
		if sindex[ii] == -1 {
			fmt.Println("Добавление")
		} else {
			// Сравним даты синхронизации.
			dif := time.Duration(sdata[ii].UpdateTime.Sub(*ldata[sindex[ii]].ServerDate)) / time.Millisecond
			//Если серверная дата синхронизации больше, чем та, когда мы с сервера забирали
			// если мы не правили, то надо просто забрать.
			// А если правили, то это конфликт.
			if dif > time.Duration(0) {
				if ldata[sindex[ii]].Updated == nil || !*ldata[sindex[ii]].Updated {
					fmt.Println("Перезапись")
				} else {
					fmt.Println("Плохо. Конфликт.")
				}
			}
		}
	}

	var toWrite []co.Udata
	// Обрабатываем локальный массив.
	// Локальный массив возвращается сразу на всех локальных пользователей.
	// Поэтому -2 записано в строках, которыми текущий пользователь не владеет.
	// Так мы быстрее проигнорируем ненужное.
	for ii := 0; ii < len(lindex); ii++ {
		if lindex[ii] == -2 {
			continue
		}
		// Нет на сервере. Надо добавлять.
		if lindex[ii] == -1 {
			tw, err := formingDataToWriteToServer(ldata[ii], key)
			if err != nil {
				fmt.Println("Формирование новой записи для сервера. Ошибка:")
				fmt.Println(err.Error())
				return
			}
			toWrite = append(toWrite, tw)
		} else {
			// Сравним даты синхронизации.
			dif := ldata[ii].ServerDate.Sub(*sdata[lindex[ii]].UpdateTime) / time.Millisecond
			//Если даты равны, но мы правили у себя,то надо выложить новый
			// вариант. Знак меньше равно тут на всякий случай. В реальности
			// по технологии такого быть не может. Но чисто математически
			// мы закрыли все теоретические варианты состояния даты.
			// В серверном цикле мы закрыли >, тут <=
			if dif <= time.Duration(0) {
				if ldata[sindex[ii]].Updated != nil && *ldata[sindex[ii]].Updated {
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
		written, _, err := chttp.WriteList(&toWrite)
		if err != nil {
			fmt.Println("Запись данных на сервер. Ошибка:")
			fmt.Println(err.Error())
			return
		}
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
		fmt.Println("Данные синхронизированы.")
	}
}

func formingDataToWriteToServer(ldata entityData, key string) (ud co.Udata, err error) {
	if ldata.ServerID != nil {
		ud.Oid = ldata.ServerID
	}
	ud.DataName = ldata.Name
	ud.CreateTime = ldata.CreateDate
	ud.DataType = ldata.Etype

	var file []byte
	filename := ""
	if *ldata.Etype == co.EntityBinary || *ldata.Etype == co.EntityText {
		if *ldata.Etype == co.EntityBinary {
			fi := misc.FileInfoFromZIP{}
			fi, err = misc.ReadFromFileProtectedZIP_file_info("ADM\\" + *ldata.File)
			if err != nil {
				return
			}
			filename = fi.FileName
		} else {
			filename = "text"
		}
		file, _, err = misc.ReadFromFileProtectedZIP("ADM\\"+*ldata.File, key)
		if err != nil {
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
		return
	}

	raw := new(bytes.Buffer)
	zipw := zp.NewWriter(raw)

	w, err := zipw.Encrypt("header", co.ServerKey)
	if err != nil {
		return
	}
	_, err = io.Copy(w, bytes.NewReader(header))
	if err != nil {
		return
	}
	if etype == co.EntityBinary || etype == co.EntityText {
		w, err = zipw.Encrypt(filename, co.ServerKey)
		if err != nil {
			return
		}
		_, err = io.Copy(w, bytes.NewReader(file))
		if err != nil {
			return
		}
	}
	zipw.Close()

	zs := base64.StdEncoding.EncodeToString(raw.Bytes())
	ud.Data = &zs
	return
}
