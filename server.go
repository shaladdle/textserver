package main

import (
	"flag"
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"log"
	"net"
	"os"
)

var (
	argAddr     = flag.String("addr", ":8000", "Address to listen on")
	argFilename = flag.String("file", "message.txt", "")
)

func listener(addr string, handlec chan net.Conn) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Error listening:", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting:", err)
			continue
		}

        log.Println("sending conn to handler")

		handlec <- conn
	}
}

func handler(texts <-chan string, conns <-chan net.Conn) {
	var ltext string
	for {
		select {
		case text := <-texts:
			ltext = text
		case conn := <-conns:
			if _, err := conn.Write([]byte(ltext)); err != nil {
				log.Println("Error writing:", err)
			}

			conn.Close()
		}
	}
}

func watcher(watcher *fsnotify.Watcher, texts chan<- string) {
	for {
		select {
		case ev := <-watcher.Event:
			if f, err := os.Open(ev.Name); err != nil {
				log.Println("Couldn't read from file")
			} else {
				if b, err := ioutil.ReadAll(f); err != nil {
					log.Println("Error reading file:", err)
                    f.Close()
                    continue
				} else {
					texts <- string(b)
				}
                f.Close()
			}
		case err := <-watcher.Error:
            log.Println("Watcher error::", err)
		}
	}
}

func main() {
	fsnot, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	if f, err := os.Open(*argFilename); err != nil {
		os.Create(*argFilename)
	} else {
		f.Close()
	}

	if err = fsnot.Watch(*argFilename); err != nil {
		log.Fatal(err)
	}

	texts := make(chan string)
	conns := make(chan net.Conn)

	go listener(*argAddr, conns)
	go watcher(fsnot, texts)
	go handler(texts, conns)

	select {}
}
