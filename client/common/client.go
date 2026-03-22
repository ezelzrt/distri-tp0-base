package common

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	MaxBatchSize  int
	BetsFilePath  string
}



// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	// bets   []domain.Bet
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
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

	file, err := os.Open(c.config.BetsFilePath)
	if err != nil {
		log.Criticalf("action: open_bets_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var batch []byte
	batchCount := 0
	for scanner.Scan() {
		select {
        case <-ctx.Done():
            log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
            return
        default:
        }

		lineCopy := append([]byte(nil), scanner.Bytes()...)
		
		if len(batch)+len(lineCopy)+1 > protocol.MaxPayloadSize {
			if err := handleMessageExchange(c, batch); err != nil {
				batchCount = 0
				break
			}
			batch = nil
			batchCount = 0
		} 

		batch = append(batch, lineCopy...)
		batch = append(batch, '\n')
		batchCount++
		
		if batchCount == c.config.MaxBatchSize {
			if err := handleMessageExchange(c, batch); err != nil {
				batchCount = 0
				break
			}	
			batch = nil
			batchCount = 0
		}
	}

	if batchCount > 0 {
		handleMessageExchange(c, batch)
	}
	c.conn.Close()
	log.Infof("action: connection_closed | result: success | client_id: %v", c.config.ID)
}

func handleMessageExchange(c *Client, payload []byte) error {
	if err := protocol.SendMsg(c.conn, protocol.BetType, payload); err != nil {
		log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	msgType, ackPayload, err := protocol.ReadMsg(c.conn)
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	if msgType != protocol.AckType {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: unexpected message type %v",
			c.config.ID,
			msgType,
		)
		return fmt.Errorf("unexpected message type %v", msgType)
	}

	if len(ackPayload) == 0 {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: empty ack payload",
			c.config.ID)
		return fmt.Errorf("empty ack payload")
	}

	if ackPayload[0] != '0' {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: server reported failure", c.config.ID)
		return fmt.Errorf("batch processing failed on server")
	}

	return nil
}
