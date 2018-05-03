## Basic Junos OpenConfig telemetry tester

This is a Go script and demonstrates how to retrieve OpenConfig telemetry KV pair data from Junos.

With version 18.1 it's also possible to subscribe to the gNMI and retrieve gRPC encoded data as well as self-describing keyvalue pairs.

This test script takes the concept of [Nilesh Simaria's JTIMon](https://github.com/nileshsimaria/jtimon) and boils it down to the raw basics. 

Do not use this for anything other than curiosity!

## Usage

This package has been created with Godep support for dependencies.

```bash
go get github.com/arsonistgopher/gojtemtestoc.git
cd $GOHOME/src/github.com/arsonistgopher/gojtemtesttoc
godep restore
go build
./gojtemtestoc
```

The script requires some command line inputs as below.

```bash
./gojtemtestoc -h
Usage of ./gojtemtestoc:
  -certdir string
    	Directory with clientCert.crt, clientKey.crt, CA.crt
  -cid string
    	Set to Client ID (default "1")
  -host string
    	Set host to IP address or FQDN DNS record (default "127.0.0.1")
  -loops int
    	Set loops to desired iterations (default 1)
  -port string
    	Set to Server Port (default "50051")
  -resource string
    	Set resource to resource path (default "/interfaces")
  -smpfreq int
    	Set to sample frequency in milliseconds (default 1000)
  -user string
    	Set to username (default "testuser")
```

Here is how to run it in case this still doesn't make sense.

```bash
./gojtemtestoc -cid 42 -host HOST -port 50051 -loops 1 -resource /interfaces -smpfreq 1000 -user jet
```
Replace `HOST` with the hostname or IP address of your code. Replace `50051` with the port your grpc server on Junos is listening on. For the resource you want telemetry on, replace `/interfaces` with your chosen OpenConfig sensor.

For the readers amongst you, note that the password field is missing. This is requested from you and the output is masked to prevent shoulder surfer dangers!
