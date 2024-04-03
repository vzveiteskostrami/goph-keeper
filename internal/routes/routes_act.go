package routes

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vzveiteskostrami/goph-keeper/internal/adb"
	"github.com/vzveiteskostrami/goph-keeper/internal/auth"
	"github.com/vzveiteskostrami/goph-keeper/internal/co"
	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
)

func UserDataWritef(w http.ResponseWriter, r *http.Request) {
	var data []co.Udata
	var err error
	if err = json.NewDecoder(r.Body).Decode(&data); err != nil {
		logging.S().Infow("Ошибка парсинга json " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(auth.CPuserID).(int64)
	completed := make(chan struct{})

	var newdata []co.Udata

	go func() {
		newdata, err = adb.WriteUserDataList(r.Context(), userID, data)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			b, _ := json.Marshal(newdata)
			w.Write(b)
		}
	case <-r.Context().Done():
		logging.S().Infow("Запись данных прервана на клиентской стороне.")
		w.WriteHeader(http.StatusGone)
	}
}

func Syncf(w http.ResponseWriter, r *http.Request) {
	/*body*/ _, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
