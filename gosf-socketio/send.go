package gosocketio

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/tobiasbeck/feathers-go/gosf-socketio/protocol"
)

var (
	ErrorSendTimeout     = errors.New("Timeout")
	ErrorSocketOverflood = errors.New("Socket overflood")
)

/**
Send message packet to socket
*/
func send(msg *protocol.Message, c *Channel, args interface{}, err error) error {
	//preventing json/encoding "index out of range" panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("socket.io send panic: ", r)
		}
	}()
	if msg.Type != protocol.MessageTypeEmit {
		if args != nil {
			json, err := json.Marshal(&args)
			if err != nil {
				return err
			}
			msg.Args = "null," + string(json)
		} else if err != nil {
			json, err := json.Marshal(&err)
			if err != nil {
				return err
			}
			msg.Args = string(json)
		}
	} else {
		json, err := json.Marshal(&args)
		if err != nil {
			return err
		}
		msg.Args = string(json)
	}

	command, err := protocol.Encode(msg)
	if err != nil {
		return err
	}

	if len(c.out) == queueBufferSize {
		return ErrorSocketOverflood
	}
	c.out <- command

	return nil
}

/**
Create packet based on given data and send it
*/
func (c *Channel) Emit(method string, args interface{}) error {
	msg := &protocol.Message{
		Type:   protocol.MessageTypeEmit,
		Method: method,
	}

	return send(msg, c, args, nil)
}

/**
Create ack packet based on given data and send it and receive response
*/
func (c *Channel) Ack(method string, args []interface{}, timeout time.Duration) (string, error) {
	msg := &protocol.Message{
		Type:   protocol.MessageTypeAckRequest,
		AckId:  c.ack.getNextId(),
		Method: method,
	}

	waiter := make(chan string)
	c.ack.addWaiter(msg.AckId, waiter)

	err := send(msg, c, args, nil)
	if err != nil {
		c.ack.removeWaiter(msg.AckId)
	}

	select {
	case result := <-waiter:
		return result, nil
	case <-time.After(timeout):
		c.ack.removeWaiter(msg.AckId)
		return "", ErrorSendTimeout
	}
}
