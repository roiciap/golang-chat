package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	loginUrl = "http://localhost:8080/login"
)

func selectApi() (url string, err error) {
	var (
		input string
	)

	fmt.Println("Wybierz api:\n1. Login")

	fmt.Scanln(&input)

	// input, err := reader.ReadString('\n')
	switch input {
	case "1":
		url = loginUrl
	default:
		err = errors.Join(err, errors.New("niepoprawny wybór"))
	}
	return
}

func prepareRequestBody() map[string]string {
	m := make(map[string]string)
	var (
		key   string
		value string
	)

	fmt.Println("budowanko ciala, aby zakonczyc wcisnij enter przy kluczu")

	for {
		fmt.Print("Klucz: ")
		n, _ := fmt.Scanln(&key)
		if n == 0 {
			break
		}
		fmt.Print("Wartość: ")
		n, _ = fmt.Scanln(&value)
		if n == 0 {
			value = ""
		}
		m[key] = value
	}
	return m
}

func sendRequest(bodyMap map[string]string, url string, client *http.Client) (res *http.Response, err error) {
	requestBody, _ := json.Marshal(bodyMap)
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(requestBody))
	if err != nil {
		return
	}
	return (*client).Do(req)

}

func main() {
	client := http.Client{Timeout: 10 * time.Second}

	url, err := selectApi()
	if err != nil {
		log.Fatalln("Natrafiono na błąd:", err.Error())
	}

	fmt.Println("uri to: ", url)
	requestBodyMap := prepareRequestBody()

	res, err := sendRequest(requestBodyMap, url, &client)
	if err != nil {
		log.Fatalln("Natrafiono na błąd:", err.Error())
	}

	fmt.Printf("Status Code: %d\n", res.StatusCode)
	resBody, _ := io.ReadAll(res.Body)
	fmt.Println(string(resBody))

}
