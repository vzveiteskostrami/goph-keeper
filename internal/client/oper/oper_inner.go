package oper

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

const refuse string = "Отказ от выполнения операции"
const tinto string = "попытка взлома. Пароль не подошёл. Файл подложен Тинто Брассом"
const otk string = "Отказ от ввода."

type authData struct {
	Login    string    `json:"login,omitempty"`
	Password string    `json:"password,omitempty"`
	Until    time.Time `json:"until,omitempty"`
}

type entityData struct {
	Owner      *string    `json:"owner,omitempty"`
	Etype      *int16     `json:"type,omitempty"`
	Name       *string    `json:"name,omitempty"`
	File       *string    `json:"file,omitempty"`
	Login      *string    `json:"login,omitempty"`
	Password   *string    `json:"password,omitempty"`
	Number     *string    `json:"number,omitempty"`
	Expired    *string    `json:"expired,omitempty"`
	Holder     *string    `json:"holder,omitempty"`
	Cvv        *string    `json:"cvv,omitempty"`
	Note       *string    `json:"note,omitempty"`
	FilePath   *string    `json:"file_name,omitempty"`
	CreateDate *time.Time `json:"create_date,omitempty"`
	LocalDate  *time.Time `json:"local_date,omitempty"`
	ServerDate *time.Time `json:"server_date,omitempty"`
	ServerID   *int64     `json:"server_id,omitempty"`
	Updated    *bool      `json:"updated,omitempty"`
}

func CheckLocalSession() (string, error) {
	ok, err := misc.FileExists("ADM\\local_token")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("Нет открытой локальной сессии.")
	}

	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		return "", err
	}
	raw, _, err := misc.ReadFromFileProtectedZIP("ADM\\local_token", key)
	if err != nil {
		return "", err
	}

	token, err := getAuthData(raw)
	if err != nil {
		return "", err
	}

	if len(token) == 0 {
		return "", errors.New("Токен пустой.")
	}

	if time.Until(token[0].Until) <= 0 {
		return "", errors.New("Время локальной сесии истекло. Авторизируйтесь.")
	}

	return token[0].Login, nil
}

func GetRegisterList() ([]authData, error) {
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		return []authData{}, err
	}
	raw, _, err := misc.ReadFromFileProtectedZIP("ADM\\register", key)
	if err != nil {
		return []authData{}, err
	}
	return getAuthData(raw)
}

// Проверка файла локального регистра на взлом.
func checkVaildRegisterFile(key string) error {
	exists, err := misc.FileExists("ADM\\register")
	if err != nil {
		return errors.New("Не удалось проверить наличие регистрационного файла. Ошибка: " + err.Error())
	}

	// Если регистрационного файла нет, это нормально. Может быть это первый заход.
	// Значит сейчас мы в процессе зарегиться в базе на серваке, и если эта регистрация пройдёт,
	// то всё чисто. Будем создавать новый регистрационный файл со всеми заморочками. Но
	// на это нам ответит регистрация на серваке. Пока не паримся.
	// Вернём ошибки нет.
	if !exists {
		return nil
	}

	// А вот если регистрационный файл уже есть, то может быть он хорош, а
	// может быть подложен нехорошими людьми (тинтобрассами). Если он хорош,
	// то распакуется с помощью key. А если нет, то значит он подложный, надо
	// сообщить об этом наверх, пусть наверху думают что с этим делать.
	// Скорее всего надо прерывать регистрацию.
	_, huck, err := misc.ReadFromFileProtectedZIP("ADM\\register", key)
	if err != nil {
		if huck {
			// Если мы не смогли открыть с нашим паролем key, полученным по алгоритму,
			// то значит нам этот файл подложили снаружи, чтобы попробовать посмотреть
			// чей-то аккаунт с новым паролем, который определили в новом регистрационном файле.
			// Ведь мы же не можем паковать ни с чем, кроме этого key.
			return errors.New(tinto + " " + err.Error())
		} else {
			return err
		}
	}
	return nil
}

func registerLocally(login string, password string) error {
	// Дисклеймер. Сюда мы должны попадать только при условии, что регистрация
	// на сервере прошла успешно, и до этого была проведена проверка на
	// подлинность файла register или его отсутствие. И то и другое состояние
	// легитимны для локальной регистрации логина и пароля. Файл регистрации
	// либо будет дополнен, либо создан.

	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		// Ситуация неопределённая. Неясно что делать программно. Надо разбираться руками
		// и башкой. Не может такого быть. Но если произошло, требует объяснения.
		return errors.New("не удалось получить ключ. Ошибка: " + err.Error())
	}

	exists, err := misc.FileExists("ADM\\register")
	if err != nil {
		// Тоже невероятная ситуация. Данная ф-ция вызывается только через ранее вызванную
		// checkVaildRegisterFile, и там эти ситуации уже обработаны. Но всё же бывает?
		return errors.New("Не удалось проверить наличие регистрационного файла. Ошибка:" + err.Error())
	}

	var list []authData
	if exists {
		r, isHuck, err := misc.ReadFromFileProtectedZIP("ADM\\register", key)
		if isHuck || err != nil {
			// Если мы не смогли открыть с нашим паролем key, полученным по алгоритму,
			// то значит нам этот файл подложили снаружи, чтобы попробовать посмотреть
			// чей-то аккаунт с новым паролем, который определили в новом регистрационном файле.
			// Ведь мы же не можем паковать ни с чем, кроме этого key.
			return errors.New(tinto + " " + err.Error())
		}
		err = json.Unmarshal(r, &list)
		if err != nil {
			return errors.New("не удалось преобразовать в json. Ошибка:" + err.Error())
		}
	}

	lr := authData{Login: login, Password: password}
	list = append(list, lr)
	b, err := json.Marshal(list)
	if err != nil {
		return errors.New("Не удалось закодировать list. Ошибка:" + err.Error())
	}
	err = misc.SaveToFileProtectedZIP("ADM\\register", "list", key, b)

	return err
}

func localAuthorization(login string, password string, place int) (bool, error) {
	if ex, _ := misc.FileExists("ADM\\register"); !ex && place == co.SessionBoth {
		key, err := misc.UnicKeyForExeDir()
		if err != nil {
			return false, err
		}

		list := []authData{{Login: login, Password: password}}
		b, err := json.Marshal(list)
		if err != nil {
			return false, errors.New("Не удалось закодировать list. Ошибка:" + err.Error())
		}
		err = misc.MakeDir("ADM")
		if err != nil {
			return false, errors.New("Не удалось создать ADM. Ошибка:" + err.Error())
		}
		err = misc.SaveToFileProtectedZIP("ADM\\register", "list", key, b)
		if err != nil {
			return false, errors.New("Не удалось сохранить list. Ошибка:" + err.Error())
		} else {
			return true, nil
		}
	}

	reg, err := GetRegisterList()
	if err != nil {
		return false, err
	}
	for _, pa := range reg {
		if pa.Login == login {
			if pa.Password == password {
				return true, nil
			}
			break
		}
	}
	return true, errors.New("Неверная пара логин/пароль.")
}

func saveLocalToken(login string, until time.Time) error {
	err := misc.MakeDir("ADM")
	if err != nil {
		return errors.New("Не удалось создать административную директорию. " + err.Error())
	}

	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		return errors.New("не удалось получить ключ. Ошибка: " + err.Error())
	}

	var token []authData
	tk := authData{Login: login, Until: until}
	token = append(token, tk)
	b, err := json.Marshal(token)
	if err != nil {
		return errors.New("Не удалось закодировать token. Ошибка:" + err.Error())
	}
	err = misc.SaveToFileProtectedZIP("ADM\\local_token", "token", key, b)

	return err
}

func getEntityList() ([]entityData, error) {
	reg := []entityData{}
	key, err := misc.UnicKeyForExeDir()
	if err != nil {
		return reg, err
	}
	raw, _, err := misc.ReadFromFileProtectedZIP("ADM\\list", key)
	if err != nil {
		return reg, err
	}
	err = json.Unmarshal(raw, &reg)
	return reg, err
}

func entityExists(owner string, name string) (bool, error) {
	ok, err := misc.FileExists("ADM\\list")
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	list, err := getEntityList()
	if err != nil {
		return false, err
	}
	for _, en := range list {
		if *en.Owner == owner {
			if *en.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func getAuthData(raw []byte) ([]authData, error) {
	var ada []authData
	err := json.Unmarshal(raw, &ada)
	return ada, err
}

func addToList(data entityData) error {
	return saveList(0, data, 0)
}

func delFromList(pos int) error {
	var data entityData
	return saveList(1, data, pos)
}

func setInList(data entityData, pos int) error {
	return saveList(2, data, pos)
}

func saveList(mode byte, data entityData, pos int) error {
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

	if mode == 0 {
		list = append(list, data)
	} else if mode == 1 {
		list = append(list[:pos], list[pos+1:]...)
	} else if mode == 2 {
		list[pos] = data
	}

	b, err := json.Marshal(list)
	if err != nil {
		return errors.New("Не удалось закодировать list. Ошибка:" + err.Error())
	}
	err = misc.SaveToFileProtectedZIP("ADM\\list", "list", key, b)

	return err
}

func typeToRussString(etype int16) string {
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
