package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestOpenDeviceStore(t *testing.T) {
	store, err := OpenDeviceStore()
	defer store.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeviceStore_Drop(t *testing.T) {
	store, err := OpenDeviceStore()
	if err != nil {
		t.Fatal(err)
	}
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err = store.Drop(); err != nil {
		t.Fatal(err)
	}

}

func TestDeviceStore_PutGet(t *testing.T) {
	store, err := OpenDeviceStore()
	defer store.Close()
	if err != nil {
		t.Fatal(err)
	}
	key := Key("abc")
	v := Value("1234")
	if err := store.Put(key, v); err != nil {
		t.Fatal(err)
	}

	value, err := store.Get("abc")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(value, v) {
		t.Fatalf("expected %s not equal %s", v, value)
	}

	fmt.Println(string(value))
}

func TestDeviceStore_GetDevice(t *testing.T) {
	store, err := OpenDeviceStore()
	if err != nil {
		t.Fatal(err)
	}
	store.PrintAll(os.Stdout)
	device, err := store.GetDevice("_tbload_device_0")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", device)
}
