package common

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}



// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	bets   []domain.Bet
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, bets []domain.Bet) *Client {
	client := &Client{
		config: config,
		bets:   bets,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    sigc := make(chan os.Signal, 1)
    signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

    // Graceful shutdown
    go func() {
        <-sigc
        cancel()
    }()
	
	c.createClientSocket()

	for _, bet := range c.bets {
		select {
        case <-ctx.Done():
            log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
            return
        default:
        }

		err := protocol.SendMsg(c.conn, protocol.BetType, bet.SerializeToCSV())
		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		msgType, _, err := protocol.ReadMsg(c.conn)
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		} else if msgType != protocol.AckType {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: unexpected message type %v",
				c.config.ID,
				msgType,
			)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | client_id: %v | dni: %v | numero: %v",
		c.config.ID,
		bet.Document,
		bet.Number,
	)
	
	// // Wait a time between sending one message and the next one
	// time.Sleep(c.config.LoopPeriod)
	
	}
	c.conn.Close()
	log.Infof("action: connection_closed | result: success | client_id: %v", c.config.ID)
}
