package topsec

import (
	"fmt"
	"strings"

	"github.com/open-cmi/gmiko/devices/topsec/ngtos"
	"github.com/open-cmi/gmiko/types"
)

func NewDevice(deviceOS string, host string, port int, user string, password string) (types.Device, error) {

	if strings.ToLower(deviceOS) == "ngtos" {
		return ngtos.NewDevice(host, port, user, password), nil
	}

	return nil, fmt.Errorf("not supported device os")
}
