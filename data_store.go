package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"io"
	"os"
	"time"
)

const (
	DbFileName    = "store.db"
	BucketDevices = "bkt_devices"
)

var (
	ErrDeviceNotExist       = errors.New("device not exist")
	ErrDeviceIdEmpty        = errors.New("deviceId is empty")
	ErrDeviceNameEmpty      = errors.New("deviceName is empty")
	ErrDeviceAuthTokenEmpty = errors.New("device's authToken is empty")
)

//DeviceStore save created device info
type DeviceStore struct {
	*bolt.DB
}

type Key string
type Value []byte

//Open opens a boltDb
func OpenDeviceStore() (*DeviceStore, error) {
	db, err := bolt.Open(DbFileName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return &DeviceStore{}, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err2 := tx.CreateBucketIfNotExists([]byte(BucketDevices)); err2 != nil {
			return err2
		}
		return nil
	})
	return &DeviceStore{db}, nil
}

//Close closes store
func (store *DeviceStore) Close() error {
	return store.DB.Close()
}

//Drop deletes db file.
func (store *DeviceStore) Drop() error {
	return os.Remove(DbFileName)
}

//Put puts key value pair
func (store *DeviceStore) Put(key Key, value Value) error {
	err := store.DB.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(BucketDevices))
		if bkt == nil {
			return fmt.Errorf("bucket not exist")
		}

		if err := bkt.Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

//Get gets value by key
func (store *DeviceStore) Get(key Key) (Value, error) {
	var data []byte

	err := store.DB.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(BucketDevices))
		if bkt == nil {
			return fmt.Errorf("bucket not exist")
		}

		data = bkt.Get([]byte(key))
		return nil
	})

	if err != nil {
		return nil, err
	}
	return data, nil
}

//Save saves device info
func (store *DeviceStore) SaveDevice(device *Device) error {
	if len(device.Name) == 0 {
		return fmt.Errorf("device name is empty")
	}

	bytes, err := json.Marshal(device)
	if err != nil {
		return err
	}

	err = store.Put(Key(device.Name), Value(bytes))
	if err != nil {
		return err
	}

	return nil
}

//Get
func (store *DeviceStore) GetDevice(deviceName string) (*Device, error) {
	value, err2 := store.Get(Key(deviceName))
	if err2 != nil {
		return &Device{}, err2
	}

	if len(value) == 0 {
		fmt.Printf("cannot get device[%s] from store", deviceName)
		return &Device{}, ErrDeviceNotExist
	}

	var device Device
	if err2 := json.Unmarshal([]byte(value), &device); err2 != nil {
		return &Device{}, err2
	}

	if len(device.Name) == 0 {
		fmt.Printf("device[%s] has not name in store", deviceName)
		return &Device{}, ErrDeviceNameEmpty
	}

	if len(device.Id) == 0 {
		fmt.Printf("device[%s] has not id in store", deviceName)
		return &Device{}, ErrDeviceIdEmpty
	}

	if len(device.AuthToken) == 0 {
		fmt.Printf("device[%s] has not authToken in store", deviceName)
		return &Device{}, ErrDeviceAuthTokenEmpty
	}

	return &device, nil
}

func (store *DeviceStore) PrintAll(writer io.Writer) error {
	fmt.Fprintln(writer, "key-values in store:")
	_ = store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketDevices))

		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Fprintf(writer, "- key=%s, value=%s\n", k, v)
		}

		return nil
	})
	return nil
}
