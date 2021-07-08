package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/godbus/dbus"
	"gitlab.com/redfield/go-argo/argo"
	"net"
	"os"
)

const (
	argoHost = "29951"
	argoPort = "5555"
)

func bouncer() {
	l, err := argo.ListenStream(argoHost + ":" + argoPort)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}

	defer l.Close()

	fmt.Println("Listening on " + argoHost + ":" + argoPort)
	for {
		// Listen for Argo connections
		conn, err := l.Accept()
		fmt.Println("Address:", conn.RemoteAddr())
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	defer conn.Close()
	dconn, err := dbus.SystemBusPrivate()
	if err != nil {
		fmt.Println("Error finding system bus: ", err)
		os.Exit(1)
	}
	defer dconn.Close()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error(), "bytes:", n)
			break
		}

		bbuf := bytes.NewBuffer(buf[:n])
		msg, decodeErr := dbus.DecodeMessage(bbuf)
		if decodeErr != nil {
			fmt.Println("Error Decoding:", decodeErr.Error())
		} else {
			fmt.Println("Decoded:", msg.String())
			fmt.Println("Headers:", msg.Headers)
			fmt.Println("Body:", msg.Body)
		}

		fmt.Printf("Bus<- (%dB) %s\n", n, string(buf))
		fmt.Printf(hex.Dump(buf[:n]))

		n, err = dconn.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing:", err.Error(), "bytes:", n)
			break
		}

		buf2 := make([]byte, 1024)
		msg, decodeErr = dbus.DecodeMessage(bytes.NewBuffer(buf2[:n]))
		if decodeErr != nil {
			fmt.Println("Error Decoding:", decodeErr.Error())
		} else {
			fmt.Println("Decoded:", msg.String())
			fmt.Println("Headers:", msg.Headers)
			fmt.Println("Body:", msg.Body)
		}

		if n, err = dconn.Read(buf2); err != nil {
			fmt.Println("Error reading (DBus):", err.Error(), "bytes:", n)
			break
		}
		fmt.Printf("Call<- (%dB) %s\n", n, string(buf2))
		fmt.Printf(hex.Dump(buf2[:n]))
		if n, err = conn.Write(buf2[:n]); err != nil {
			fmt.Println("Error writing (Argo):", err.Error(), "bytes:", n)
			break
		}
	}

	conn.Close()
}
