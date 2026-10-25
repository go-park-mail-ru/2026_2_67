package delivery

import (
	"fmt"
	"net/http"

	"github.com/metalagman/ddgo"
)

const userAgentKey = "User-Agent"

func deviceNameOf(r *http.Request) string {
	userAgent := r.Header.Get(userAgentKey)
	if userAgent == "" {
		return ""
	}

	detector, err := ddgo.New()
	if err != nil {
		return ""
	}

	result, err := detector.Parse(userAgent)
	if err != nil {
		return ""
	}

	if result.Bot.IsBot {
		return ""
	}

	return fmt.Sprintf("%s %s:%s:%s %s", result.Client.Name, result.Client.Engine, result.OS.Name, result.Device.Brand, result.Device.Model)
}
