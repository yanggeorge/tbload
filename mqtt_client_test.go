package main

import (
	"testing"
	"time"
)

func TestNewMqttClient(t *testing.T) {
	clientId := "dfasfasdfas"
	brokerUrl := "tcp://139.219.2.82:1883"
	user := "qvZmJEXmFUAf0lkyuIm6"
	client := NewMqttClient(clientId, user, "", brokerUrl)
	_, err := client.ConnectAndWait(5*time.Second, false)
	if err != nil {
		t.Fatal(err)
	}

	var payload string
	payload = `{"ts":1451649600512, "values":{"_tbload_key":1.4}}`

	topic := "v1/devices/me/telemetry"
	_, err = client.PublishAndWait(topic, payload, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

}
