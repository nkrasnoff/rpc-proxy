package main

import (
	//	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/godbus/dbus"
	"gitlab.com/redfield/go-argo/argo"
	"net"
	"os"
)

const (
	argoHost = "29951"
	argoPort = "7777"
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

	out_chan := make(chan []byte)
	in_chan := make(chan []byte)

	go dbus_w(dconn, in_chan)
	go argo_send(conn, out_chan)

	go argo_read(conn, in_chan)
	go dbus_r(dconn, out_chan)
	for {

	}

}

func argo_read(conn net.Conn, ch chan<- []byte) {

	for {

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error(), "bytes:", n)
			break
		}

		//bbuf := bytes.NewBuffer(buf[:n])

		ch <- buf[:n]

	}

}

func argo_send(conn net.Conn, ch <-chan []byte) {

	for {

		msg := <-ch

		fmt.Printf(hex.Dump(msg))
		n, err := conn.Write(msg)
		if err != nil {
			fmt.Println("Error Writing:", err.Error(), "bytes:", n)
			break
		}

	}

}

func dbus_w(dconn *dbus.Conn, ach <-chan []byte) {

	for {
		buf := <-ach
		/*msg, decodeErr := dbus.DecodeMessage(bbuf)
		if decodeErr != nil {
			fmt.Println("Error Decoding:", decodeErr.Error())
		} else {
			fmt.Println("Decoded:", msg.String())
			fmt.Println("Headers:", msg.Headers)
			fmt.Println("Body:", msg.Body)
		}
		*/

		//fmt.Printf("Bus<- (%dB) %s\n", n, string(buf))
		fmt.Printf(hex.Dump(buf))
		n, err := dconn.Write(buf)
		if err != nil {
			fmt.Println("Error Writing:", err.Error(), "bytes:", n)
			break
		}
	}

}

func dbus_r(dconn *dbus.Conn, ch chan<- []byte) {

	for {
		fmt.Println("this is actually working")
		buf := make([]byte, 1024)
		n, err := dconn.Read(buf)
		if err != nil {
			fmt.Println("Error Reading:", err.Error(), "bytes:", n)
			break
		}
		ch <- buf[:n]
	}
}
