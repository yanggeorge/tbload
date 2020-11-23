package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrUsage is returned when a usage message was printed and the process
	// should simply exit with an error.
	ErrUsage = errors.New("usage")

	// ErrUnknownCommand is returned when a CLI command is not specified.
	ErrUnknownCommand = errors.New("unknown command")
)

const (
	ToolPrefix       = "_tbload"
	DeviceNamePrefix = ToolPrefix + "_device"
	KeyInitCmdInfo   = ToolPrefix + "_init_cmd_info"
	KeyRunCmdInfo    = ToolPrefix + "_run_cmd_info"
	KeySummary       = ToolPrefix + "_summary"
)

func main() {
	m := NewMain()
	if err := m.Run(os.Args[1:]...); err == ErrUsage {
		os.Exit(2)
	} else if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// Main represents the main program execution.
type Main struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// NewMain returns a new instance of Main connect to the standard input/output.
func NewMain() *Main {
	return &Main{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Run executes the program.
func (m *Main) Run(args ...string) error {
	// Require a command at the beginning.
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fmt.Fprintln(m.Stderr, m.Usage())
		return ErrUsage
	}

	// Execute command.
	switch args[0] {
	case "help":
		fmt.Fprintln(m.Stderr, m.Usage())
		return ErrUsage
	case "init":
		return newInitCommand(m).Run(args[1:]...)
	case "info":
		return newInfoCommand(m).Run(args[1:]...)
	case "run":
		return newRunCommand(m).Run(args[1:]...)
	case "clean":
		return newCleanCommand(m).Run(args[1:]...)
	default:
		return ErrUnknownCommand
	}
}

type InfoCommand struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func newInfoCommand(m *Main) *InfoCommand {
	return &InfoCommand{
		Stdin:  m.Stdin,
		Stdout: m.Stdout,
		Stderr: m.Stderr,
	}
}

func (cmd *InfoCommand) Run(args ...string) error {

	fs := flag.NewFlagSet("info", flag.ContinueOnError)

	help := fs.Bool("h", false, "print this screen")
	printKVs := fs.Bool("d", false, "print key/value pairs in data store")
	//printSummary := fs.Bool("s", false, "print summary")

	if err := fs.Parse(args); err != nil {
		return err
	} else if *help {
		fs.Usage()
		return nil
	}

	store, err := OpenDeviceStore()
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	if *printKVs {
		return store.PrintAll(cmd.Stdout)
	}

	//if *printSummary {
	//	if summary, err := store.GetSummary(KeySummary); err != nil {
	//		return err
	//	} else {
	//		printSummary(summary)
	//	}
	//}

	return nil
}

type InitCommand struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	info InitCommandInfo
}

type InitCommandInfo struct {
	ServerHost string `json:"serverHost"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	DeviceNum  int    `json:"deviceNum"`
}

func newInitCommand(m *Main) *InitCommand {
	return &InitCommand{
		Stdin:  m.Stdin,
		Stdout: m.Stdout,
		Stderr: m.Stderr,
	}
}

func (cmd *InitCommand) Run(args ...string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	help := fs.Bool("h", false, "print this screen")

	fs.StringVar(&cmd.info.ServerHost, "serverHost", "", "thingsboard server host (e.g., http://demo.thingsboard.io)")
	fs.StringVar(&cmd.info.Username, "username", "", "name of thingsboard tenant user (e.g., tenant@thingsboard.org)")
	fs.StringVar(&cmd.info.Password, "password", "", " password of thingsboard tenant user")
	fs.IntVar(&cmd.info.DeviceNum, "deviceNum", 10, "number of devices to test")

	if err := fs.Parse(args); err != nil {
		return err
	} else if *help {
		fs.Usage()
		return nil
	} else if len(cmd.info.ServerHost) == 0 || len(cmd.info.Username) == 0 || len(cmd.info.Password) == 0 {
		fs.Usage()
		return ErrUsage
	}

	if err := checkLogin(cmd); err != nil {
		return err
	}

	if err := save(cmd.info); err != nil {
		return err
	}

	if err := createDevices(cmd); err != nil {
		return err
	}

	return nil
}

//checkLogin check account validity.
func checkLogin(cmd *InitCommand) error {
	serverHost := cmd.info.ServerHost
	username := cmd.info.Username
	password := cmd.info.Password

	user, err := NewTenantUser(serverHost, username, password)
	if err != nil {
		return err
	}

	if err = user.Login(); err != nil {
		return err
	}

	return nil
}

//save saves cmd info in store
func save(info InitCommandInfo) error {
	store, err := OpenDeviceStore()
	defer func() { _ = store.Close() }()
	if err != nil {
		return err
	}

	return info.Store(store)
}

//forEach iterates on deviceName
func (info *InitCommandInfo) forEach(fn func(info *InitCommandInfo, index int, deviceName string) error) error {
	for i := 0; i < info.DeviceNum; i++ {
		name := DeviceNamePrefix + "_" + strconv.Itoa(i)
		err := fn(info, i, name)
		if err != nil {
			return err
		}
	}
	return nil
}

//store saves info in store
func (info *InitCommandInfo) Store(store *DeviceStore) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return store.Put(KeyInitCmdInfo, data)
}

func (info *InitCommandInfo) Restore(store *DeviceStore) error {
	value, err := store.Get(KeyInitCmdInfo)
	if err != nil {
		return err
	}

	if len(string(value)) == 0 {
		return fmt.Errorf("%s not exist", KeyInitCmdInfo)
	}

	return json.Unmarshal(value, info)
}

//createDevices creates devices on thingsboard
func createDevices(cmd *InitCommand) error {
	info := cmd.info

	user, err := NewTenantUser(info.ServerHost, info.Username, info.Password)
	if err != nil {
		return err
	}

	err = user.Login()
	if err != nil {
		return err
	}

	store, err := OpenDeviceStore()
	defer func() { _ = store.Close() }()
	if err != nil {
		return err
	}

	return info.forEach(func(info *InitCommandInfo, index int, deviceName string) error {
		device, err2 := user.GetDevice(deviceName)
		if err2 != nil {
			return err2
		}

		//if exist, save info then return
		if len(device.Id) > 0 {
			return store.SaveDevice(device)
		}

		device, err2 = NewDevice(deviceName)
		if err2 != nil {
			return err2
		}

		if err2 = user.CreateDevice(device); err2 != nil {
			return err2
		}

		device.AuthToken, err2 = user.GetDeviceAuthToken(device.Id)
		if err2 == ErrDeviceAuthTokenEmpty {
			fmt.Printf("device[%s] got empty AuthToken\n", deviceName)
		} else if err2 != nil {
			return err2
		}

		fmt.Fprintf(cmd.Stdout, "created device[%s]\n", deviceName)
		return store.SaveDevice(device)
	})
}

type RunCommand struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	info RunCommandInfo
}

type RunCommandInfo struct {
	BrokerUrl      string `json:"brokerUrl"`
	MessageNum     int    `json:"messageNum"`
	Timeout        int    `json:"timeout"`
	ConcurrentNum  int    `json:"concurrentNum"`
	StartNum       int    `json:"startNum"`
	ConnectTimeout int    `json:"connectTimeout"`
}

func newRunCommand(m *Main) *RunCommand {
	return &RunCommand{
		Stdin:  m.Stdin,
		Stdout: m.Stdout,
		Stderr: m.Stderr,
	}
}

func (cmd *RunCommand) Run(args ...string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)

	help := fs.Bool("h", false, "print this screen")
	fs.StringVar(&cmd.info.BrokerUrl, "brokerUrl", "", "thingsboard mqtt transport (e.g., tcp://demo.thingsboard.io:1883)")
	fs.IntVar(&cmd.info.Timeout, "timeout", 10, "timeout of wait")
	fs.IntVar(&cmd.info.MessageNum, "messageNum", 10, "number of message to send")
	fs.IntVar(&cmd.info.ConcurrentNum, "concurrent", 1, "concurrent number of device to connect")
	fs.IntVar(&cmd.info.StartNum, "startNum", 0, "start num of device to start")
	fs.IntVar(&cmd.info.ConnectTimeout, "connectTimeout", 0, "connect timeout. 0 means wait until total timeout ")

	if err := fs.Parse(args); err != nil {
		return err
	} else if *help {
		fs.Usage()
		return nil
	} else if len(cmd.info.BrokerUrl) == 0 {
		fs.Usage()
		return nil
	}

	store, err := OpenDeviceStore()
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()
	//read initCommandInfo
	var initCmdInfo InitCommandInfo
	err = initCmdInfo.Restore(store)
	if err != nil {
		return err
	}

	//calculate condition variable
	if cmd.info.StartNum < 0 || cmd.info.StartNum+1 > initCmdInfo.DeviceNum {
		return fmt.Errorf("startNum[%d] out of range[0-%d]", cmd.info.StartNum, initCmdInfo.DeviceNum)
	}
	concurrent := cmd.info.ConcurrentNum
	startNum := cmd.info.StartNum
	endNum := cmd.info.StartNum + cmd.info.ConcurrentNum - 1
	if endNum+1 > initCmdInfo.DeviceNum {
		fmt.Printf("warning: startNum + concurrent[%d] out of range[%d]", endNum, initCmdInfo.DeviceNum)
		endNum = initCmdInfo.DeviceNum - 1
		concurrent = endNum - startNum + 1
		fmt.Printf("warning: real concurrent number will be %d", concurrent)
	}

	timeout := time.Duration(cmd.info.Timeout) * time.Second
	connectTimeout := timeout
	if cmd.info.ConnectTimeout > 0 {
		connectTimeout = time.Duration(cmd.info.ConnectTimeout) * time.Second
	}

	//init clients for every device
	clients := make([]*MqttClient, concurrent)
	defer func() { closeClients(clients) }()

	if err = initCmdInfo.forEach(func(info *InitCommandInfo, index int, deviceName string) error {
		device, err2 := store.GetDevice(deviceName)
		fmt.Fprintf(cmd.Stdout, "get %v\n", device)

		if err2 != nil {
			return err2
		}

		//filter the range between [startNum, endNum]
		if device.SeqNo < startNum || device.SeqNo > endNum {
			return nil
		}

		clientId := deviceName
		username := device.AuthToken
		password := ""
		brokerUrl := cmd.info.BrokerUrl
		clients[index-startNum] = NewMqttClient(clientId, username, password, brokerUrl)
		return nil
	}); err != nil {
		return err
	}
	fmt.Fprintf(cmd.Stdout, "%d clients init\n", len(clients))

	//connects all devices
	connectedQueue := make(chan Connected)
	for _, c := range clients {
		go func(c *MqttClient) {
			fmt.Fprintf(cmd.Stdout, "client[%s] begin connect..\n", c.Id)
			connected, _ := c.ConnectAndWait(timeout, connectTimeout)
			fmt.Fprintf(cmd.Stdout, "client[%s] connected = %v \n", c.Id, connected)
			connectedQueue <- connected
		}(c)
	}

	var nConnected int
	var counter int
	timer := time.NewTimer(timeout)
	isConnectTimeout := false
	for loop := true; loop; {
		select {
		case v := <-connectedQueue:
			counter += 1
			if v {
				nConnected += 1
			}

			if counter == concurrent {
				fmt.Fprintln(cmd.Stdout, "complete ..")
				loop = false
			}

		case <-timer.C:
			fmt.Fprintln(cmd.Stdout, "timeout ..")
			isConnectTimeout = true
			loop = false
		}
	}

	if isConnectTimeout {
		return fmt.Errorf("only %d/%d devices connected!", nConnected, concurrent)
	}

	if nConnected < counter {
		return fmt.Errorf("only %d/%d devices connected!", nConnected, concurrent)
	}
	fmt.Printf("all %d devices connected!\n", nConnected)

	//send messages
	fmt.Fprintln(cmd.Stdout, "sending messages...")
	resultQueue := make(chan Result, concurrent)
	for _, c := range clients {

		go func(c *MqttClient) {
			startTime := time.Now()
			ts := startTime.Unix() * 1000
			var published Published
			var count int
			timer := time.NewTimer(timeout)
			for i := 0; i < cmd.info.MessageNum; i++ {
				select {
				case <-timer.C:
					fmt.Printf("timeout [%s]\n", c.Id)
					resultQueue <- Result{
						ClientId:         c.Id,
						Event:            TimeoutExceededEvent,
						Error:            true,
						PublishDoneTime:  time.Since(startTime),
						MessagePublished: count,
					}
					break
				default:
					ts += 1
					template := `{"ts":%d, "values":{"%s_key":1.4}}`
					payload := fmt.Sprintf(template, ts, ToolPrefix)
					published, _ = c.PublishAndWait(tbPubTopic, payload, timeout)
					if !published {
						resultQueue <- Result{
							ClientId:         c.Id,
							Event:            PublishFailEvent,
							Error:            true,
							PublishDoneTime:  time.Since(startTime),
							MessagePublished: count,
						}
						break
					} else {
						count += 1
					}
				}
			}

			if count == cmd.info.MessageNum {
				//fmt.Printf("complete [%s]\n", c.Id)
				resultQueue <- Result{
					ClientId:         c.Id,
					Event:            PublishCompleteEvent,
					Error:            false,
					PublishDoneTime:  time.Since(startTime),
					MessagePublished: count,
				}
			}
		}(c)

	}

	timer = time.NewTimer(timeout * 2)
	results := make([]Result, concurrent)
	j := 0
LOOP2:
	for {
		select {
		case r := <-resultQueue:
			results[j] = r
			j += 1
			if j == concurrent {
				break LOOP2
			}
		case <-timer.C:
			fmt.Fprintf(cmd.Stdout, "receive timeout...\n")
			break LOOP2
		}
	}
	fmt.Fprintf(cmd.Stdout, "received %d \n", j)

	summary, err := buildSummary(concurrent, cmd.info.MessageNum, results)
	if err != nil {
		return err
	}
	printSummary(summary)
	//store.SaveSummary(KeySummary, summary)
	return nil
}

func closeClients(clients []*MqttClient) {
	for _, c := range clients {
		if c != nil && c.Client != nil {
			c.Client.Disconnect(5)
		}
	}
}

type CleanCommand struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func newCleanCommand(m *Main) *CleanCommand {
	return &CleanCommand{
		Stdin:  m.Stdin,
		Stdout: m.Stdout,
		Stderr: m.Stderr,
	}
}

func (cmd *CleanCommand) Run(args ...string) error {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)

	help := fs.Bool("h", false, "print this screen")

	if err := fs.Parse(args); err != nil {
		fs.Usage()
		return err
	} else if *help {
		fs.Usage()
		return nil
	}
	return cleanDevices(cmd)
}

//cleanDevices delete all devices on thingsboard and remove store file
func cleanDevices(cmd *CleanCommand) error {
	store, err := OpenDeviceStore()
	defer func() { _ = store.Close() }()
	if err != nil {
		return err
	}

	var info InitCommandInfo
	if err = info.Restore(store); err != nil {
		return err
	}

	fmt.Fprintf(cmd.Stdout, "%+v\n", info)

	var user *TenantUser
	if user, err = NewTenantUser(info.ServerHost, info.Username, info.Password); err != nil {
		return err
	}

	if err = user.Login(); err != nil {
		return err
	}

	if err = info.forEach(func(info *InitCommandInfo, index int, deviceName string) error {
		device, err2 := store.GetDevice(deviceName)
		if err2 != nil {
			return err2
		}

		fmt.Fprintf(cmd.Stdout, "to del device=%+v\n", device)

		return user.DeleteDevice(device.Id)
	}); err != nil {
		return err
	}

	//delete store
	return store.Drop()
}

// Usage returns the help message.
func (m *Main) Usage() string {
	return strings.TrimLeft(`
tbload is a MQTT load testing tool to ThingsBoard's MQTT broker. 

Usage:

	tbload command [arguments]

The commands are:

    clean       delete all devices 
    info        print key/value pair in store
    init        create devices
    help        print this screen
    run         connect devices and publish messages  

Use "tbload [command] -h" for more information about a command.
`, "\n")
}
