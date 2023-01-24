package xoxo

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ascii8/nakama-go"
	"github.com/google/uuid"
	"github.com/rs/xid"
)

type Client struct {
	cl       *nakama.Client
	conn     *nakama.Conn
	debug    bool
	userId   string
	username string
	logf     func(string, ...interface{})
	persist  bool

	ticketId string
	matchId  string
	state    *MatchState
	waiting  bool

	rw sync.RWMutex
}

func NewClient(opts ...Option) *Client {
	cl := &Client{
		logf:    func(string, ...interface{}) {},
		waiting: true,
	}
	cl.cl = nakama.New(
		nakama.WithURL("http://127.0.0.1:7352"),
		nakama.WithServerKey("xoxo-go_server"),
		nakama.WithAuthHandler(cl),
	)
	for _, o := range opts {
		o(cl)
	}
	if cl.debug {
		nakama.WithTransport(&http.Transport{
			DisableCompression: true,
		})(cl.cl)
		nakama.WithLogger(cl.logf)(cl.cl)
	}
	if cl.userId == "" {
		cl.userId = uuid.New().String()
	}
	if cl.username == "" {
		cl.username = xid.New().String()
	}
	return cl
}

func Dial(ctx context.Context, opts ...Option) (*Client, error) {
	cl := NewClient(opts...)
	if err := cl.Open(ctx); err != nil {
		return nil, err
	}
	return cl, nil
}

func (cl *Client) Open(ctx context.Context) error {
	conn := cl.conn
	if conn == nil {
		opts := []nakama.ConnOption{
			nakama.WithConnHandler(cl),
			nakama.WithConnPersist(cl.persist),
		}
		if cl.debug {
			opts = append(opts, nakama.WithConnFormat("json"))
		}
		var err error
		if conn, err = cl.cl.NewConn(ctx, opts...); err != nil {
			return err
		}
		cl.conn = conn
	}
	return nil
}

func (cl *Client) Close() error {
	_ = cl.Leave(context.Background())
	cl.rw.Lock()
	defer cl.rw.Unlock()
	if cl.conn != nil {
		_ = cl.conn.CloseWithStopErr(true, nil)
	}
	cl.state = nil
	return nil
}

func (cl *Client) Connected() bool {
	conn := cl.conn
	return conn != nil && conn.Connected()
}

func (cl *Client) State() *MatchState {
	return cl.state
}

func (cl *Client) Ready(ctx context.Context) bool {
	ch := make(chan bool, 1)
	go func() {
		defer close(ch)
		for {
			if state := cl.state; state != nil && state.State.RematchCountdown == 0 {
				ch <- true
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(50 * time.Millisecond):
			}
		}
	}()
	select {
	case <-ctx.Done():
		return false
	case res := <-ch:
		return res
	}
}

func (cl *Client) Next(ctx context.Context) bool {
	ch := make(chan bool, 1)
	go func() {
		defer close(ch)
		for {
			cl.rw.RLock()
			waiting, state := cl.waiting, cl.state
			cl.rw.RUnlock()
			switch {
			case waiting || state == nil:
			case state.State.RematchCountdown != 0:
				return
			case state.YourTurn:
				ch <- true
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()
	select {
	case <-ctx.Done():
		return false
	case res := <-ch:
		return res
	}
}

func (cl *Client) AuthHandler(ctx context.Context, nakamaClient *nakama.Client) error {
	return nakamaClient.AuthenticateDevice(ctx, cl.userId, true, cl.username)
}

func (cl *Client) ConnectHandler(ctx context.Context) {
	cl.logf("Connect!")
}

func (cl *Client) DisconnectHandler(ctx context.Context, err error) {
	cl.logf("Disconnect: %v", err)
}

func (cl *Client) ErrorHandler(ctx context.Context, msg *nakama.ErrorMsg) {
	cl.logf("ErrorHandler: %+v", msg)
}

func (cl *Client) ChannelMessageHandler(ctx context.Context, msg *nakama.ChannelMessageMsg) {
	cl.logf("ChannelMessage: %+v", msg)
}

func (cl *Client) ChannelPresenceEventHandler(ctx context.Context, msg *nakama.ChannelPresenceEventMsg) {
	cl.logf("ChannelPresenceEvent: %+v", msg)
}

func (cl *Client) MatchDataHandler(ctx context.Context, msg *nakama.MatchDataMsg) {
	cl.logf("MatchData: %+v", msg)
	state := new(MatchState)
	if err := state.Unmarshal(msg.Data); err != nil {
		cl.logf("unable to unmarshal MatchData: %v", err)
		state = nil
	}
	cl.rw.Lock()
	defer cl.rw.Unlock()
	cl.waiting, cl.state = state == nil, state
}

func (cl *Client) MatchPresenceEventHandler(ctx context.Context, msg *nakama.MatchPresenceEventMsg) {
	cl.logf("MatchPresenceEvent: %+v", msg)
}

func (cl *Client) MatchmakerMatchedHandler(ctx context.Context, msg *nakama.MatchmakerMatchedMsg) {
	cl.logf("MatchmakerMatched: %+v", msg)
	matchId := msg.GetMatchId()
	cl.logf("MatchmakerMatched: joining match %q", matchId)
	cl.conn.MatchJoinAsync(ctx, matchId, nil, func(msg *nakama.MatchMsg, err error) {
		switch {
		case err != nil:
			cl.logf("error: MatchmakerMatched: unable to join match: %v", err)
		default:
			cl.rw.Lock()
			defer cl.rw.Unlock()
			cl.matchId = msg.GetMatchId()
			cl.logf("MatchmakerMatched: joined match %q", cl.matchId)
		}
	})
}

func (cl *Client) NotificationsHandler(ctx context.Context, msg *nakama.NotificationsMsg) {
	cl.logf("Notifications: %+v", msg)
}

func (cl *Client) StatusPresenceEventHandler(ctx context.Context, msg *nakama.StatusPresenceEventMsg) {
	cl.logf("StatusPresenceEvent: %+v", msg)
}

func (cl *Client) StreamDataHandler(ctx context.Context, msg *nakama.StreamDataMsg) {
	cl.logf("StreamData: %+v", msg)
}

func (cl *Client) StreamPresenceEventHandler(ctx context.Context, msg *nakama.StreamPresenceEventMsg) {
	cl.logf("StreamPresence: %+v", msg)
}

func (cl *Client) Join(ctx context.Context) error {
	cl.logf("Join: joining match")
	cl.conn.MatchmakerAddAsync(ctx, nakama.MatchmakerAdd("*", 2, 2), func(msg *nakama.MatchmakerTicketMsg, err error) {
		switch {
		case err != nil:
			cl.logf("Join: unable to join match: %v", err)
		default:
			cl.rw.Lock()
			defer cl.rw.Unlock()
			ticketId := msg.GetTicket()
			cl.logf("Join: added matchmaker ticket %q", ticketId)
			cl.ticketId = ticketId
		}
	})
	return nil
}

func (cl *Client) Leave(ctx context.Context) error {
	cl.logf("Leave: leaving match")
	cl.rw.Lock()
	defer cl.rw.Unlock()
	if cl.ticketId != "" {
		cl.conn.MatchmakerRemoveAsync(ctx, cl.ticketId, nil)
	}
	if cl.matchId != "" {
		cl.conn.MatchLeaveAsync(ctx, cl.matchId, nil)
	}
	cl.ticketId, cl.matchId, cl.waiting, cl.state = "", "", true, nil
	return nil
}

func (cl *Client) Move(ctx context.Context, row, col int) error {
	cl.logf("Move: moving %d, %d", row, col)
	cl.rw.RLock()
	matchId, state := cl.matchId, cl.state
	cl.rw.RUnlock()
	if matchId == "" || state == nil {
		return fmt.Errorf("no active match")
	}
	data, err := NewMove(row, col).Marshal()
	if err != nil {
		return fmt.Errorf("unable to marshal move: %w", err)
	}
	cl.rw.Lock()
	defer cl.rw.Unlock()
	cl.waiting = true
	return cl.conn.MatchDataSend(ctx, matchId, OpCodeMove, data, true, nil)
}

type Option func(*Client)

func WithServerKey(serverKey string) Option {
	return func(cl *Client) {
		nakama.WithServerKey(serverKey)(cl.cl)
	}
}

func WithURL(urlstr string) Option {
	return func(cl *Client) {
		nakama.WithURL(urlstr)(cl.cl)
	}
}

func WithLogf(logf func(string, ...interface{})) Option {
	return func(cl *Client) {
		cl.logf = logf
	}
}

func WithUserId(userId string) Option {
	return func(cl *Client) {
		cl.userId = userId
	}
}

func WithUsername(username string) Option {
	return func(cl *Client) {
		cl.username = username
	}
}

func WithDebug() Option {
	return func(cl *Client) {
		cl.debug = true
	}
}

func WithPersist() Option {
	return func(cl *Client) {
		cl.persist = true
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
