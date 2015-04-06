package ircx

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestNewBot(t *testing.T) {
	config := func(b *Bot) {
		b.user = "test-user"
	}
	b := New("irc.example.org", "test-bot", config)
	if b.server != "irc.example.org" {
		t.Fatalf("Wanted server %s, got %s", "irc.example.org", b.server)
	}
	if b.name != "test-bot" {
		t.Fatalf("Wanted name %s, got %s", "test-bot", b.name)
	}
	if b.user != "test-user" {
		t.Fatalf("Wanted user %s, got %s", "test-user", b.user)
	}
}

func TestConnect(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	b := New(l.Addr().String(), "test-bot")

	not := make(chan struct{})
	go dummyHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock server: %v", err)
	}

	// block on 2 seconds or recieving that the mock server has been connected to
	select {
	case <-not:
		return
	case <-time.After(2 * time.Second):
		t.Fatal("dummy server did not get connected to after 2 seconds")
	}
}

func TestSendsData(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	b := New(l.Addr().String(), "test-bot", WithLogin("test-user", "test-password"))

	not := make(chan string)
	go echoHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock server: %v", err)
	}

	// We should get back the connect info. If 500ms has happened and we haven't gotten anything
	// we're either not connected right, or all of the data has been sent.
	data := []string{}
	for {
		select {
		case d := <-not:
			data = append(data, d)
		case <-time.After(250 * time.Millisecond):
			goto DONE
		}
	}
DONE:
	d := b.connectMessages()
	for k, v := range d {
		if v.String() != data[k] {
			t.Fatalf("Should have recieved %s, got %s", d[k], v)
		}
	}
}

func TestSendsDataWithPassword(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	b := New(l.Addr().String(), "test-bot")

	not := make(chan string)
	go echoHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock server: %v", err)
	}

	// We should get back the connect info. If 500ms has happened and we haven't gotten anything
	// we're either not connected right, or all of the data has been sent.
	data := []string{}
	for {
		select {
		case d := <-not:
			data = append(data, d)
		case <-time.After(250 * time.Millisecond):
			goto DONE
		}
	}
DONE:
	d := b.connectMessages()
	for k, v := range d {
		if v.String() != data[k] {
			t.Fatalf("Should have recieved %s, got %s", d[k], v)
		}
	}
}

func TestReconnect(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	b := New(l.Addr().String(), "test-bot")

	not := make(chan struct{})
	go dcHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock server: %v", err)
	}
	tries := 0
	for {
		select {
		case <-not:
			tries++
			if tries > 2 {
				return
			}
		case <-time.After(1 * time.Second):
			t.Fatal("dummy server did not get reconnected in time, reconnect is broken")
		}
	}
}

func dummyHelper(l net.Listener, not chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func(c net.Conn) {
			not <- struct{}{}
		}(conn)
	}
}

func echoHelper(l net.Listener, not chan string) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func(c net.Conn) {
			rdr := bufio.NewReader(c)
			for {
				d, _, _ := rdr.ReadLine()
				not <- string(d)
			}
		}(conn)
	}
}

func dcHelper(l net.Listener, not chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func(c net.Conn) {
			time.Sleep(500 * time.Millisecond)
			c.Close()
			not <- struct{}{}
		}(conn)
	}
}
