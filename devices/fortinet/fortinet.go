package fortinet

import (
	"fmt"
	"strings"

	"github.com/open-cmi/gmiko/devices/fortinet/fortios"
	"github.com/open-cmi/gmiko/types"
)

func NewDevice(deviceOS string, host string, port int, user string, password string) (types.Device, error) {

	if strings.ToLower(deviceOS) == "fortios" {
		return fortios.NewDevice(host, port, user, password), nil
	}

	return nil, fmt.Errorf("not supported device type")
}
