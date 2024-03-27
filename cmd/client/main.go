package main

import (
	"fmt"

	"github.com/vzveiteskostrami/goph-keeper/internal/cconfig"
	"github.com/vzveiteskostrami/goph-keeper/internal/chttp"
	"github.com/vzveiteskostrami/goph-keeper/internal/oper"
)

func main() {
	if err := cconfig.ReadData(); err != nil {
		fmt.Println("Ошибка чтения конфигурации:", err)
		return
	}

	cfg := cconfig.Get()

	if cfg.Operation == nil || *cfg.Operation == "" {
		fmt.Println("Не указана операция \"command line>client -o=<operation>\"")
		return
	}

	// Если операция требует обращения к серверу, сначала просто проверим его
	// наличие в системе. Если его нет, то и затевать ничего не надо.
	if *cfg.Operation == "registration" ||
		*cfg.Operation == "login" ||
		*cfg.Operation == "sync" {
		err := chttp.CheckServerPresent()
		if err != nil {
			fmt.Println("Проверка сервера:", err)
			fmt.Println("Сервер неработоспособен. Операция в данный момент невозможна.")
			fmt.Println("Попробуйте выполнить операцию позднее.")
			return
		}
	}

	if *cfg.Operation == "registration" {
		oper.Registration(*cfg.Login, "")
	} else if *cfg.Operation == "login" {
		oper.Authorization(*cfg.Login, "", 0)
	} else if *cfg.Operation == "sync" {
		oper.Syncronize()
	} else {
		fmt.Println("Не опознана операция " + *cfg.Operation)
	}
}

/*

	//req, err := http.NewRequest("GET", `https://example.net`, nil)
	//ctx, _ := context.WithTimeout(context.TODO(), 200 * time.Milliseconds)
	//resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	// Be sure to handle errors.
	//defer resp.Body.Close()

	data := `{"name": "Иванов Иван", "email": "ivan@example.com"}`
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	response, err := client.Post("http://localhost:8080/echo", "application/json", strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer response.Body.Close()

	fmt.Printf("Status Code: %d\r\n", response.StatusCode)
	for k, v := range response.Header {
		// заголовок может иметь несколько значений,
		// но для простоты запросим только первое
		fmt.Printf("%s: %v\r\n", k, v[0])
	}
	body, _ := io.ReadAll(response.Body)
	fmt.Println("Body:\r\n", string(body))
}
*/
