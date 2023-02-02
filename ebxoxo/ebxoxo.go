package ebxoxo

import (
	"context"
	"errors"
	"fmt"
	"image/color"

	"github.com/ascii8/xoxo-go/ebxoxo/assets"
	"github.com/ascii8/xoxo-go/xoxo"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

const (
	windowWidth  = 640
	windowHeight = 1136
)

func Run(ctx context.Context, logger zerolog.Logger, debug bool, urlstr, key string) error {
	ebiten.SetWindowTitle("XOXO")
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetWindowClosingHandled(true)
	scaling := ebiten.DeviceScaleFactor()
	if scaling == 0.0 {
		scaling = 1.0
	}
	width, height := ebiten.ScreenSizeInFullscreen()
	switch {
	case scaling == 1.0 && ((width > 3000 && height > 2000) || (width > 2000 && height > 3000)):
		scaling = 1.15
	}
	logger.Debug().
		Float64("scale", scaling).
		Int("width", width).
		Int("height", height).
		Msg("window")
	ebiten.SetWindowSize(int(windowWidth*scaling), int(windowHeight*scaling))
	game = New(ctx, logger, debug, scaling, urlstr, key)
	if err := ebiten.RunGame(game); err != nil && !errors.Is(err, ebiten.Termination) {
		return err
	}
	return nil
}

func Shutdown() {
	game.Shutdown()
}

var game *Game

type Game struct {
	ctx      context.Context
	logger   zerolog.Logger
	debug    bool
	scaling  float64
	url      string
	key      string
	userId   string
	username string
	exiting  bool
	err      error
	cl       *xoxo.Client
	join     *Button
	leave    *Button
	board    *Board
	tick     int
}

func New(ctx context.Context, logger zerolog.Logger, debug bool, scaling float64, urlstr, key string) *Game {
	return &Game{
		ctx:      ctx,
		logger:   logger,
		debug:    debug,
		scaling:  scaling,
		url:      urlstr,
		key:      key,
		userId:   uuid.New().String(),
		username: xid.New().String(),
		tick:     -1,
	}
}

func (g *Game) init() error {
	if err := assets.Init(windowWidth, windowHeight); err != nil {
		return err
	}
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
	g.join = NewButton(
		"Join",
		100, 800,
		414, 108,
		color.White, color.RGBA{255, 0, 127, 255},
		assets.Btn, assets.BtnActive,
	)
	g.leave = NewButton(
		"Leave",
		100, 800,
		414, 108,
		color.White, color.RGBA{255, 0, 127, 255},
		assets.Btn, assets.BtnActive,
	)
	go g.cl.Open(g.ctx)
	return nil
}

func (g *Game) Update() error {
	g.tick++
	switch {
	case g.exiting:
		return ebiten.Termination
	case ebiten.IsWindowBeingClosed():
		g.Shutdown()
		return nil
	case g.cl != nil:
		return nil
	}
	if err := g.init(); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// background
	screen.DrawImage(assets.Bg, assets.BgOpts)
	connected := g.Connected()
	x, y := ebiten.CursorPosition()
	switch state := g.cl.State(); {
	case connected && state == nil:
		// draw title
		// logo
		// empty board + TIC TAC TOE
		g.join.Draw(screen, x, y, g.tick)
	case connected:
		// draw board/match
	}
	if connected {
		text.Draw(screen, "Connected.", assets.Din24, 16, windowHeight-72, color.White)
	}
	if g.debug {
		text.Draw(
			screen,
			fmt.Sprintf("FPS: %0.0f Ticks: %0.0f\n(%d,%d)", ebiten.ActualFPS(), ebiten.ActualTPS(), x, y),
			assets.Din16, 16, windowHeight-40, color.White,
		)
	}
}

func (g *Game) LayoutF(float64, float64) (float64, float64) {
	return windowWidth, windowHeight
}

func (g *Game) ShutdownWithError(err error) {
	g.exiting, g.err = true, err
}

func (g *Game) Connected() bool {
	cl := g.cl
	return cl != nil && cl.Connected()
}

func (g *Game) Err() error {
	return g.err
}

func (g *Game) Layout(int, int) (int, int) {
	panic("should never be called")
}

func (g *Game) Shutdown() {
	g.logger.Debug().Msg("Shutdown")
	cl := g.cl
	g.exiting = true
	if cl != nil {
		cl.Leave(context.Background())
	}
}

func (g *Game) ConnectHandler(ctx context.Context) {
}
