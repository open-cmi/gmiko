package gmiko

import "github.com/open-cmi/gmiko/types"

type DeviceOption func(interface{}) error

func SecretOption(secret string) func(device interface{}) error {
	return func(device interface{}) error {
		device.(types.Device).SetSecret(secret)
		return nil
	}
}

func TimeoutOption(timeout int) func(device interface{}) error {
	return func(device interface{}) error {
		device.(types.Device).SetTimeout(timeout)
		return nil
	}
}
