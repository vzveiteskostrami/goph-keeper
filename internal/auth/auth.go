package auth

import (
	"context"
	"errors"
	"net/http"

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
		var ok bool

		cu, err := r.Cookie("token")

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else if userID, err = misc.GetUserData(cu.Value); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

		if err != nil {
			logging.S().Error(err)
			return
		}

		ok, err = adb.UserIDExists(userID)
		if err != nil {
			logging.S().Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if !ok {
			err = errors.New("userId не найден в системе")
			logging.S().Error(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		c := context.WithValue(r.Context(), CPuserID, userID)

		next.ServeHTTP(w, r.WithContext(c))
	})
}
