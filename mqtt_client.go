package main

import (
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

const (
	qos                 = 1
	tbPubTopic          = "v1/devices/me/telemetry"
	retained            = false
	autoReconnect       = true
	messageChannelDepth = 1000
)

var (
	ErrMqttClientConnectTimeout = errors.New("mqtt client connect timeout")
)

type MqttClient struct {
	mqtt.Client
	Id string
}

type Connected bool
type Published bool

type Result struct {
	ClientId string
	Event    string
	Error    bool

	MessagePublished int
	PublishTime      time.Duration
	PublishDoneTime  time.Duration
}

func NewMqttClient(clientId, username, password, brokerUrl string) *MqttClient {
	options := mqtt.NewClientOptions().SetClientID(clientId).SetUsername(username).SetPassword(password).AddBroker(brokerUrl)
	if autoReconnect {
		options.SetAutoReconnect(autoReconnect)
		options.SetMessageChannelDepth(messageChannelDepth)
	}
	client := mqtt.NewClient(options)
	return &MqttClient{client, clientId}
}

func (c *MqttClient) ConnectAndWait(totalTimeout, connectTimeout time.Duration) (Connected, error) {
	timer := time.NewTimer(totalTimeout)
	for retry := true; retry; {
		select {
		case <-timer.C:
			retry = false
		default:
			token := c.Connect()
			connected := token.WaitTimeout(connectTimeout)
			err := token.Error()
			if connected && err == nil {
				return true, nil
			} else if err != nil {
				return false, err
			}
			fmt.Printf("retry [%s]\n", c.Id)
		}
	}
	return false, ErrMqttClientConnectTimeout
}

func (c *MqttClient) PublishAndWait(topic, payload string, timeout time.Duration) (Published, error) {
	token := c.Publish(topic, qos, retained, payload)
	published := token.WaitTimeout(timeout)
	err := token.Error()
	if published == false || err != nil {
		if err != nil {
			fmt.Printf("err=%s", err)
		}
		return false, errors.New("publish fail")
	}

	return true, nil
}
