package main

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

func TestTenantUser_Login(t *testing.T) {
	//s := "http://localhost:4200"
	//u := "tenant@thingsboard.org"
	//p := "tenant"

	//s := "http://localhost:3000"
	//u := "fuhaiyinga@enn.cn"
	//p := "123456"

	s := "http://139.219.2.82:8080"
	u := "alenym@163.com"
	p := "123456"
	user, err := NewTenantUser(s, u, p)
	if err != nil {
		t.Fatal(err)
	}
	err = user.Login()
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%+v", *user)
	bytes, _ := json.Marshal(user)
	log.Println(string(bytes))

	device, err := NewDevice("device-abc1")
	if err != nil {
		t.Fatal(err)
	}

	err = user.CreateDevice(device)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", device)

}

func TestTenantUser_DeleteDevice(t *testing.T) {
	s := "http://139.219.2.82:8080"
	u := "alenym@163.com"
	p := "123456"
	user, err := NewTenantUser(s, u, p)
	if err != nil {
		t.Fatal(err)
	}
	err = user.Login()
	if err != nil {
		t.Fatal(err)
	}

	err = user.DeleteDevice("a0ab59a0-258e-11eb-8359-3b3967916152")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTenantUser_GetDevice(t *testing.T) {
	s := "http://139.219.2.82:8080"
	u := "alenym@163.com"
	p := "123456"
	user, err := NewTenantUser(s, u, p)
	if err != nil {
		t.Fatal(err)
	}
	err = user.Login()
	if err != nil {
		t.Fatal(err)
	}

	deviceName := "camera-1"

	device, err := user.GetDevice(deviceName)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("device=%+v", device)
}

func TestTenantUser_GetDeviceAuthToken(t *testing.T) {
	s := "http://139.219.2.82:8080"
	u := "alenym@163.com"
	p := "123456"
	user, err := NewTenantUser(s, u, p)
	if err != nil {
		t.Fatal(err)
	}
	err = user.Login()
	if err != nil {
		t.Fatal(err)
	}

	deviceId := "cd831170-11d4-11eb-b965-630b63ac6a0e"

	authToken, err := user.GetDeviceAuthToken(deviceId)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("authToken=%s", authToken)
	expected := "qvZmJEXmFUAf0lkyuIm6"
	if authToken != expected {
		t.Fatalf("not equal")
	}
}
