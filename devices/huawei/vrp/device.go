package vrp

import (
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Device struct {
	Host     string
	Port     int
	User     string
	Password string
	Return   string
	client   *ssh.Client
	session  *ssh.Session
	reader   io.Reader
	writer   io.WriteCloser
	outChan  chan []byte
	Mutex    sync.Mutex
}

func NewDevice(host string, port int, user string, password string) *Device {
	return &Device{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Return:   "\n",
		outChan:  make(chan []byte),
	}
}

var devicePattern *regexp.Regexp

func (c *Device) SetTimeout(timeout int) {
}

func (c *Device) SetSecret(secret string) {
}

func (c *Device) ReadRoutine() {

	for {
		var b []byte = make([]byte, 10200)
		n, err := c.reader.Read(b)
		if err != nil {
			break
		}
		if n == 0 {
			continue
		}

		c.outChan <- b[:n]
	}
}

func (c *Device) ReadCommandUntil(command string, expectPattern *regexp.Regexp, timeout int) ([]byte, error) {
	var res string
	var loop bool = true

	var commandIndex int = -1
	var lastRecv int64 = time.Now().Unix()
	for loop {
		select {
		case recv := <-c.outChan:
			res += string(recv)
			lastRecv = time.Now().Unix()
			if commandIndex == -1 {
				commandIndex = strings.Index(res, command)
			}
			if commandIndex != -1 && expectPattern.MatchString(res[commandIndex:]) {
				loop = false
			}
		case <-time.After(1 * time.Second):
			now := time.Now().Unix()
			if now-lastRecv > int64(timeout) {
				loop = false
			}
		}
	}

	return []byte(res), nil
}

func (c *Device) ReadUntil(expectPattern *regexp.Regexp, timeout int) ([]byte, error) {
	var res string
	var loop bool = true

	var lastRecv int64 = time.Now().Unix()
	for loop {
		select {
		case recv := <-c.outChan:
			res += string(recv)
			lastRecv = time.Now().Unix()
			if expectPattern.MatchString(res) {
				loop = false
			}
		case <-time.After(1 * time.Second):
			now := time.Now().Unix()
			if now-lastRecv > int64(timeout) {
				loop = false
			}
		}
	}

	return []byte(res), nil
}

func (c *Device) ReadTimeout(timeout int) []byte {
	var res string
	var loop bool = true

	var lastRecv int64 = time.Now().Unix()
	for loop {
		select {
		case recv := <-c.outChan:
			res += string(recv)
			lastRecv = time.Now().Unix()
		case <-time.After(1 * time.Second):
			now := time.Now().Unix()
			if now-lastRecv > int64(timeout) {
				loop = false
			}
		}
	}

	return []byte(res)
}

func (c *Device) Disconnect() {
	c.session.Close()
	c.client.Close()
}

func (c *Device) connect() error {
	// 设置SSH连接的配置
	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Password), // 或者使用 ssh.PublicKeys(你的公钥)
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境请使用ssh.FixedHostKey(hostKey)来确保安全
		Config: ssh.Config{
			KeyExchanges: []string{"diffie-hellman-group1-sha1", "diffie-hellman-group-exchange-sha1"},
			Ciphers:      []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		},
		Timeout: 5 * time.Second,
	}

	// 连接到SSH服务器
	target := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	client, err := ssh.Dial("tcp", target, config)
	if err != nil {
		return err
	}

	// 创建一个session
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return err
	}

	//设置terminalmodes的方式
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	//建立伪终端
	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		session.Close()
		client.Close()
		return err
	}

	//设置session的标准输入是stdin
	writer, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return err
	}
	reader, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		client.Close()
		return err
	}

	err = session.Shell()
	if err != nil {
		session.Close()
		client.Close()
		return err
	}
	c.reader = reader
	c.writer = writer

	c.session = session
	c.client = client

	go c.ReadRoutine()
	c.ReadUntil(devicePattern, 1)
	return nil
}

func (c *Device) Connect(maxTries int) error {
	var tries int = 0
	var err error
	for tries < maxTries {
		err = c.connect()
		if err == nil {
			break
		}
		tries++
		time.Sleep(1 * time.Second)
	}
	return err
}

func (c *Device) RunCommand(command string) ([]byte, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	runCmd := command
	if !strings.HasSuffix(runCmd, "\n") {
		runCmd += c.Return
	}
	c.writer.Write([]byte(runCmd))

	outByte, err := c.ReadCommandUntil(strings.Trim(command, "\r\n "), devicePattern, 2)
	resp := strings.Trim(strings.TrimPrefix(string(outByte), command), "\r\n")
	return []byte(resp), err
}

func (c *Device) ConfigCommand(command string) error {
	if !strings.HasSuffix(command, "\n") {
		command += c.Return
	}
	_, err := c.writer.Write([]byte(command))
	return err
}

func (c *Device) ConfigCommandSet(cmds []string) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	var err error
	for _, cmd := range cmds {
		err = c.ConfigCommand(cmd)
		if err != nil {
			break
		}
	}
	c.ReadTimeout(1)
	return err
}

func init() {
	devicePattern = regexp.MustCompile(`[<\[][\w\(\)\-]+[>\]]`)
}
