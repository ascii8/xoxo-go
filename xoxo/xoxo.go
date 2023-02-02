package xoxo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	OpCodeMove  = 1
	OpCodeState = 2
)

type Winner int

func (w *Winner) UnmarshalJSON(buf []byte) error {
	switch string(buf) {
	case "1":
		*w = 1
	case "2":
		*w = 2
	case "false":
		*w = 0
	default:
		return fmt.Errorf("invalid winner %q", buf)
	}
	return nil
}

func (w Winner) MarshalJSON() ([]byte, error) {
	if w == 0 {
		return []byte("false"), nil
	}
	return []byte(strconv.Itoa(w.Int())), nil
}

func (w Winner) Int() int {
	return int(w)
}

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
	Winner           Winner   `json:"winner,omitempty"`
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
		return fmt.Errorf("invalid row %d (%d)", move.Row, row)
	case col < 0 || 3 <= col:
		return fmt.Errorf("invalid col %d (%d)", move.Col, col)
	case s.Cells[row][col] != -1:
		return fmt.Errorf("invalid move at row %d, col %d (%d, %d)", move.Row, move.Col, row, col)
	case s.Winner != 0:
		return fmt.Errorf("match already won by player %d", s.Winner)
	case s.Draw:
		return fmt.Errorf("match is a draw")
	case s.PlayerTurn != 1 && s.PlayerTurn != 2:
		return fmt.Errorf("invalid player turn")
	}
	i, p, found := 0, 0, false
	for ; i < len(s.Players); i++ {
		if s.Players[i].UserId == userId {
			found = true
			break
		}
	}
	switch p = i + 1; {
	case !found:
		return fmt.Errorf("unable to locate player with user id %q", userId)
	case s.PlayerTurn != p:
		return fmt.Errorf("it is not player %d's turn, it is player %d's turn", p, s.PlayerTurn)
	default:
		s.Cells[row][col] = p
		switch p {
		case 1:
			s.PlayerTurn = 2
		case 2:
			s.PlayerTurn = 1
		}
	}
	// determine if there is a winner
loop:
	for p, s.Winner = 1, 0; p <= 2; p++ {
		for i = 0; i < 8; i++ {
			if isWinner(p, s.Cells, coords[i]) {
				s.Winner = Winner(p)
				break loop
			}
		}
	}
	// check draw
	for i, s.Draw = 0, s.Winner == 0; s.Draw && i < 9; i++ {
		s.Draw = s.Draw && s.Cells[i/3][i%3] != -1
	}
	if s.Winner != 0 || s.Draw {
		s.PlayerTurn = -1
	}
	return nil
}

func (s *State) String() string {
	p1, p2 := "(nil)", "(nil)"
	if len(s.Players) > 0 {
		p1 = s.Players[0].UserId
	}
	if len(s.Players) > 1 {
		p2 = s.Players[1].UserId
	}
	v := make([]interface{}, 9)
	for i := 0; i < 9; i++ {
		v[i] = getCellAsRune(i/3, i%3, s.Cells)
	}
	return fmt.Sprintf(
		"1:%s 2:%s turn:%d winner:%d draw:%t cells:[%c%c%c %c%c%c %c%c%c]",
		append([]interface{}{
			p1,
			p2,
			s.PlayerTurn,
			s.Winner.Int(),
			s.Draw,
		}, v...)...)
}

func (state *State) Available() [][]int {
	var v [][]int
	for i := 0; i < 9; i++ {
		if state.Cells[i/3][i%3] == -1 {
			v = append(v, []int{i / 3, i % 3})
		}
	}
	return v
}

func getCellAsRune(i, j int, cells [][]int) rune {
	switch cells[i][j] {
	case 1:
		return 'O'
	case 2:
		return 'X'
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
