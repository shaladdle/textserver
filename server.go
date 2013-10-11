package main

import (
	"flag"
	"fmt"
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
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
	clients := map[string]net.Conn{}

	key := func(conn net.Conn) string {
		return conn.RemoteAddr().String()
	}

	send := func(conn net.Conn, msg string) {
		if _, err := fmt.Fprintf(conn, "%s", msg); err != nil {
			delete(clients, key(conn))
			log.Println("Error writing:", err)
			log.Println("Closing connection")
			conn.Close()
		}
	}

	ltext := ""
	for {
		select {
		case text := <-texts:
			ltext = text
			for _, conn := range clients {
				send(conn, ltext)
			}
		case conn := <-conns:
			clients[key(conn)] = conn
			send(conn, ltext)
		}
	}
}

func watcher(texts chan<- string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher error:", err)
	}

	if f, err := os.Open(*argFilename); err != nil {
		os.Create(*argFilename)
	} else {
		f.Close()
	}

	if err = watcher.Watch(*argFilename); err != nil {
		log.Fatal("Watch error:", err)
	}

	pipeFile := func(ev *fsnotify.FileEvent, texts chan<- string) {
		if f, err := os.Open(ev.Name); err != nil {
			log.Println("Couldn't read from file")
		} else {
			if b, err := ioutil.ReadAll(f); err != nil {
				log.Println("Error reading file:", err)
				f.Close()
				return
			} else {
				texts <- string(b)
			}
			f.Close()
		}
	}

	var (
		done   <-chan time.Time
		lastev *fsnotify.FileEvent
	)

	for {
		select {
		case <-done:
			pipeFile(lastev, texts)
		case ev := <-watcher.Event:
			switch {
			case ev.IsRename():
				if err = watcher.Watch(*argFilename); err != nil {
					log.Fatal("Loop watch error:", err)
				}
			default:
				log.Println(ev)
				lastev = ev
				done = time.After(100 * time.Millisecond)
			}
		case err := <-watcher.Error:
			log.Println("Watcher error::", err)
		}
	}
}

func main() {
	texts := make(chan string)
	conns := make(chan net.Conn)

	go listener(*argAddr, conns)
	go watcher(texts)
	go handler(texts, conns)

	select {}
}
