package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	applicationJson    = "application/json"
	loginEnding        = "/api/auth/login"
	saveDeviceEnding   = "/api/device"
	deleteDeviceEnding = "/api/device/"
	getDeviceIdEnding  = "/api/tenant/devices?deviceName=%s"
	getDeviceAuthToken = "/api/device/%s/credentials"
)

//tenant user
type TenantUser struct {
	ServerHost string `json:"-"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Jwt        string `json:"-"`
}

//Token is token
type Token struct {
	Token string
}

//device info
type Device struct {
	Id           string
	Name         string
	Type         string
	AdditionInfo string
	Label        string
	AuthToken    string
	SeqNo        int
}

//NewDevice returns a default Device Pointer
func NewDevice(name string) (*Device, error) {
	seqNo, err := extractSeqNo(name)
	if err != nil {
		return &Device{}, err
	}
	return &Device{Name: name, Type: "default", SeqNo: seqNo}, nil
}

func extractSeqNo(name string) (int, error) {
	splits := strings.Split(name, "_")
	if len(splits) == 0 {
		panic("device name not end with number")
	}
	seqNo, err := strconv.Atoi(splits[len(splits)-1])
	if err != nil {
		return 0, err
	}
	return seqNo, nil
}

//TBDevice correspond to response data struct from tb rest api
type TBDevice struct {
	AdditionalInfo string   `json:"additionalInfo"`
	CreatedTime    int64    `json:"createdTime"`
	CustomerId     EntityId `json:"customerId"`
	Id             EntityId `json:"id"`
	Label          string   `json:"label"`
	Name           string   `json:"name"`
	TenantId       EntityId `json:"tenantId"`
	Type           string   `json:"type"`
}

type EntityId struct {
	Id         string `json:"id"`
	EntityType string `json:"entityType"`
}

type Credential struct {
	Id               EntityId `json:"id"`
	CreatedTime      int64    `json:"createdTime"`
	DeviceId         EntityId `json:"deviceId"`
	CredentialsType  string   `json:"credentialsType"`
	CredentialsId    string   `json:"credentialsId"`
	CredentialsValue string   `json:"CredentialsValue"`
}

func NewTenantUser(serverHost, username, password string) (*TenantUser, error) {
	return &TenantUser{ServerHost: serverHost, Username: username, Password: password}, nil
}

//Login get token
func (u *TenantUser) Login() error {
	url := u.ServerHost + loginEnding
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, applicationJson, bytes.NewReader([]byte(data)))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var token Token
	if err = json.Unmarshal(all, &token); err != nil {
		return err
	}

	u.Jwt = fmt.Sprintf("Bearer %s", token.Token)

	return nil
}

type SaveDeviceRequest struct {
	DeviceName string `json:"name"`
	DeviceType string `json:"type"`
}

//CreateDevice creates device on thingsboard
func (u *TenantUser) CreateDevice(device *Device) error {
	url := u.ServerHost + saveDeviceEnding

	saveDeviceRequest := SaveDeviceRequest{device.Name, device.Type}
	data, err := json.Marshal(saveDeviceRequest)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	request.Header = map[string][]string{"Content-Type": {applicationJson}, "X-Authorization": {u.Jwt}}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()

	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	//log.Printf("response = %s\n" , string(all))
	var tbDevice TBDevice
	if err = json.Unmarshal(all, &tbDevice); err != nil {
		return err
	}
	//log.Printf("tbDevice = %+v\n", tbDevice)

	if len(tbDevice.Id.Id) == 0 {
		return fmt.Errorf("create fail|msg=%s", string(all))
	}

	device.Id = tbDevice.Id.Id
	device.AdditionInfo = tbDevice.AdditionalInfo
	device.Label = tbDevice.Label

	return nil
}

//DeleteDevice deletes device on thingsboard
func (u *TenantUser) DeleteDevice(deviceId string) error {
	url := u.ServerHost + deleteDeviceEnding + deviceId

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	request.Header = map[string][]string{"X-Authorization": {u.Jwt}}

	if _, err := http.DefaultClient.Do(request); err != nil {
		return err
	}

	return nil
}

//GetDevice gets device from thingsboard
func (u *TenantUser) GetDevice(deviceName string) (*Device, error) {
	url := u.ServerHost + fmt.Sprintf(getDeviceIdEnding, deviceName)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return &Device{}, err
	}
	request.Header = map[string][]string{"X-Authorization": {u.Jwt}}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return &Device{}, err
	}
	defer func() { _ = response.Body.Close() }()

	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return &Device{}, err
	}
	//log.Printf("%s\n", string(all))

	var tbDevice TBDevice
	if err = json.Unmarshal(all, &tbDevice); err != nil {
		return &Device{}, err
	}
	//log.Printf("%+v\n", tbDevice)

	return &Device{
		Id:           tbDevice.Id.Id,
		Name:         deviceName,
		Type:         tbDevice.Type,
		AdditionInfo: tbDevice.AdditionalInfo,
		Label:        tbDevice.Label}, nil
}

//GetDeviceAuthToken gets authToken of device on thingsboard
func (u *TenantUser) GetDeviceAuthToken(deviceId string) (string, error) {
	url := u.ServerHost + fmt.Sprintf(getDeviceAuthToken, deviceId)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header = map[string][]string{"X-Authorization": {u.Jwt}}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()

	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	//log.Printf("%s\n", string(all))
	var credential Credential
	if err := json.Unmarshal(all, &credential); err != nil {
		return "", err
	}

	if len(credential.CredentialsId) == 0 {
		return "", ErrDeviceAuthTokenEmpty
	}

	//log.Printf("%+v\n", credential)
	return credential.CredentialsId, nil
}
