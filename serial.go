package main

import (
	"flag"
	"github.com/tarm/goserial"
	"log"
	"time"
)

var (
	serport = flag.String("port", "/dev/ttyUSB0", "")
	baud    = flag.Int("baud", 9600, "")
)

func main() {
	flag.Parse()
	c := &serial.Config{Name: *serport, Baud: *baud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	for {
		if _, err = s.Write([]byte{'a'}); err != nil {
			log.Fatal(err)
		}

		time.Sleep(1000 * time.Millisecond)

		/*
				b := make([]byte, 128)
				if n, err := s.Read(b); err != nil {
					log.Fatal("Error reading:", err)
				} else {
		            log.Println("Read", n, "bytes\t", string(b))
				}
		*/
	}
}
