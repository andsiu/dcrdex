// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package comms

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"decred.org/dcrdex/dex"
	"decred.org/dcrdex/dex/msgjson"
	"github.com/gorilla/websocket"
)

const (
	// bufferSize is buffer size for a websocket connection's read channel.
	readBuffSize = 128

	// The maximum time in seconds to write to a connection.
	writeWait = time.Second * 3

	// reconnetInterval is the initial and increment between reconnect tries.
	reconnectInterval = 5 * time.Second

	// maxReconnetInterval is the maximum allowed reconnect interval.
	maxReconnectInterval = time.Minute

	// DefaultResponseTimeout is the default timeout for responses after a
	// request is successfully sent.
	DefaultResponseTimeout = 30 * time.Second
)

// ConnectionStatus represents the current status of the websocket connection.
type ConnectionStatus uint32

const (
	Disconnected ConnectionStatus = iota
	Connected
	InvalidCert
)

// ErrInvalidCert is the error returned when attempting to use an invalid cert
// to set up a ws connection.
var ErrInvalidCert = fmt.Errorf("invalid certificate")

// ErrCertRequired is the error returned when a ws connection fails because no
// cert was provided.
var ErrCertRequired = fmt.Errorf("certificate required")

// WsConn is an interface for a websocket client.
type WsConn interface {
	NextID() uint64
	IsDown() bool
	Send(msg *msgjson.Message) error
	Request(msg *msgjson.Message, respHandler func(*msgjson.Message)) error
	RequestWithTimeout(msg *msgjson.Message, respHandler func(*msgjson.Message), expireTime time.Duration, expire func()) error
	Connect(ctx context.Context) (*sync.WaitGroup, error)
	MessageSource() <-chan *msgjson.Message
}

// When the DEX sends a request to the client, a responseHandler is created
// to wait for the response.
type responseHandler struct {
	expiration *time.Timer
	f          func(*msgjson.Message)
}

// WsCfg is the configuration struct for initializing a WsConn.
type WsCfg struct {
	// URL is the websocket endpoint URL.
	URL string

	// The maximum time in seconds to wait for a ping from the server. This
	// should be larger than the server's ping interval to allow for network
	// latency.
	PingWait time.Duration

	// The server's certificate.
	Cert []byte

	// ReconnectSync runs the needed reconnection synchronization after
	// a reconnect.
	ReconnectSync func()

	// ConnectEventFunc runs whenever connection status changes.
	//
	// NOTE: Disconnect event notifications may lag behind actual
	// disconnections.
	ConnectEventFunc func(ConnectionStatus)

	// Logger is the logger for the WsConn.
	Logger dex.Logger

	// NetDialContext specifies an optional dialer context to use.
	NetDialContext func(context.Context, string, string) (net.Conn, error)
}

// wsConn represents a client websocket connection.
type wsConn struct {
	// 64-bit atomic variables first. See
	// https://golang.org/pkg/sync/atomic/#pkg-note-BUG.
	rID    uint64
	cancel context.CancelFunc
	wg     sync.WaitGroup
	log    dex.Logger
	cfg    *WsCfg
	tlsCfg *tls.Config
	readCh chan *msgjson.Message

	wsMtx sync.Mutex
	ws    *websocket.Conn

	connectionStatus uint32 // atomic

	reqMtx       sync.RWMutex
	respHandlers map[uint64]*responseHandler

	reconnectCh chan struct{} // trigger for immediate reconnect
}

// NewWsConn creates a client websocket connection.
func NewWsConn(cfg *WsCfg) (WsConn, error) {
	if cfg.PingWait < 0 {
		return nil, fmt.Errorf("ping wait cannot be negative")
	}

	var tlsConfig *tls.Config
	if len(cfg.Cert) > 0 {

		uri, err := url.Parse(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("error parsing URL: %w", err)
		}

		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		if ok := rootCAs.AppendCertsFromPEM(cfg.Cert); !ok {
			return nil, ErrInvalidCert
		}

		tlsConfig = &tls.Config{
			RootCAs:    rootCAs,
			MinVersion: tls.VersionTLS12,
			ServerName: uri.Hostname(),
		}
	}

	return &wsConn{
		cfg:          cfg,
		log:          cfg.Logger,
		tlsCfg:       tlsConfig,
		readCh:       make(chan *msgjson.Message, readBuffSize),
		respHandlers: make(map[uint64]*responseHandler),
		reconnectCh:  make(chan struct{}, 1),
	}, nil
}

// IsDown indicates if the connection is known to be down.
func (conn *wsConn) IsDown() bool {
	return atomic.LoadUint32(&conn.connectionStatus) != uint32(Connected)
}

// setConnectionStatus updates the connection's status and runs the
// ConnectEventFunc in case of a change.
func (conn *wsConn) setConnectionStatus(status ConnectionStatus) {
	oldStatus := atomic.SwapUint32(&conn.connectionStatus, uint32(status))
	statusChange := oldStatus != uint32(status)
	if statusChange && conn.cfg.ConnectEventFunc != nil {
		conn.cfg.ConnectEventFunc(status)
	}
}

// connect attempts to establish a websocket connection.
func (conn *wsConn) connect(ctx context.Context) error {
	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  conn.tlsCfg,
	}
	if conn.cfg.NetDialContext != nil {
		dialer.NetDialContext = conn.cfg.NetDialContext
	} else {
		dialer.Proxy = http.ProxyFromEnvironment
	}
	ws, _, err := dialer.DialContext(ctx, conn.cfg.URL, nil)
	if err != nil {
		var e x509.UnknownAuthorityError
		if errors.As(err, &e) {
			conn.setConnectionStatus(InvalidCert)
			if conn.tlsCfg == nil {
				return ErrCertRequired
			}
			return ErrInvalidCert
		}
		conn.setConnectionStatus(Disconnected)
		return err
	}

	// Set the initial read deadline for the first ping. Subsequent read
	// deadlines are set in the ping handler.
	err = ws.SetReadDeadline(time.Now().Add(conn.cfg.PingWait))
	if err != nil {
		conn.log.Errorf("set read deadline failed: %v", err)
		return err
	}

	ws.SetPingHandler(func(string) error {
		now := time.Now()

		// Set the deadline for the next ping.
		err := ws.SetReadDeadline(now.Add(conn.cfg.PingWait))
		if err != nil {
			conn.log.Errorf("set read deadline failed: %v", err)
			return err
		}

		// Respond with a pong.
		err = ws.WriteControl(websocket.PongMessage, []byte{}, now.Add(writeWait))
		if err != nil {
			// read loop handles reconnect
			conn.log.Errorf("pong write error: %v", err)
			return err
		}

		return nil
	})

	conn.wsMtx.Lock()
	// If keepAlive called connect, the wsConn's current websocket.Conn may need
	// to be closed depending on the error that triggered the reconnect.
	if conn.ws != nil {
		conn.close()
	}
	conn.ws = ws
	conn.wsMtx.Unlock()

	conn.setConnectionStatus(Connected)
	conn.wg.Add(1)
	go func() {
		defer conn.wg.Done()
		conn.read(ctx)
	}()

	return nil
}

func (conn *wsConn) close() {
	// Attempt to send a close message in case the connection is still live.
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye")
	_ = conn.ws.WriteControl(websocket.CloseMessage, msg,
		time.Now().Add(50*time.Millisecond)) // ignore any error
	// Forcibly close the underlying connection.
	conn.ws.Close()
}

// read fetches and parses incoming messages for processing. This should be
// run as a goroutine. Increment the wg before calling read.
func (conn *wsConn) read(ctx context.Context) {
	reconnect := func() {
		conn.setConnectionStatus(Disconnected)
		conn.reconnectCh <- struct{}{}
	}

	for {
		msg := new(msgjson.Message)

		// Lock since conn.ws may be set by connect.
		conn.wsMtx.Lock()
		ws := conn.ws
		conn.wsMtx.Unlock()

		// The read itself does not require locking since only this goroutine
		// uses read functions that are not safe for concurrent use.
		err := ws.ReadJSON(msg)
		// Drop the read error on context cancellation.
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			// Read timeout should flag the connection as down asap.
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				conn.log.Errorf("Read timeout on connection to %s.", conn.cfg.URL)
				reconnect()
				return
			}

			var mErr *json.UnmarshalTypeError
			if errors.As(err, &mErr) {
				// JSON decode errors are not fatal, log and proceed.
				conn.log.Errorf("json decode error: %v", mErr)
				continue
			}

			// TODO: Now that wsConn goroutines have contexts that are canceled
			// on shutdown, we do not have to infer the source and severity of
			// the error; just reconnect in ALL other cases, and remove the
			// following legacy checks.

			// Expected close errors (1000 and 1001) ... but if the server
			// closes we still want to reconnect. (???)
			if websocket.IsCloseError(err, websocket.CloseGoingAway,
				websocket.CloseNormalClosure) ||
				strings.Contains(err.Error(), "websocket: close sent") {
				reconnect()
				return
			}

			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Op == "read" {
				if strings.Contains(opErr.Err.Error(),
					"use of closed network connection") {
					conn.log.Errorf("read quitting: %v", err)
					reconnect()
					return
				}
			}

			// Log all other errors and trigger a reconnection.
			conn.log.Errorf("read error (%v), attempting reconnection", err)
			reconnect()
			// Successful reconnect via connect() will start read() again.
			return
		}

		// If the message is a response, find the handler.
		if msg.Type == msgjson.Response {
			handler := conn.respHandler(msg.ID)
			if handler == nil {
				b, _ := json.Marshal(msg)
				conn.log.Errorf("No handler found for response: %v", string(b))
				continue
			}
			// Run handlers in a goroutine so that other messages can be
			// received. Include the handler goroutines in the WaitGroup to
			// allow them to complete if the connection master desires.
			conn.wg.Add(1)
			go func() {
				defer conn.wg.Done()
				handler.f(msg)
			}()
			continue
		}
		conn.readCh <- msg
	}
}

// keepAlive maintains an active websocket connection by reconnecting when
// the established connection is broken. This should be run as a goroutine.
func (conn *wsConn) keepAlive(ctx context.Context) {
	rcInt := reconnectInterval
	for {
		select {
		case <-conn.reconnectCh:
			// Prioritize context cancellation even if there are reconnect
			// requests.
			if ctx.Err() != nil {
				return
			}

			conn.log.Infof("Attempting to reconnect to %s...", conn.cfg.URL)
			err := conn.connect(ctx)
			if err != nil {
				conn.log.Errorf("Reconnect failed. Scheduling reconnect to %s in %.1f seconds.",
					conn.cfg.URL, rcInt.Seconds())
				time.AfterFunc(rcInt, func() {
					conn.reconnectCh <- struct{}{}
				})
				// Increment the wait up to PingWait.
				if rcInt < maxReconnectInterval {
					rcInt += reconnectInterval
				}
				continue
			}

			conn.log.Info("Successfully reconnected.")
			rcInt = reconnectInterval

			// Synchronize after a reconnection.
			if conn.cfg.ReconnectSync != nil {
				conn.cfg.ReconnectSync()
			}

		case <-ctx.Done():
			return
		}
	}
}

// NextID returns the next request id.
func (conn *wsConn) NextID() uint64 {
	return atomic.AddUint64(&conn.rID, 1)
}

// Connect connects the client. Any error encountered during the initial
// connection will be returned. An auto-(re)connect goroutine will be started,
// even on error. To terminate it, use Stop() or cancel the context.
func (conn *wsConn) Connect(ctx context.Context) (*sync.WaitGroup, error) {
	var ctxInternal context.Context
	ctxInternal, conn.cancel = context.WithCancel(ctx)

	err := conn.connect(ctxInternal)
	if err != nil {
		// If the certificate is invalid or missing, do not start the reconnect
		// loop, and return an error with no WaitGroup.
		if errors.Is(err, ErrInvalidCert) || errors.Is(err, ErrCertRequired) {
			conn.cancel()
			conn.wg.Wait() // probably a no-op
			close(conn.readCh)
			return nil, err
		}

		// The read loop would normally trigger keepAlive, but it wasn't started
		// on account of a connect error.
		conn.log.Errorf("Initial connection failed, starting reconnect loop: %v", err)
		time.AfterFunc(5*time.Second, func() {
			conn.reconnectCh <- struct{}{}
		})
	}

	conn.wg.Add(2)
	go func() {
		defer conn.wg.Done()
		conn.keepAlive(ctxInternal)
	}()
	go func() {
		defer conn.wg.Done()
		<-ctxInternal.Done()
		conn.setConnectionStatus(Disconnected)
		conn.wsMtx.Lock()
		if conn.ws != nil {
			conn.log.Debug("Sending close 1000 (normal) message.")
			conn.close()
		}
		conn.wsMtx.Unlock()

		close(conn.readCh) // signal to MessageSource receivers that the wsConn is dead
	}()

	return &conn.wg, err
}

// Stop can be used to close the connection and all of the goroutines started by
// Connect. Alternatively, the context passed to Connect may be canceled.
func (conn *wsConn) Stop() {
	conn.cancel()
}

func (conn *wsConn) SendRaw(b []byte) error {
	conn.wsMtx.Lock()
	defer conn.wsMtx.Unlock()
	return conn.ws.WriteMessage(websocket.TextMessage, b)
}

// Send pushes outgoing messages over the websocket connection. Sending of the
// message is synchronous, so a nil error guarantees that the message was
// successfully sent. A non-nil error may indicate that the connection is known
// to be down, the message failed to marshall to JSON, or writing to the
// websocket link failed.
func (conn *wsConn) Send(msg *msgjson.Message) error {
	if conn.IsDown() {
		return fmt.Errorf("cannot send on a broken connection")
	}

	// Marshal the Message first so that we don't send junk to the peer even if
	// it fails to marshal completely, which gorilla/websocket.WriteJSON does.
	b, err := json.Marshal(msg)
	if err != nil {
		conn.log.Errorf("Failed to marshal message: %v", err)
		return err
	}

	conn.wsMtx.Lock()
	defer conn.wsMtx.Unlock()
	err = conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		conn.log.Errorf("Send: failed to set write deadline: %v", err)
		return err
	}

	err = conn.ws.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		conn.log.Errorf("Send: WriteMessage error: %v", err)
		return err
	}
	return nil
}

// Request sends the Request-type msgjson.Message to the server and does not
// wait for a response, but records a callback function to run when a response
// is received. A response must be received within DefaultResponseTimeout of the
// request, after which the response handler expires and any late response will
// be ignored. To handle expiration or to set the timeout duration, use
// RequestWithTimeout. Sending of the request is synchronous, so a nil error
// guarantees that the request message was successfully sent.
func (conn *wsConn) Request(msg *msgjson.Message, f func(*msgjson.Message)) error {
	return conn.RequestWithTimeout(msg, f, DefaultResponseTimeout, func() {})
}

// RequestWithTimeout sends the Request-type message and does not wait for a
// response, but records a callback function to run when a response is received.
// If the server responds within expireTime of the request, the response handler
// is called, otherwise the expire function is called. If the response handler
// is called, it is guaranteed that the response Message.ID is equal to the
// request Message.ID. Sending of the request is synchronous, so a nil error
// guarantees that the request message was successfully sent and that either the
// response handler or expire function will be run; a non-nil error guarantees
// that neither function will run.
//
// For example, to wait on a response or timeout:
//
// errChan := make(chan error, 1)
// err := conn.RequestWithTimeout(reqMsg, func(msg *msgjson.Message) {
//     errChan <- msg.UnmarshalResult(responseStructPointer)
// }, timeout, func() {
//     errChan <- fmt.Errorf("timed out waiting for '%s' response.", route)
// })
// if err != nil {
//     return err // request error
// }
// return <-errChan // timeout or response error
func (conn *wsConn) RequestWithTimeout(msg *msgjson.Message, f func(*msgjson.Message), expireTime time.Duration, expire func()) error {
	if msg.Type != msgjson.Request {
		return fmt.Errorf("Message is not a request: %v", msg.Type)
	}
	// Register the response and expire handlers for this request.
	conn.logReq(msg.ID, f, expireTime, expire)
	err := conn.Send(msg)
	if err != nil {
		// Neither expire nor the handler should run. Stop the expire timer
		// created by logReq and delete the response handler it added. The
		// caller receives a non-nil error to deal with it.
		conn.log.Errorf("(*wsConn).Request(route '%s') Send error (%v), unregistering msg ID %d handler",
			msg.Route, err, msg.ID)
		conn.respHandler(msg.ID) // drop the responseHandler logged by logReq that is no longer necessary
	}
	return err
}

func (conn *wsConn) expire(id uint64) bool {
	conn.reqMtx.Lock()
	defer conn.reqMtx.Unlock()
	_, removed := conn.respHandlers[id]
	delete(conn.respHandlers, id)
	return removed
}

// logReq stores the response handler in the respHandlers map. Requests to the
// client are associated with a response handler.
func (conn *wsConn) logReq(id uint64, respHandler func(*msgjson.Message), expireTime time.Duration, expire func()) {
	conn.reqMtx.Lock()
	defer conn.reqMtx.Unlock()
	doExpire := func() {
		// Delete the response handler, and call the provided expire function if
		// (*wsLink).respHandler has not already retrieved the handler function
		// for execution.
		if conn.expire(id) {
			expire()
		}
	}
	conn.respHandlers[id] = &responseHandler{
		expiration: time.AfterFunc(expireTime, doExpire),
		f:          respHandler,
	}
}

// respHandler extracts the response handler for the provided request ID if it
// exists, else nil. If the handler exists, it will be deleted from the map.
func (conn *wsConn) respHandler(id uint64) *responseHandler {
	conn.reqMtx.Lock()
	defer conn.reqMtx.Unlock()
	cb, ok := conn.respHandlers[id]
	if ok {
		cb.expiration.Stop()
		delete(conn.respHandlers, id)
	}
	return cb
}

// MessageSource returns the connection's read source. The returned chan will
// receive requests and notifications from the server, but not responses, which
// have handlers associated with their request. The same channel is returned on
// each call, so there must only be one receiver. When the connection is
// shutdown, the channel will be closed.
func (conn *wsConn) MessageSource() <-chan *msgjson.Message {
	return conn.readCh
}
