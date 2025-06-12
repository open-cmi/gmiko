package ruijie

import (
	"fmt"
	"strings"

	"github.com/open-cmi/gmiko/devices/ruijie/rgos"
	"github.com/open-cmi/gmiko/types"
)

func NewDevice(deviceOS string, host string, port int, user string, password string) (types.Device, error) {
	if strings.ToLower(deviceOS) == "rgos" {
		return rgos.NewDevice(host, port, user, password), nil
	}

	return nil, fmt.Errorf("not supported device type")
}
