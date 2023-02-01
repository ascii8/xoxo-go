package fynexoxo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ascii8/xoxo-go/xoxo"
	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

var game *Game

func Run(ctx context.Context, logger zerolog.Logger, debug bool, urlstr, key string) error {
	game = New(ctx, logger, debug, urlstr, key)
	return game.Run()
}

func Shutdown() {
	game.Shutdown()
}

type Game struct {
	ctx            context.Context
	logger         zerolog.Logger
	debug          bool
	url            string
	key            string
	userId         string
	username       string
	cl             *xoxo.Client
	app            fyne.App
	window         fyne.Window
	connectedLabel *widget.Label
	turnLabel      *widget.Label
	cellButtons    []*widget.Button
}

func New(ctx context.Context, logger zerolog.Logger, debug bool, urlstr, key string) *Game {
	g := &Game{
		ctx:      ctx,
		logger:   logger,
		debug:    debug,
		url:      urlstr,
		key:      key,
		userId:   uuid.New().String(),
		username: xid.New().String(),
	}
	g.init()
	g.cl = xoxo.NewClient(
		xoxo.WithURL(g.url),
		xoxo.WithServerKey(g.key),
		xoxo.WithUserId(g.userId),
		xoxo.WithUsername(g.username),
		xoxo.WithLogf(func(s string, v ...interface{}) {
			g.logger.Debug().CallerSkipFrame(1).Msgf(s, v...)
		}),
		xoxo.WithDebug(),
		xoxo.WithPersist(),
		xoxo.WithHandler(g),
	)
	return g
}

func (g *Game) init() {
	g.app = app.New()
	g.window = g.app.NewWindow("XOXO")
	g.connectedLabel = widget.NewLabel("...")
	g.turnLabel = widget.NewLabel("")
	g.cellButtons = make([]*widget.Button, 9)
	grid := container.New(layout.NewGridLayoutWithColumns(3))
	for i := 0; i < 9; i++ {
		g.cellButtons[i] = widget.NewButton(" ", g.move(i/3, i%3))
		grid.Add(g.cellButtons[i])
	}
	top := container.NewHBox(widget.NewLabel("XOXO"), g.turnLabel)
	content := container.NewBorder(
		top,
		widget.NewButton("Join", g.join),
		nil,
		nil,
		grid,
	)
	g.window.SetContent(container.NewBorder(
		nil,
		g.connectedLabel,
		nil,
		nil,
		content,
	))
	g.window.Resize(fyne.Size{Width: 640, Height: 1136})
	g.window.SetFixedSize(true)
}

func (g *Game) join() {
	g.logger.
		Debug().
		Msg("join")
	if g.cl.Connected() {
		if err := g.cl.Join(g.ctx); err != nil {
			g.logger.
				Debug().
				Err(err).
				Msg("unable to join")
		}
	}
}

func (g *Game) move(row, col int) func() {
	return func() {
		g.logger.
			Debug().
			Int("row", row).
			Int("col", col).
			Msg("move")
		if err := g.cl.Move(g.ctx, row, col); err != nil {
			g.logger.
				Debug().
				Err(err).
				Msg("unable to move")
		}
	}
}

func (g *Game) Run() error {
	if err := g.cl.Open(g.ctx); err != nil {
		return err
	}
	g.window.ShowAndRun()
	return nil
}

func (g *Game) Shutdown() {
	g.logger.
		Debug().
		Msg("Shutdown")
	g.window.Close()
	g.app.Quit()
}

func (g *Game) ConnectHandler(ctx context.Context) {
	g.connectedLabel.SetText("Connected.")
	g.turnLabel.SetText("")
}

func (g *Game) DisconnectHandler(ctx context.Context, err error) {
	g.connectedLabel.SetText("Disconnected!")
	g.turnLabel.SetText("")
	go func() {
		<-time.After(1 * time.Second)
		for i := 0; !g.cl.Connected(); i++ {
			g.connectedLabel.SetText(strings.Repeat(".", 1+i%5))
			<-time.After(1 * time.Second)
		}
	}()
}

func (g *Game) StateHandler(ctx context.Context) {
	g.logger.
		Debug().
		Msg("state change")
	state := g.cl.State()
	s := ""
	switch {
	case state == nil:
	case state.State.Winner != 0:
		winner := 'O'
		if state.State.Winner == 2 {
			winner = 'X'
		}
		s = fmt.Sprintf("Player %d (%c) wins! %d...", state.State.Winner, winner, state.State.RematchCountdown)
	case state.State.Draw:
		s = fmt.Sprintf("Draw! %d...", state.State.RematchCountdown)
	case !state.YourTurn:
		s = "Waiting Other Player"
	case state.YourTurn:
		s = "Your Turn!"
	}
	g.turnLabel.SetText(s)
	for i := 0; i < 9; i++ {
		s := ""
		switch {
		case state == nil:
		case state.State.Cells[i/3][i%3] == 1:
			s = "O"
		case state.State.Cells[i/3][i%3] == 2:
			s = "X"
		}
		g.cellButtons[i].SetText(s)
	}
}
