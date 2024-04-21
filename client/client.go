package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	client := http.Client{}
	req1, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/register", bytes.NewBuffer([]byte(`{"login": "dan", "password": "0000"}`)))
	resp1, _ := client.Do(req1)
	body1, _ := io.ReadAll(resp1.Body)
	fmt.Println(string(body1))
	req2, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/login", bytes.NewBuffer([]byte(`{"login": "dan", "password": "000"}`)))
	resp2, _ := client.Do(req2)
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Println(string(body2))
	req3, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/login", bytes.NewBuffer([]byte(`{"login": "dan", "password": "0000"}`)))
	resp3, _ := client.Do(req3)
	body3, _ := io.ReadAll(resp3.Body)
	fmt.Println(string(body3))
}
