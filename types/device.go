package types

type Device interface {
	Connect(maxTries int) error
	Disconnect()
	RunCommand(cmd string) ([]byte, error)
	ConfigCommandSet(cmds []string) error
	SetSecret(secret string)
	SetTimeout(timeout int)
}
