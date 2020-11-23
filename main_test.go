package main

import (
	"fmt"
	"testing"
)

func TestInitCommand_Run(t *testing.T) {
	main := NewMain()
	args := []string{"init", "-serverHost", "http://www.enniot.net", "-username", "alenym@163.com", "-password", "123456", "-deviceNum", "1"}
	if err := main.Run(args...); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommand_Run(t *testing.T) {
	main := NewMain()
	args := []string{"run", "-brokerUrl", "http://localhost:1883", "-timeout", "200"}
	if err := main.Run(args...); err != nil {
		t.Fatal(err)
	}
}

func TestCleanCommand_Run(t *testing.T) {
	main := NewMain()
	args := []string{"clean"}
	if err := main.Run(args...); err != nil {
		t.Fatal(err)
	}
}

func TestInfoCommand_Run(t *testing.T) {
	main := NewMain()
	args := []string{"info", "-h"}
	if err := main.Run(args...); err != nil {
		t.Fatal(err)
	}
}

func TestInitCommandInfo_Restore(t *testing.T) {
	store, err := OpenDeviceStore()
	defer func() { _ = store.Close() }()
	if err != nil {
		t.Fatal(err)
	}

	var info InitCommandInfo
	if err = info.Restore(store); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", info)

}
