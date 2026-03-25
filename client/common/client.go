package common

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	shouldReturn := sendBets(ctx, c)
	if shouldReturn {
		return
	}

	winners, shouldReturn := handleWinnersMessageExchange(c)
	if shouldReturn {
		return
	}

	log.Infof("action: consulta_ganadores | result: success | client_id: %v | cant_ganadores: %v", c.config.ID, len(winners))

	c.conn.Close()
	log.Infof("action: connection_closed | result: success | client_id: %v", c.config.ID)
}

func sendBets(ctx context.Context, c *Client) bool {
	file, err := os.Open(c.config.BetsFilePath)
	if err != nil {
		log.Criticalf("action: open_bets_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return true
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var batch []byte
	batchCount := 0
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
			return true
		default:
		}

		lineCopy := append([]byte(nil), scanner.Bytes()...)

		if len(batch)+len(lineCopy)+1 > protocol.MaxPayloadSize {
			if err := handleSendBetsMessageExchange(c, batch, false); err != nil {
				batchCount = 0
				return true
			}
			batch = nil
			batchCount = 0
		}

		if batchCount == c.config.MaxBatchSize {
			if err := handleSendBetsMessageExchange(c, batch, false); err != nil {
				batchCount = 0
				return true
			}
			batch = nil
			batchCount = 0
		}

		batch = append(batch, lineCopy...)
		batch = append(batch, '\n')
		batchCount++
	}


	if err := handleSendBetsMessageExchange(c, batch, true); err != nil {
		return true
	}
	return false
}

func handleSendBetsMessageExchange(c *Client, payload []byte, eof bool) error {
	agencyID, err := strconv.Atoi(c.config.ID)
	if err != nil {
		log.Errorf("action: convert_agency_id | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}
	if err := protocol.SendMsg(c.conn, protocol.BetType, uint16(agencyID), eof, payload); err != nil {
		log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	msgType, _, _, ackPayload, err := protocol.ReadMsg(c.conn)
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

func handleWinnersMessageExchange(c *Client) ([]string, bool) {
	agencyID, err := strconv.Atoi(c.config.ID)
	if err != nil {
		log.Errorf("action: convert_agency_id | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return nil, true
	}

	err = protocol.SendMsg(c.conn, protocol.WinnerQueryType, uint16(agencyID), false, nil)
	if err != nil {
		log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return nil, true
	}

	eof := false
	var winners []string
	for !eof {

		msgType, _, eofFlag, payload, err := protocol.ReadMsg(c.conn)
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return nil, true
		}
		
		if msgType != protocol.WinnerResponseType {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: unexpected message type %v", c.config.ID, msgType)
			return nil, true
		}
		eof = eofFlag
		payloadStr := string(payload)
		lines := strings.Split(payloadStr, "\n")
		winners = append(winners, lines...)
	}
	return winners, false
}
