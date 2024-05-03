package utils

import (
	"io"
	"log"
	"net/http"
)

func HTTPGet(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("HTTPGet error:", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error on reading response body (HTTPGet function):", err)
	}
	return body
}
