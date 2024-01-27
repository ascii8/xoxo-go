package gioxoxo

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/ascii8/xoxo-go/xoxo"
	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

const windowWidth, windowHeight = 640, 900

var game *Game

func Run(ctx context.Context, logger zerolog.Logger, debug bool, urlstr, key string) error {
	game = New(ctx, logger, debug, urlstr, key)
	return game.Run()
}

func Shutdown() {
	game.Shutdown()
}

type Game struct {
	ctx              context.Context
	logger           zerolog.Logger
	debug            bool
	url              string
	key              string
	userId           string
	username         string
	cl               *xoxo.Client
	window           *app.Window
	connectedLabel   string
	turnLabel        string
	cellButtonLabels []string
	join             *widget.Clickable
	cellButtons      []*widget.Clickable
}

func New(ctx context.Context, logger zerolog.Logger, debug bool, urlstr, key string) *Game {
	g := &Game{
		ctx:              ctx,
		logger:           logger,
		debug:            debug,
		url:              urlstr,
		key:              key,
		userId:           uuid.New().String(),
		username:         xid.New().String(),
		connectedLabel:   ".",
		cellButtonLabels: make([]string, 9),
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
	const width, height = windowWidth, windowHeight
	g.window = app.NewWindow(
		app.Title("XOXO"),
		app.Size(width, height),
		app.MinSize(width, height),
		app.MaxSize(width, height),
		app.Decorated(false),
	)
	g.join = new(widget.Clickable)
	g.cellButtons = make([]*widget.Clickable, 9)
	for i := 0; i < 9; i++ {
		g.cellButtons[i] = new(widget.Clickable)
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
	go g.run()
	app.Main()
	return nil
}

func (g *Game) run() {
	f := g.layout()
	for {
		event := g.window.NextEvent()
		switch ev := event.(type) {
		case system.DestroyEvent:
			g.logger.
				Debug().
				Err(ev.Err).
				Msg("DestroyEvent")
			os.Exit(0)
		case system.FrameEvent:
			f(ev)
		default:
			if ev == nil {
				continue
			}
			g.logger.
				Debug().
				Str("type", fmt.Sprintf("%T", ev)).
				Msg("window event")
		}
	}
}

func (g *Game) layout() func(system.FrameEvent) {
	th := material.NewTheme()
	gofont.Collection()
	var ops op.Ops
	var grid component.GridState
	return func(ev system.FrameEvent) {
		gtx := layout.NewContext(&ops, ev)
		// handle join
		if g.join.Clicked(gtx) {
			g.cl.JoinAsync(g.ctx, func(err error) {
				if err != nil {
					g.logger.
						Debug().
						Err(err).
						Msg("unable to join")
				}
			})
		}
		// handle cell buttons
		for i := 0; i < 9; i++ {
			if g.cellButtons[i].Clicked(gtx) {
				cell := i
				g.cl.MoveAsync(g.ctx, cell/3, cell%3, func(err error) {
					if err != nil {
						g.logger.
							Debug().
							Err(err).
							Int("row", cell/3).
							Int("col", cell%3).
							Msg("unable to move")
					}
				})
			}
		}
		layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceEvenly,
		}.Layout(
			gtx,
			// label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    25,
					Bottom: 25,
					Left:   25,
					Right:  25,
				}.Layout(gtx, material.Label(th, 32, "XOXO "+g.turnLabel).Layout)
			}),
			// grid
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return component.Grid(th, &grid).Layout(
					gtx,
					3, 3,
					func(_ layout.Axis, _, _ int) int {
						return (windowWidth - 10) / 3
					},
					func(gtx layout.Context, row, col int) layout.Dimensions {
						return layout.Inset{
							Top:  7,
							Left: 7,
						}.Layout(
							gtx,
							material.Button(
								th,
								g.cellButtons[row*3+col],
								g.cellButtonLabels[row*3+col],
							).Layout,
						)
					},
				)
			}),
			// join button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    25,
					Bottom: 25,
					Right:  25,
					Left:   25,
				}.Layout(
					gtx,
					material.Button(th, g.join, "Join").Layout,
				)
			}),
			// connected label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    10,
					Bottom: 10,
					Right:  10,
					Left:   10,
				}.Layout(
					gtx,
					material.Label(th, 18, g.connectedLabel).Layout,
				)
			}),
		)
		ev.Frame(gtx.Ops)
	}
}

func (g *Game) Shutdown() {
	g.logger.
		Debug().
		Msg("Shutdown")
	g.window.Perform(system.ActionClose)
}

func (g *Game) ConnectHandler(ctx context.Context) {
	g.connectedLabel = "Connected."
	g.turnLabel = ""
	g.window.Invalidate()
}

func (g *Game) DisconnectHandler(ctx context.Context, err error) {
	g.connectedLabel = "Disconnected!"
	g.turnLabel = ""
	g.window.Invalidate()
	go func() {
		<-time.After(1 * time.Second)
		for i := 0; !g.cl.Connected(); i++ {
			g.connectedLabel = strings.Repeat(".", 1+i%5)
			<-time.After(1 * time.Second)
			g.window.Invalidate()
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
	g.turnLabel = s
	for i := 0; i < 9; i++ {
		s := ""
		switch {
		case state == nil:
		case state.State.Cells[i/3][i%3] == 1:
			s = "O"
		case state.State.Cells[i/3][i%3] == 2:
			s = "X"
		}
		g.cellButtonLabels[i] = s
	}
	g.window.Invalidate()
}
