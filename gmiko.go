package gmiko

import (
	"fmt"

	"github.com/open-cmi/gmiko/devices/cisco"
	"github.com/open-cmi/gmiko/devices/fortinet"
	"github.com/open-cmi/gmiko/devices/h3c"
	"github.com/open-cmi/gmiko/devices/huawei"
	"github.com/open-cmi/gmiko/devices/ruijie"
	"github.com/open-cmi/gmiko/types"
)

func NewDevice(manu string, deviceOS string, host string, port int, username string, password string, Options ...DeviceOption) (types.Device, error) {
	var device types.Device
	var err error

	//create the Device
	if manu == "h3c" {
		device, err = h3c.NewDevice(deviceOS, host, port, username, password)
	} else if manu == "huawei" {
		device, err = huawei.NewDevice(deviceOS, host, port, username, password)
	} else if manu == "cisco" {
		device, err = cisco.NewDevice(deviceOS, host, port, username, password)
	} else if manu == "ruijie" {
		device, err = ruijie.NewDevice(deviceOS, host, port, username, password)
	} else if manu == "fortinet" {
		device, err = fortinet.NewDevice(deviceOS, host, port, username, password)
	} else {
		return nil, fmt.Errorf("not supported manufature")
	}

	if err != nil {
		return nil, err
	}

	// running Options Functions.
	for _, option := range Options {
		err := option(device)
		if err != nil {
			return nil, err
		}
	}

	return device, nil
}
