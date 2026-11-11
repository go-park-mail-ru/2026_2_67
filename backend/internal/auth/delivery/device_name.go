package delivery

import (
	"fmt"
	"net/http"

	"github.com/metalagman/ddgo"
)

const userAgentKey = "User-Agent"
const defaultDeviceName = "Unknown-Device-Name"
const errDeviceName = "Err-Device-Name"

func deviceNameOf(r *http.Request) string {
	userAgent := r.Header.Get(userAgentKey)
	if userAgent == "" {
		return defaultDeviceName
	}

	detector, err := ddgo.New()
	if err != nil {
		return errDeviceName
	}

	result, err := detector.Parse(userAgent)
	if err != nil {
		return errDeviceName
	}

	if result.Bot.IsBot {
		return defaultDeviceName
	}

	return fmt.Sprintf("%s %s:%s:%s %s", result.Client.Name, result.Client.Engine, result.OS.Name, result.Device.Brand, result.Device.Model)
}
