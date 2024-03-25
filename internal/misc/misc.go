package misc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/golang-jwt/jwt/v4"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
)

const cryptoKey = "sxcbuascyghauyasuywqwbicqwugeygqwgueyqwgeuqwywnmxzcpoeqweyiqwenbnmbfghm"
const SecretKey = "pomidoryichesnok"

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

type RegInfo struct {
	Login    *string `json:"login,omitempty"`
	Password *string `json:"password,omitempty"`
}

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

func Hash256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func CheckLuhn(s string) bool {
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

func MakeToken(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserData(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		logging.S().Errorw(err.Error())
		return -1, err
	}

	if !token.Valid {
		logging.S().Errorw("Token is not valid: " + tokenString)
		return -1, errors.New("token is not valid")
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, nil
}

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

func StatusIntToStr(n *int16) (r string) {
	r = ""
	if n == nil {
		return
	}
	if *n == 0 {
		r = "NEW"
	} else if *n == 1 {
		r = "PROCESSING"
	} else if *n == 2 {
		r = "INVALID"
	} else if *n == 3 {
		r = "PROCESSED"
	}
	return
}

func StatusStrToInt(s string) (r int16) {
	r = -1
	if s == "NEW" {
		r = 0
	} else if s == "PROCESSING" {
		r = 1
	} else if s == "INVALID" {
		r = 2
	} else if s == "PROCESSED" {
		r = 3
	}
	return
}
