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

func Echof(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed := make(chan struct{})

	go func() {
		w.Write(body)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
	case <-r.Context().Done():
		logging.S().Infow("Эхо прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}

func Readyf(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func Sessionf(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func UserDataListf(w http.ResponseWriter, r *http.Request) {
	var info co.RequestList
	var err error
	if err = json.NewDecoder(r.Body).Decode(&info); err != nil {
		logging.S().Infow("Ошибка парсинга json " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(auth.CPuserID).(int64)
	completed := make(chan struct{})

	var data []co.Udata

	go func() {
		data, err = adb.GetUserDataList(r.Context(), userID, info)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			b, _ := json.Marshal(data)
			w.Write(b)
		}
	case <-r.Context().Done():
		logging.S().Infow("Чтение данных прервано на клиентской стороне.")
		w.WriteHeader(http.StatusGone)
	}
}
