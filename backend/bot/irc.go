// Purpose: Minimal Twitch IRC client.

package bot

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type IRCConn struct {
	c   net.Conn
	r   *bufio.Reader
	w   *bufio.Writer
	nick string
}

func DialIRC(nick, oauth string) (*IRCConn, error) {
	// Twitch IRC TLS: irc.chat.twitch.tv:6697
	d := &net.Dialer{Timeout: 10 * time.Second}
	c, err := tls.DialWithDialer(d, "tcp", "irc.chat.twitch.tv:6697", &tls.Config{})
	if err != nil {
		return nil, err
	}
	irc := &IRCConn{
		c:    c,
		r:    bufio.NewReader(c),
		w:    bufio.NewWriter(c),
		nick: nick,
	}

	// PASS + NICK
	if err := irc.writeLine("PASS " + oauth); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := irc.writeLine("NICK " + nick); err != nil {
		_ = c.Close()
		return nil, err
	}
	// Request tags (optional)
	_ = irc.writeLine("CAP REQ :twitch.tv/tags")

	return irc, nil
}

func (i *IRCConn) Close() error {
	return i.c.Close()
}

func (i *IRCConn) Join(channel string) error {
	channel = strings.TrimPrefix(channel, "#")
	if channel == "" {
		return errors.New("empty channel")
	}
	return i.writeLine("JOIN #" + channel)
}

func (i *IRCConn) Say(channel, text string) error {
	channel = strings.TrimPrefix(channel, "#")
	if channel == "" {
		return errors.New("empty channel")
	}
	// PRIVMSG #chan :text
	return i.writeLine(fmt.Sprintf("PRIVMSG #%s :%s", channel, text))
}

type IRCMessage struct {
	Raw  string
	Text string
}

func (i *IRCConn) ReadMessage() (IRCMessage, error) {
	line, err := i.r.ReadString('\n')
	if err != nil {
		return IRCMessage{}, err
	}
	line = strings.TrimRight(line, "\r\n")

	// Respond to PING
	if strings.HasPrefix(line, "PING ") {
		_ = i.writeLine("PONG " + strings.TrimPrefix(line, "PING "))
	}

	// Extract trailing text after " :"
	text := ""
	if idx := strings.Index(line, " :"); idx >= 0 && idx+2 < len(line) {
		text = line[idx+2:]
	}

	return IRCMessage{Raw: line, Text: text}, nil
}

func (i *IRCConn) writeLine(s string) error {
	if _, err := i.w.WriteString(s + "\r\n"); err != nil {
		return err
	}
	return i.w.Flush()
}
