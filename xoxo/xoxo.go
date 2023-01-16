package xoxo

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	OpCodeMove  = 1
	OpCodeState = 2
)

type Player struct {
	Node      string `json:"node,omitempty"`
	SessionId string `json:"session_id,omitempty"`
	UserId    string `json:"user_id,omitempty"`
	Username  string `json:"username,omitempty"`
}

type State struct {
	Cells            [][]int  `json:"cells,omitempty"`
	PlayerTurn       int      `json:"player_turn"`
	Players          []Player `json:"players"`
	Winner           int      `json:"winner,omitempty"`
	Draw             bool     `json:"draw,omitempty"`
	RematchCountdown int      `json:"rematch_countdown,omitempty"`
}

func NewState() *State {
	cells := make([][]int, 3)
	for i := 0; i < 3; i++ {
		cells[i] = []int{-1, -1, -1}
	}
	return &State{
		Cells:      cells,
		PlayerTurn: 1,
	}
}

func (s *State) Add(node, sessionId, userId, username string) error {
	if len(s.Players) == 2 {
		return fmt.Errorf("cannot have more than 2 players in a game")
	}
	for _, p := range s.Players {
		if p.UserId == userId {
			return fmt.Errorf("player %s already added", p.UserId)
		}
	}
	s.Players = append(s.Players, Player{
		Node:      node,
		SessionId: sessionId,
		UserId:    userId,
		Username:  username,
	})
	return nil
}

func (s *State) Move(userId string, move Move) error {
	row, col := move.Row-1, move.Col-1
	switch {
	case row < 0 || 3 <= row:
		return fmt.Errorf("invalid row %d", row)
	case col < 0 || 3 <= col:
		return fmt.Errorf("invalid col %d", col)
	case s.Cells[row][col] != -1:
		return fmt.Errorf("invalid move at row %d, col %d", row, col)
	}
	i, found := 0, false
	for ; i < len(s.Players); i++ {
		if s.Players[i].UserId == userId {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unable to locate player with user id %q", userId)
	}
	// do move
	s.Cells[row][col] = i + 1
	switch s.PlayerTurn {
	case 1:
		s.PlayerTurn = 2
	case 2:
		s.PlayerTurn = 1
	default:
		return fmt.Errorf("invalid PlayerTurn %d", s.PlayerTurn)
	}
	// determine if there is a winner
	s.Winner = 0
loop:
	for p := 1; p <= 2; p++ {
		for i := 0; i < 8; i++ {
			if isWinner(p, s.Cells, coords[i]) {
				s.Winner = p
				break loop
			}
		}
	}
	s.Draw = true
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			s.Draw = s.Draw && s.Cells[i][j] != -1
		}
	}
	return nil
}

func (s *State) String() string {
	v := make([]interface{}, 9)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			v[i*3+j] = getCellAsRune(i, j, s.Cells)
		}
	}
	switch {
	case s.Winner != 0:
		v = append(v, fmt.Sprintf(" (completed, winner: %d)", s.Winner))
	case s.Draw:
		v = append(v, " (completed, draw)")
	}
	return fmt.Sprintf("[%c%c%c,%c%c%c,%c%c%c]%s", v...)
}

func getCellAsRune(i, j int, cells [][]int) rune {
	switch cells[i][j] {
	case 1:
		return 'X'
	case 2:
		return 'O'
	}
	return '.'
}

type MatchState struct {
	ActivePlayer *Player `json:"active_player,omitempty"`
	OtherPlayer  *Player `json:"other_player,omitempty"`
	State        *State  `json:"state,omitempty"`
	YourTurn     bool    `json:"your_turn"`
}

func (m *MatchState) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *MatchState) Unmarshal(buf []byte) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.DisallowUnknownFields()
	return dec.Decode(m)
}

type Move struct {
	Row int `json:"row,omitempty"`
	Col int `json:"col,omitempty"`
}

func NewMove(row, col int) Move {
	return Move{
		Row: row + 1,
		Col: col + 1,
	}
}

func (m Move) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *Move) Unmarshal(buf []byte) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	dec.DisallowUnknownFields()
	return dec.Decode(m)
}

func isWinner(p int, c [][]int, w [6]int) bool {
	return c[w[0]][w[1]] == p &&
		c[w[2]][w[3]] == p &&
		c[w[4]][w[5]] == p
}

var coords = [8][6]int{
	{0, 0, 0, 1, 0, 2}, // row 0
	{1, 0, 1, 1, 1, 2}, // row 1
	{2, 0, 2, 1, 2, 2}, // row 2
	{0, 0, 1, 0, 2, 0}, // col 0
	{0, 1, 1, 1, 2, 1}, // col 1
	{0, 2, 1, 2, 2, 2}, // col 2
	{0, 0, 1, 1, 2, 2}, // top left to bottom right
	{2, 0, 1, 1, 0, 2}, // bottom left to top right
}
