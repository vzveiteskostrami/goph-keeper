package oper

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

const refuse string = "Отказ от выполнения операции"
const tinto string = "попытка взлома. Пароль не подошёл. Файл подложен Тинто Брассом"

//type currUser struct {
//	Login       string    `json:"login,omitempty"`
//	Expired     time.Time `json:"expired,omitempty"`
//	CurrentPath string    `json:"current_path,omitempty"`
//}

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

type localRegister struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
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

	var list []localRegister
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

	lr := localRegister{Login: login, Password: password}
	list = append(list, lr)
	b, err := json.Marshal(list)
	if err != nil {
		return errors.New("Не удалось закодировать list. Ошибка:" + err.Error())
	}
	err = misc.SaveToFileProtectedZIP("ADM\\register", "list", key, b)

	return err
}

/*
func fixCurrentUser(login string, until time.Time) {
	err := makeLocalDir("ADM")
	if err != nil {
		fmt.Println("Не удалось создать административную директорию. Ошибка:")
		fmt.Print(err)
		return
	}

	permissions := fs.FileMode(0644) // or whatever you need

	raw := new(bytes.Buffer)
	zipw := zp.NewWriter(raw)

	w, err := zipw.Encrypt("data", "golang")
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(w, bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	zipw.Close()
	os.WriteFile("file.zip", raw.Bytes(), permissions)

	if dir != "" {
		e, _ = exists("LOGI\\" + dir)
		if !e {
			err = os.Mkdir("LOGI\\"+dir, 0777)
			if err != nil {
				return err
			}
		}
		dir += "\\"
	}

}
*/

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func makeLocalDir(dirname string) error {
	e, _ := exists(dirname)
	if !e {
		err := os.Mkdir(dirname, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func ExecPath() string {
	var here = os.Args[0]
	here, err := filepath.Abs(here)
	if err != nil {
		return ""
	}
	return filepath.Dir(here)
}
