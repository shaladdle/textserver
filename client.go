package main

import (
	"bufio"
	"flag"
	"github.com/tarm/goserial"
	"log"
	"net"
)

var (
	argAddr = flag.String("addr", "localhost:8000", "")
	serport = flag.String("port", "/dev/ttyUSB0", "")
	baud    = flag.Int("baud", 9600, "")
)

func subscribe() {
	conn, err := net.Dial("tcp", *argAddr)
	if err != nil {
		log.Fatal("Couldn't connect to serveer:", err)
	}
	defer conn.Close()

	c := &serial.Config{Name: *serport, Baud: *baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	r := bufio.NewReader(conn)
	for {
		text, err := r.ReadString('\n')
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		if _, err = s.Write([]byte(text)); err != nil {
			log.Println("Error writing to serial:", err)
		}
	}
}

func main() {
	flag.Parse()

	for {
		time.Sleep(1000 * time.Second())
	}
}
