package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"

	irc "github.com/fluffle/goirc/client"
)

const (
	ident    = "ircalert"
	realname = "github.com/wuvt/ircalert"
	version  = "github.com/wuvt/ircalert"
)

func main() {
	var nick string
	var server string
	var ssl bool
	var proxy string
	var channel string
	var message string
	var joinFirst bool
	flag.StringVar(&nick, "nick", "ircalert", "IRC nick to use")
	flag.StringVar(&server, "server", "", "IRC server")
	flag.BoolVar(&ssl, "ssl", false, "Use TLS to connect to the IRC server")
	flag.StringVar(&proxy, "proxy", "", "Proxy server to use")
	flag.StringVar(&channel, "channel", "", "Channel to send message to")
	flag.StringVar(&message, "message", "", "Message to send")
	flag.BoolVar(&joinFirst, "join", false, "Join the channel before sending")
	flag.Parse()

	serverHost, _, err := net.SplitHostPort(server)
	if err != nil {
		panic(err)
	}

	if len(channel) <= 0 {
		log.Fatal("-channel is required.")
	}

	if len(message) <= 0 {
		log.Fatal("-message is required.")
	}

	cfg := irc.NewConfig(nick)
	cfg.Server = server
	cfg.Me.Ident = nick
	cfg.Me.Name = realname
	cfg.Version = version

	if ssl {
		cfg.SSL = true
		cfg.SSLConfig = &tls.Config{ServerName: serverHost}
		log.Print("TLS is enabled.")
	}

	if len(proxy) > 0 {
		cfg.Proxy = proxy
		log.Printf("Using proxy: %q", proxy)
	}

	c := irc.Client(cfg)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		log.Printf("Connected to %q", server)

		if joinFirst {
			conn.Join(channel)
			log.Printf("Joined %q", channel)
		}

		conn.Privmsg(channel, message)

		if joinFirst {
			conn.Part(channel)
			log.Printf("Parted %q", channel)
		}

		conn.Quit()
	})

	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		log.Print("Disconnected")
		quit <- true
	})

	if err := c.Connect(); err != nil {
		log.Fatal("Connection error: ", err)
	}

	// Wait for disconnect
	<-quit
}
