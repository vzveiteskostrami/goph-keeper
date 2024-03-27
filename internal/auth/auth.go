package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/vzveiteskostrami/goph-keeper/internal/adb"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
	"github.com/vzveiteskostrami/goph-keeper/internal/misc"
)

type ContextParamName string

var (
	CPuserID ContextParamName = "UserID"
)

func AuthHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID int64 = 0
		var until time.Time
		var ok bool

		cu, err := r.Cookie("token")

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else if userID, until, err = misc.GetUserData(cu.Value); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

		if err != nil {
			logging.S().Error(err)
			return
		}

		if time.Until(until) <= 0 {
			s := "время сессии истекло. Пройдите процедуру авторизации"
			err = errors.New(s)
			logging.S().Error(err)
			s = "В" + s[2:] + "."
			http.Error(w, s, http.StatusUnauthorized)
			return
		}

		ok, err = adb.UserIDExists(userID)
		if err != nil {
			logging.S().Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if !ok {
			s := "userId не найден в системе"
			err = errors.New(s)
			logging.S().Error(err)
			s = "U" + s[1:] + "."
			http.Error(w, s, http.StatusUnauthorized)
			return
		}

		c := context.WithValue(r.Context(), CPuserID, userID)

		next.ServeHTTP(w, r.WithContext(c))
	})
}
