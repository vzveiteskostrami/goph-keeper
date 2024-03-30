package misc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	zp "github.com/alexmullins/zip"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sandipmavani/hardwareid"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
)

const cryptoKey = "sxcbuascyghauyasuywqwbicqwugeygqwgueyqwgeuqwywnmxzcpoeqweyiqwenbnmbfghm"
const SecretKey = "pomidoryichesnok2bananailavaSH_vash"

// Структура JWT токена. Содержит идентификатрор пользователя
// и дату/время, до которого токен валиден.
type Claims struct {
	jwt.RegisteredClaims
	UserID int64
	Until  time.Time
}

type FileInfoFromZIP struct {
	FileName string
	FileSize int64
}

var (
	unicKeyForDir string
)

// Структура регистрации/авторизации.
type RegInfo struct {
	Login           *string `json:"login,omitempty"`
	Password        *string `json:"password,omitempty"`
	SessionDuration *int64  `json:"session_duration,omitempty"`
}

// Шифрование необходимой строки
func Shifro(s string) (string, error) {
	key := makeKey() // ключ шифрования

	// NewCipher создает и возвращает новый cipher.Block.
	// Ключевым аргументом должен быть ключ AES, 16, 24 или 32 байта
	// для выбора AES-128, AES-192 или AES-256.
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// NewGCM возвращает заданный 128-битный блочный шифр
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	// создаём вектор инициализации
	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(aesgcm.Seal(nonce, nonce, []byte(s), nil)), nil
}

// Дешифрование необходимой строки
func DeShifro(s string) (string, error) {
	key := makeKey()
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	encrypted, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}

	nonce, encrypted := encrypted[:aesgcm.NonceSize()], encrypted[aesgcm.NonceSize():]

	// расшифровываем
	decrypted, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// Создание хэша
func Hash256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// Создание JWT токена для пользователя.
func MakeToken(userID int64, until time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		Until:  until,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Чтения данных пользователя из JWT токена.
func GetUserData(tokenString string) (int64, time.Time, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		logging.S().Errorw(err.Error())
		return -1, time.Now(), err
	}

	if !token.Valid {
		logging.S().Errorw("Token is not valid: " + tokenString)
		return -1, time.Now(), errors.New("token is not valid")
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, claims.Until, nil
}

// Формирование структуры с регистрационной/авторизационной информацией
// из входящаго потока.
func ExtractRegInfo(r io.Reader) (RegInfo, error) {
	var regIn RegInfo
	if err := json.NewDecoder(r).Decode(&regIn); err != nil {
		return regIn, err
	}
	if regIn.Login == nil || regIn.Password == nil || *regIn.Login == "" || *regIn.Password == "" {
		return regIn, errors.New("не указан логин/пароль")
	}
	return regIn, nil
}

// Формирование уникального ключа для упаковки данных с привязкой к месту.
// Адрес BIOS (и тд, если не получилось считать) привязка к машине,
// директория - привязка к конкретному месту на машине, и немного
// произвольной информации для догадавшихся про привязку к месту и
// желающих сломать файл вручную.
func UnicKeyForExeDir() (string, error) {
	if unicKeyForDir == "" {
		b, err := hardwareid.GetBIOSSerialNumber()
		if err != nil || b == "" {
			b, err = hardwareid.GetDiskDriverSerialNumber()
			if err != nil || b == "" {
				b, err = hardwareid.GetCPUPorcessorID()
				if err != nil || b == "" {
					b, err = hardwareid.ID()
					if err != nil || b == "" {
						b, err = hardwareid.GetPhysicalId()
						if err != nil {
							return "", errors.New("Ошибка получения id для key: " + err.Error())
						}
					}
				}
			}
		}

		b = strings.ReplaceAll(b, "\n", "")
		b = strings.ReplaceAll(b, "\r", "")
		b = strings.ReplaceAll(b, " ", "")
		b = strings.ReplaceAll(b, ":", "q")

		here, err := os.Executable()
		//here := os.Args[0]
		if err != nil {
			return "", errors.New("Ошибка Executable: " + err.Error())
		}
		here, err = filepath.Abs(here)
		if err != nil {
			return "", errors.New("Ошибка ABS: " + err.Error())
		}
		here = filepath.Dir(here)
		here = strings.ReplaceAll(here, ":", "!")
		here = strings.ReplaceAll(here, "\\", "_")
		here = strings.ReplaceAll(here, "/", "z")
		here = strings.ReplaceAll(here, " ", "2")

		here = strings.ToLower(here)
		unicKeyForDir = b + "mz29sp4kFrOAzxcR58sqj" + here
	}

	return unicKeyForDir, nil
}

// Сохранение информации в ZIP файл с паролем.
func SaveToFileProtectedZIP(fname string, iname string, key string, b []byte) error {
	permissions := fs.FileMode(0644)

	raw := new(bytes.Buffer)
	zipw := zp.NewWriter(raw)

	w, err := zipw.Encrypt(iname, key)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewReader(b))
	if err != nil {
		return err
	}
	zipw.Close()
	err = os.WriteFile(fname, raw.Bytes(), permissions)
	return err
}

// Сохранение информации в ZIP файл с паролем.
func SaveToFileProtectedZIP_f(fname string, key string, dname string, header []byte, fdname string, r io.Reader) error {
	permissions := fs.FileMode(0644)

	raw := new(bytes.Buffer)
	zipw := zp.NewWriter(raw)

	w, err := zipw.Encrypt(dname, key)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, bytes.NewReader(header))
	if err != nil {
		return err
	}

	f, err := zipw.Encrypt(fdname, key)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	zipw.Close()
	err = os.WriteFile(fname, raw.Bytes(), permissions)
	return err
}

func ReadFromFileProtectedZIP(fname string, key string) ([]byte, bool, error) {
	var b []byte
	f, err := os.Open(fname)
	if err != nil {
		return b, false, err
	}
	r, err := io.ReadAll(f)
	if err != nil {
		return b, false, err
	}

	zipr, err := zp.NewReader(bytes.NewReader(r), int64(len(r)))
	if err != nil {
		return b, false, err
	}

	for _, z := range zipr.File {
		z.SetPassword(key)
		rr, err := z.Open()
		if err != nil {
			// Если мы не смогли открыть с нашим паролем key, полученным по алгоритму,
			// то значит нам этот файл подложили снаружи, чтобы попробовать посмотреть
			// чей-то аккаунт с новым паролем, который определили в новом регистрационном файле.
			// Ведь мы же не можем паковать ни с чем, кроме этого key.
			return b, true, err
		}
		b, _ = io.ReadAll(rr)
		rr.Close()
	}
	return b, false, nil
}

func ReadFromFileProtectedZIP_finfo(fname string, key string) (FileInfoFromZIP, bool, error) {
	var fi FileInfoFromZIP
	f, err := os.Open(fname)
	if err != nil {
		return fi, false, err
	}
	r, err := io.ReadAll(f)
	if err != nil {
		return fi, false, err
	}

	zipr, err := zp.NewReader(bytes.NewReader(r), int64(len(r)))
	if err != nil {
		return fi, false, err
	}

	for a, z := range zipr.File {
		if a == 1 {
			z.SetPassword(key)
			fi.FileName = z.FileHeader.FileInfo().Name()
			fi.FileSize = z.FileHeader.FileInfo().Size()
			//rr, err := z.Open()
			//if err != nil {
			// Если мы не смогли открыть с нашим паролем key, полученным по алгоритму,
			// то значит нам этот файл подложили снаружи, чтобы попробовать посмотреть
			// чей-то аккаунт с новым паролем, который определили в новом регистрационном файле.
			// Ведь мы же не можем паковать ни с чем, кроме этого key.
			//	return fi, true, err
			//}
			//b, _ = io.ReadAll(rr)
			//rr.Close()
		}
	}
	return fi, false, nil
}

func ReadFromFileProtectedZIP_fonly_w(fname string, key string, w io.Writer) (bool, error) {
	f, err := os.Open(fname)
	if err != nil {
		return false, err
	}
	r, err := io.ReadAll(f)
	if err != nil {
		return false, err
	}

	zipr, err := zp.NewReader(bytes.NewReader(r), int64(len(r)))
	if err != nil {
		return false, err
	}

	for a, z := range zipr.File {
		if a == 1 {
			z.SetPassword(key)
			rr, err := z.Open()
			if err != nil {
				// Если мы не смогли открыть с нашим паролем key, полученным по алгоритму,
				// то значит нам этот файл подложили снаружи, чтобы попробовать посмотреть
				// чей-то аккаунт с новым паролем, который определили в новом регистрационном файле.
				// Ведь мы же не можем паковать ни с чем, кроме этого key.
				return true, err
			}
			_, err = io.Copy(w, rr)
			rr.Close()
			if err != nil {
				return true, err
			}
		}
	}
	return false, nil
}

// Полный путь до исполняемого файла
func ExecPath() string {
	var here string
	here, err := os.Executable()
	if err != nil {
		return ""
	}
	here, err = filepath.Abs(here)
	if err != nil {
		return ""
	}
	return filepath.Dir(here)
}

// Проверка наличия файла
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Создание директории, если её нет
func MakeDir(dirname string) error {
	e, _ := Exists(dirname)
	if !e {
		err := os.Mkdir(dirname, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

// Проверка наличия/отсутствия файла по определённому маршруту
func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

// Проверка любого типа переменной на nil.
func IsNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	objValue := reflect.ValueOf(obj)
	// проверяем, что тип значения ссылочный, то есть в принципе может быть равен nil
	if objValue.Kind() != reflect.Ptr {
		return false
	}
	// проверяем, что значение равно nil
	//  важно, что IsNil() вызывает панику, если value не является ссылочным типом. Поэтому всегда проверяйте на Kind()
	if objValue.IsNil() {
		return true
	}

	return false
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func makeKey() []byte {
	return []byte(cryptoKey)[:32]
}

func generateRandom(size int) ([]byte, error) {
	// генерируем криптостойкие случайные байты в b
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

/*
func checkLuhn(s string) bool {
	sum := 0
	nDigits := len(s)
	parity := nDigits % 2
	arr := []byte(s)
	for ii := 0; ii < nDigits; ii++ {
		if arr[ii] < 48 || arr[ii] > 57 {
			return false
		}
		digit := int(arr[ii] - 48)
		if ii%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return (sum % 10) == 0
}
*/
