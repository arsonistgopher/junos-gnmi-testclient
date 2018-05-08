## Basic Junos gNMI Test Client

This test script exercises the gNMI capabilities on Junos.

At the time of writing, this has been tested on a 18.1 vMX and supports configuration GET and SET.

If you want telemetry data that would otherwise come from operational commands, use the telemetry scripts for testing or NETCONF.

Do not use this for anything other than curiosity!

## Usage

This package has been created with Godep support for dependencies.

```bash
go get github.com/arsonistgopher/junos-gnmi-testclient.git
cd $GOHOME/src/github.com/arsonistgopher/junos-gnmi-testclient
godep restore
go build
```

Also note, if you do not want to build this from source, three pre-compiled binaries have been included in the repo.

```bash
junos-gnmi-testclient-junos-0.1   = Compiled for FreeBSD (runs on Junos itself)
junos-gnmi-testclient-linux-0.1   = Compiled for x64 based Linux
junos-gnmi-testclient-osx-0.1     = Compiled for OSX
```

Before we get in to the nitty gritty execution detail, be warned that this application has support for TLS/SSL and it is always preferred that you use it. Below is a config snippet for Junos to enable SSL for gRPC. This also assumes basic knowledge of setting CA detail on Junos and registering certs etc.

```bash
set system services extension-service request-response grpc ssl local-certificate vmx01.domain
set system services extension-service request-response grpc ssl mutual-authentication certificate-authority CA
set system services extension-service request-response grpc ssl mutual-authentication client-certificate-request require-certificate
```

The script requires some command line inputs as below.

```bash
./junos-gnmi-testclient-osx-0.1 -h
2018/05/08 21:42:12 -----------------------------------
2018/05/08 21:42:12 Junos gNMI Configuration Test Tool
2018/05/08 21:42:12 -----------------------------------
2018/05/08 21:42:12 Run the app with -h for options

Usage of ./junos-gnmi-testclient-osx-0.1:
  -certdir string
    	Directory with client.crt, client.key, CA.crt
  -cid string
    	Set to Client ID (default "1")
  -enc string
    	Encoding, either ASCII or JSON as of 18.1 (default "JSON")
  -host string
    	Set host to IP address or FQDN DNS record (default "127.0.0.1")
  -port string
    	Set to Server Port, defaults to 32767 (default "32767")
  -resource string
    	Set resource to resource path (default "/interfaces")
  -shmodels
    	If set to true, then show supported models, else, do not
  -user string
    	Set to username (default "testuser")
```

Here is how to run it in case this still doesn't make sense.

```bash
./junos-gnmi-testclient-osx-0.1 -certdir CLIENTCERT -user jet -host vmx01.domain -resource /chassis
2018/05/08 21:40:49 -----------------------------------
2018/05/08 21:40:49 Junos gNMI Configuration Test Tool
2018/05/08 21:40:49 -----------------------------------
2018/05/08 21:40:49 Run the app with -h for options

Enter Password:
----- VERSION -----
0.4.0
----- ENCODINGS SUPPORTED -----
ASCII
JSON_IETF
----- GET DATA -----
{"fpc": [{"name": 0, "pic": [{"interface-type": "ge", "name": 0}], "number-of-ports": "12", "lite-mode": [null]}]}
```
