package utils

import (
	"io"
	"net/http"

	"github.com/mbicl/mbicf_bot/adminlog"
	"github.com/mbicl/mbicf_bot/config"
)

func HTTPGet(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		adminlog.SendMessage("HTTPGet error: "+err.Error(), config.Ctx, config.B)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		adminlog.SendMessage("Error on reading response body (HTTPGet function):"+err.Error(), config.Ctx, config.B)
	}
	return body
}
