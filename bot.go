package ircx

import (
	"crypto/tls"
	"log"
	"math"
	"net"
	"time"

	"github.com/sorcix/irc"
)

type Bot struct {
	// required options
	server string
	name   string

	// optional params
	password  string
	user      string
	tlsConfig *tls.Config

	events    chan *irc.Message
	sender    ServerSender
	callbacks map[string][]Callback
	reader    *irc.Decoder
	writer    *irc.Encoder
	conn      net.Conn
	tries     float64
	connected bool
}

func New(server, name string, f ...func(*Bot)) *Bot {
	b := &Bot{
		server:    server,
		name:      name,
		connected: false,
		events:    make(chan *irc.Message),
		callbacks: make(map[string][]Callback),
		tries:     0,
	}
	for _, v := range f {
		v(b)
	}
	return b
}

// Connect attempts to connect to the given IRC server
func (b *Bot) Connect() error {
	var conn net.Conn
	var err error
	if b.tlsConfig == nil {
		conn, err = net.Dial("tcp", b.server)
	} else {
		conn, err = tls.Dial("tcp", b.server, b.tlsConfig)
	}
	if err != nil {
		return err
	}
	b.conn = conn
	b.reader = irc.NewDecoder(conn)
	b.writer = irc.NewEncoder(conn)
	b.sender = ServerSender{writer: &b.writer}
	for _, msg := range b.connectMessages() {
		err := b.writer.Encode(msg)
		if err != nil {
			return err
		}
	}
	log.Println("Connected to", b.server)
	b.tries = 0
	b.connected = true
	go b.readLoop()
	return nil
}

// Reconnect
func (b *Bot) Reconnect() {
	if b.connected {
		b.conn.Close()
		for err := b.Connect(); err != nil; err = b.Connect() {
			duration := time.Duration(math.Pow(2.0, b.tries)*200) * time.Millisecond
			log.Printf("Unable to connect to %s, waiting %s", b.server, duration.String())
			time.Sleep(duration)
			b.tries++
		}
	}
}

// ReadLoop sets a timeout of 300 seconds, and then attempts to read
// from the IRC server. If there is an error, it calls Reconnect
func (b *Bot) readLoop() {
	for {
		b.conn.SetDeadline(time.Now().Add(300 * time.Second))
		msg, err := b.reader.Decode()
		if err != nil {
			log.Println("Error:", err)
			b.Reconnect()
			return
		}
		b.events <- msg
	}
}
