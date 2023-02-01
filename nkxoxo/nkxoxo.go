package nkxoxo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ascii8/xoxo-go/xoxo"
	"github.com/heroiclabs/nakama-common/runtime"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const tickRate = 1

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.
		WithField("date", time.Now()).
		Debug("backend loaded")
	if err := initializer.RegisterMatch("xoxo", newMatch); err != nil {
		return err
	}
	if err := initializer.RegisterMatchmakerMatched(matchmakerMatched); err != nil {
		return err
	}
	return nil
}

func matchmakerMatched(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
	logger.Debug("creating xoxo match")
	for i, entry := range entries {
		l := logger
		properties := entry.GetProperties()
		keys := maps.Keys(properties)
		slices.Sort(keys)
		for _, k := range keys {
			l = l.WithField(k, properties[k])
		}
		l.Debug(fmt.Sprintf("matched user %d", i))
	}
	return nk.MatchCreate(ctx, "xoxo", map[string]interface{}{
		"invited": entries,
	})
}

type match struct{}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	return match{}, nil
}

func (m match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	logger.
		Debug("MatchInit")
	return newMatchState(), tickRate, ""
}

func (m match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	logger.
		WithField("tick", tick).
		WithField("presence", presence).
		Debug("MatchJoinAttempt")
	s := state.(*matchState)
	if err := s.add(presence); err != nil {
		return s, false, err.Error()
	}
	return s, true, ""
}

func (m match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	logger.
		WithField("tick", tick).
		WithField("presences", len(presences)).
		Debug("MatchJoin")
	s := state.(*matchState)
	if len(s.presences) == 2 {
		if err := s.broadcastState(logger, dispatcher); err != nil {
			logger.
				WithField("tick", tick).
				WithField("error", err).
				Debug("MatchJoin unable to broadcast state")
		}
	}
	return s
}

func (m match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	logger.
		WithField("tick", tick).
		Debug("MatchLeave")
	s := state.(*matchState)
	if err := s.broadcastState(logger, dispatcher); err != nil {
		logger.
			WithField("tick", tick).
			WithField("error", err).
			Debug("MatchJoin unable to broadcast state")
	}
	s.termTick = tick
	return s
}

func (m match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	s := state.(*matchState)
	l := logger.WithField("tick", tick)
	switch {
	case len(s.presences) != 2 && s.termTick == 0:
		l.
			Debug("MatchLoop waiting for players")
		return s
	case s.termTick != 0 && tick-5 > s.termTick:
		l.Debug("MatchLoop terminating")
		return nil
	}
	for _, m := range messages {
		data, userId := m.GetData(), m.GetUserId()
		l := l.WithField("user_id", userId)
		l.
			WithField("data", data).
			Debug("MatchLoop received message")
		if m.GetOpCode() == xoxo.OpCodeMove {
			var move xoxo.Move
			if err := move.Unmarshal(data); err != nil {
				l.
					WithField("data", data).
					WithField("error", err).
					Debug("MessageLoop unable to decode message")
				continue
			}
			l = l.WithField("move", move)
			l.
				WithField("state", s.state.String()).
				Debug("MatchLoop move")
			if err := s.state.Move(userId, move); err != nil {
				l.
					WithField("error", err).
					Debug("MessageLoop unable to move")
			}
			// ended
			if s.state.Winner != 0 || s.state.Draw {
				s.state.RematchCountdown = 10 * tickRate
			}
			if err := s.broadcastState(logger, dispatcher); err != nil {
				l.
					WithField("error", err).
					Debug("MatchLoop unable to broadcast state")
			}
		}
	}
	if s.state.RematchCountdown > 0 {
		s.state.RematchCountdown--
		if s.state.RematchCountdown == 0 {
			s.rematch()
		}
		if err := s.broadcastState(logger, dispatcher); err != nil {
			l.
				WithField("error", err).
				Debug("MatchLoop unable to broadcast state")
		}
	}
	return s
}

func (m match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	logger.
		WithField("tick", tick).
		Debug("MatchTerminate")
	return nil
}

func (m match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

type matchState struct {
	state     *xoxo.State
	presences []runtime.Presence
	termTick  int64
}

func newMatchState() *matchState {
	return &matchState{
		state: xoxo.NewState(),
	}
}

func (s *matchState) rematch() {
	s.state = xoxo.NewState()
	if err := s.state.Add(
		s.presences[0].GetNodeId(),
		s.presences[0].GetSessionId(),
		s.presences[0].GetUserId(),
		s.presences[0].GetUsername(),
	); err != nil {
		panic(err)
	}
	if err := s.state.Add(
		s.presences[1].GetNodeId(),
		s.presences[1].GetSessionId(),
		s.presences[1].GetUserId(),
		s.presences[1].GetUsername(),
	); err != nil {
		panic(err)
	}
}

func (s *matchState) add(presence runtime.Presence) error {
	if len(s.presences) == 2 {
		return fmt.Errorf("cannot have more than 2 players in a game")
	}
	for _, p := range s.presences {
		if p.GetUserId() == presence.GetUserId() {
			return fmt.Errorf("presence %s already added", p.GetUserId())
		}
	}
	if err := s.state.Add(
		presence.GetNodeId(),
		presence.GetSessionId(),
		presence.GetUserId(),
		presence.GetUsername(),
	); err != nil {
		return err
	}
	s.presences = append(s.presences, presence)
	return nil
}

func (s *matchState) broadcastState(logger runtime.Logger, dispatcher runtime.MatchDispatcher) error {
	if len(s.presences) != 2 {
		return fmt.Errorf("invalid presences length %d", len(s.presences))
	}
	logger.
		WithField("state", s.state.String()).
		Debug("broadcast state")
	active, other := &s.state.Players[0], &s.state.Players[1]
	if s.state.PlayerTurn == 2 {
		active, other = other, active
	}
	for i := 0; i < 2; i++ {
		data, err := (&xoxo.MatchState{
			ActivePlayer: active,
			OtherPlayer:  other,
			State:        s.state,
			YourTurn:     s.state.PlayerTurn == (i + 1),
		}).Marshal()
		if err != nil {
			return fmt.Errorf("unable to marshal message for %s: %w", s.presences[i].GetSessionId(), err)
		}
		logger.
			WithField("data", data).
			Debug("sending")
		if err := dispatcher.BroadcastMessage(xoxo.OpCodeState, data, s.presences[i:i+1], nil, true); err != nil {
			return fmt.Errorf("unable to broadcast message for %s: %w", s.presences[i].GetSessionId(), err)
		}
	}
	return nil
}
