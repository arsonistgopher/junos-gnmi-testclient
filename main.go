package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	auth_pb "github.com/arsonistgopher/junos-gnmi-testclient/authentication"
	gnmipb "github.com/arsonistgopher/junos-gnmi-testclient/proto/gnmi"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {

	log.Println("-----------------------------------")
	log.Println("Junos gNMI Configuration Test Tool")
	log.Println("-----------------------------------")
	log.Print("Run the app with -h for options\n\n")

	// Parse flags
	var host = flag.String("host", "127.0.0.1", "Set host to IP address or FQDN DNS record")
	var resource = flag.String("resource", "/interfaces", "Set resource to resource path")
	var user = flag.String("user", "testuser", "Set to username")
	var port = flag.String("port", "50051", "Set to Server Port")
	var cid = flag.String("cid", "1", "Set to Client ID")
	var certDir = flag.String("certdir", "", "Directory with clientCert.crt, clientKey.crt, CA.crt")
	var encoding = flag.String("enc", "JSON", "Encoding, either ASCII or JSON as of 18.1")
	flag.Parse()

	// Set host
	hostandport := *host + ":" + *port

	// Grab password
	fmt.Print("Enter Password: \n")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Error reading password: %v", err)
	}
	password := string(bytePassword)

	// gRPC options
	var opts []grpc.DialOption

	// Are we going to run with TLS?
	runningWithTLS := false
	if *certDir != "" {
		runningWithTLS = true
	}

	// If we're running with TLS
	if runningWithTLS {

		// Grab x509 cert/key for client
		cert, err := tls.LoadX509KeyPair(fmt.Sprintf("%s/client.crt", *certDir), fmt.Sprintf("%s/client.key", *certDir))

		if err != nil {
			log.Fatalf("Could not load certFile: %v", err)
		}
		// Create certPool for CA
		certPool := x509.NewCertPool()

		// Get CA
		ca, err := ioutil.ReadFile(fmt.Sprintf("%s/CA.crt", *certDir))
		if err != nil {
			log.Fatalf("could not read ca certificate: %s", err)
		}

		// Append CA cert to pool
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			log.Fatal("Failed to append client certs")
		}

		// build creds
		creds := credentials.NewTLS(&tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{cert},
			ServerName:   *host,
		})

		if err != nil {
			log.Fatalf("Could not load clientCert: %v", err)
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else { // Else we're not running with TLS
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(hostandport, opts...)
	if err != nil {
		logrus.Fatalf("Error opening grpc.Dial(): %v", err)
	}
	// lazy close
	defer conn.Close()

	// Check for auth
	l := auth_pb.NewLoginClient(conn)
	dat, err := l.LoginCheck(context.Background(), &auth_pb.LoginRequest{UserName: *user, Password: password, ClientId: *cid})

	if err != nil {
		logrus.Fatalf("Could not login: %v", err)
	}

	if dat.Result == false {
		logrus.Fatalf("LoginCheck failed\n")
	}

	// Let's get a list of capabilities
	cap := &gnmipb.CapabilityRequest{}
	c := gnmipb.NewGNMIClient(conn)
	ctx := context.Background()

	resp, err := c.Capabilities(ctx, cap)

	if err != nil {
		logrus.Fatalf("Error getting capabilities: %v", err)
	}

	models := resp.GetSupportedModels()
	encodings := resp.GetSupportedEncodings()
	gnmiversion := resp.GetGNMIVersion()

	fmt.Println("----- VERSION -----")
	fmt.Println(gnmiversion)

	fmt.Println("----- MODELS -----")
	for _, m := range models {
		fmt.Println(m)
	}

	fmt.Println("----- ENCODINGS SUPPORTED -----")
	for _, e := range encodings {
		fmt.Println(e)
	}

	gpath, err := xpathToGNMIpath(*resource)

	if err != nil {
		logrus.Fatalf("Error: %v", err)
	}

	pp, err := StringToPath(pathToString(gpath), StructuredPath, StringSlicePath)

	if err != nil {
		logrus.Fatalf("Error: %v", err)
	}

	// Figure out what encoding we're asking for from command line arguments

	encodingUpper := strings.ToUpper(*encoding)

	var reqenc gnmipb.Encoding

	switch encodingUpper {
	case "JSON":
		reqenc = gnmipb.Encoding_JSON_IETF
	case "ASCII":
		reqenc = gnmipb.Encoding_ASCII
	default:
		reqenc = gnmipb.Encoding_JSON_IETF
	}

	// Up to 18.1, only CONFIG is supported for GET and SET hasn't been tested yet from my perspective
	getrequest := &gnmipb.GetRequest{
		Type:     gnmipb.GetRequest_CONFIG,
		Path:     []*gnmipb.Path{pp},
		Encoding: reqenc,
	}

	resp2, err := c.Get(ctx, getrequest)

	if err != nil {
		logrus.Fatalf("Error: %v", err)
	}

	fmt.Println("----- GET DATA -----")

	// Get the notification
	noti := resp2.GetNotification()

	// Iterate over the notifications
	for _, v1 := range noti {
		// Get the update within each notification
		upd := v1.GetUpdate()

		// Iterate over the update
		for _, v2 := range upd {
			// Logical split for encoding support
			switch reqenc {
			case gnmipb.Encoding_JSON_IETF:
				jsonBytes := v2.Val.GetJsonIetfVal()
				fmt.Println(string(jsonBytes))

			case gnmipb.Encoding_ASCII:
				ascii := v2.Val.GetAsciiVal()
				fmt.Println(ascii)
			}
		}
	}
}
