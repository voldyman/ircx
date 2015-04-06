package ircx

import "github.com/sorcix/irc"

// connectMessages is a list of IRC messages to send when attempting to
// connect to the IRC server.
func (b *Bot) connectMessages() []*irc.Message {
	messages := []*irc.Message{}
	messages = append(messages, &irc.Message{
		Command:  irc.USER,
		Params:   []string{b.user, "0", "*"},
		Trailing: b.user,
	})
	messages = append(messages, &irc.Message{
		Command: irc.NICK,
		Params:  []string{b.name},
	})
	if b.password != "" {
		messages = append(messages, &irc.Message{
			Command: irc.PASS,
			Params:  []string{b.password},
		})
	}
	return messages
}
