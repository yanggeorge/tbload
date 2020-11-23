# tbload 

`tbload` is a MQTT load testing tool to ThingsBoard's MQTT broker. 

## Build

You need golang version 1.15 to build the binaries.
```
$ git clone https://github.com/yanggeorge/tbload.git
$ cd tbload 
$ go mod vendor
$ go build .
```

This will generate binary file `tbload`.

if download error, please try to set `GOPROXY`. 

## Run

`tbload` has some commands. Running `tbload` directly will show help.

```bash
The commands are:

    clean       delete all devices 
    info        print key/value pair in store
    init        create devices
    help        print this screen
    run         connect devices and publish messages  

Use "tbload [command] -h" for more information about a command.
```

## Example

We use `demo.thingsboard.io` to show the example. The username is my account.

### 1. create devices on demo.thingsboard.io`
```bash
$ tbload init -serverHost http://demo.thingsboard.io -username alenym@gmail.com -password 123456 -deviceNum 1
created device[_tbload_device_0]
```
The device with deviceName="_tbload_device_0" has been created And devices' info has been kept in `store.db` file.
If you sign in `demo.thingsboard.io` with username `alenym@gmail.com` and password `123456`, "_tbload_device_0" will be there.

### 2. show the created devices info
```bash
$ tbload info -d
key-values in store:
- key=_tbload_device_0, value={"Id":"0ea65490-2d74-11eb-b06e-43eee10da81f","Name":"_tbload_device_0","Type":"default","AdditionInfo":"","Label":"","AuthToken":"31c318vQh1mAqARq3oVz","SeqNo":0}
- key=_tbload_init_cmd_info, value={"serverHost":"http://demo.thingsboard.io","username":"alenym@gmail.com","password":"123456","deviceNum":1}
```

### 3. run mqtt load test

The broker url is `tcp://demo.thingsboard.io:1883`.

```bash
$ tbload run -brokerUrl tcp://demo.thingsboard.io:1883 -messageNum 10
get &{0ea65490-2d74-11eb-b06e-43eee10da81f _tbload_device_0 default   31c318vQh1mAqARq3oVz 0}
1 clients init
client[_tbload_device_0] begin connect..
client[_tbload_device_0] connected = true 
complete ..
all 1 devices connected!
sending messages...
received 1 

# Configuration
Concurrent Clients: 1
Messages / Client:  10

# Results
Published Messages: 10 (100%)
Completed:          1 (100%)
Errors:             0 (0%)

# Publishing Throughput
Fastest: 15 msg/sec
Slowest: 15 msg/sec
Median: 15 msg/sec

  < 15 msg/sec  100%
```

### 4. delete created test devices.
```bash
$ tbload clean
{ServerHost:http://demo.thingsboard.io Username:alenym@gmail.com Password:123456 DeviceNum:1}
to del device=&{Id:0ea65490-2d74-11eb-b06e-43eee10da81f Name:_tbload_device_0 Type:default AdditionInfo: Label: AuthToken:31c318vQh1mAqARq3oVz SeqNo:0}
```




 

