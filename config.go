package ircx

import "crypto/tls"

type Config func(*Bot)

func WithLogin(user, pass string) Config {
	return func(b *Bot) {
		b.user = user
		b.password = pass
	}
}

func WithTLS(config *tls.Config) Config {
	return func(b *Bot) {
		b.tlsConfig = config
	}
}
