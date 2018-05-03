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
├── junos-gnmi-testclient-junos-0.1   = Compiled for FreeBSD (runs on Junos itself)
├── junos-gnmi-testclient-linux-0.1   = Compiled for x64 based Linux
├── junos-gnmi-testclient-osx-0.1     = Compiled for OSX
```

The script requires some command line inputs as below.

```bash
./junos-gnmi-testclient -h
2018/05/03 18:20:07 -----------------------------------
2018/05/03 18:20:07 Junos gNMI Configuration Test Tool
2018/05/03 18:20:07 -----------------------------------
2018/05/03 18:20:07 Run the app with -h for options

Usage of ./junos-gnmi-testclient:
  -certdir string
    	Directory with clientCert.crt, clientKey.crt, CA.crt
  -cid string
    	Set to Client ID (default "1")
  -enc string
    	Encoding, either ASCII or JSON as of 18.1 (default "JSON")
  -host string
    	Set host to IP address or FQDN DNS record (default "127.0.0.1")
  -port string
    	Set to Server Port (default "50051")
  -resource string
    	Set resource to resource path (default "/interfaces")
  -user string
    	Set to username (default "testuser")
```

Here is how to run it in case this still doesn't make sense.

```bash
./junos-gnmi-testclient -host vmx01.corepipe.co.uk -port 50051 -user jet -resource system
2018/05/03 18:21:00 -----------------------------------
2018/05/03 18:21:00 Junos gNMI Configuration Test Tool
2018/05/03 18:21:00 -----------------------------------
2018/05/03 18:21:00 Run the app with -h for options

Enter Password:
----- VERSION -----
0.4.0
----- MODELS -----
<snip>
```

For the readers amongst you, note that the password field is missing. This is requested from you and the output is masked to prevent shoulder surfer dangers!
