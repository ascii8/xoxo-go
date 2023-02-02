package xoxo

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ascii8/nktest"
	"github.com/ascii8/xoxo-go/xoxo"
	"golang.org/x/sync/errgroup"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	ctx = nktest.WithAlwaysPullFromEnv(ctx, "PULL")
	ctx = nktest.WithUnderCIFromEnv(ctx, "CI")
	ctx = nktest.WithHostPortMap(ctx)
	var opts []nktest.BuildConfigOption
	if os.Getenv("CI") == "" {
		opts = append(opts, nktest.WithDefaultGoEnv(), nktest.WithDefaultGoVolumes())
	}
	nktest.Main(ctx, m,
		nktest.WithDir("."),
		nktest.WithBuildConfig("./cmd/nkxoxo", opts...),
	)
}

func TestKeep(t *testing.T) {
	keep := os.Getenv("KEEP")
	if keep == "" {
		return
	}
	d, err := time.ParseDuration(keep)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	ctx, cancel, nk := nktest.WithCancel(context.Background(), t)
	defer cancel()
	urlstr, err := nk.RunProxy(ctx, nktest.WithAddr("127.0.0.1:7352"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	t.Logf("local: %s", nk.HttpLocal())
	t.Logf("grpc: %s", nk.GrpcLocal())
	t.Logf("http: %s", nk.HttpLocal())
	t.Logf("console: %s", nk.ConsoleLocal())
	t.Logf("http_key: %s", nk.HttpKey())
	t.Logf("server_key: %s", nk.ServerKey())
	t.Logf("proxy: %s", urlstr)
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}

func TestMove(t *testing.T) {
	for i, v := range cellTests() {
		test := v
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			moveTest(t, test.seed, test.winner, test.draw, test.cells)
		})
	}
}

func TestMatch(t *testing.T) {
	for i, v := range cellTests() {
		test := v
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			matchTest(t, test.seed, test.winner, test.draw, test.cells)
		})
	}
}

func moveTest(t *testing.T, seed int64, winner int, draw bool, exp []int) {
	t.Logf("seed: %d winner: %d draw: %t", seed, winner, draw)
	r := rand.New(rand.NewSource(seed))
	r1, r2 := rand.New(rand.NewSource(r.Int63())), rand.New(rand.NewSource(r.Int63()))
	state := xoxo.NewState()
	for i := 0; i < 2; i++ {
		if err := state.Add("", "", strconv.Itoa(i), ""); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	}
	for i := 0; state.Winner == 0 && !state.Draw; i = (i + 1) % 2 {
		v := state.Available()
		if len(v) == 0 {
			break
		}
		var rr *rand.Rand
		switch {
		case i%2 == 0:
			rr = r1
		default:
			rr = r2
		}
		move := v[rr.Intn(len(v))]
		if err := state.Move(strconv.Itoa(i), xoxo.NewMove(move[0], move[1])); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	}
	if state.Winner.Int() != winner {
		t.Errorf("expected winner: %d, got: %d", winner, state.Winner)
	}
	if state.Draw != draw {
		t.Errorf("expected draw: %t, got: %t", draw, state.Draw)
	}
	cells := make([]int, 9)
	copy(cells[0:3], state.Cells[0][:])
	copy(cells[3:6], state.Cells[1][:])
	copy(cells[6:9], state.Cells[2][:])
	if !reflect.DeepEqual(cells, exp) {
		t.Errorf("expected cells:\n%v\ngot:\n%v", exp, cells)
	}
	t.Logf("state: %s", state)
}

func matchTest(t *testing.T, seed int64, winner int, draw bool, cells []int) {
	t.Logf("seed: %d winner: %d draw: %t", seed, winner, draw)
	r := rand.New(rand.NewSource(seed))
	s1, s2 := r.Int63(), r.Int63()
	ctx, cancel, nk := nktest.WithCancel(context.Background(), t)
	defer cancel()
	urlstr, err := nk.RunProxy(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	t.Logf("proxy: %s", urlstr)
	res := new(matchResult)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(runMatch(t, ctx, s1, s2, urlstr, res))
	eg.Go(runMatch(t, ctx, s1, s2, urlstr, nil))
	if err := eg.Wait(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if res.winner != winner {
		t.Errorf("expected winner: %d, got: %d", winner, res.winner)
	}
	if res.draw != draw {
		t.Errorf("expected draw: %t, got: %t", draw, res.draw)
	}
	if !reflect.DeepEqual(res.cells, cells) {
		t.Errorf("expected cells:\n%v\ngot:\n%v", cells, res.cells)
	}
	<-time.After(1500 * time.Millisecond)
}

func runMatch(t *testing.T, ctx context.Context, s1, s2 int64, urlstr string, res *matchResult) func() error {
	r1, r2 := rand.New(rand.NewSource(s1)), rand.New(rand.NewSource(s2))
	return func() error {
		cl, err := xoxo.Dial(ctx, xoxo.WithURL(urlstr), xoxo.WithLogf(t.Logf), xoxo.WithDebug())
		if err != nil {
			return err
		}
		if err := cl.Join(ctx); err != nil {
			return err
		}
		for cl.Next(ctx) {
			state := cl.State()
			t.Logf("player turn %q (%d)", state.ActivePlayer.UserId, state.State.PlayerTurn)
			r := r1
			if state.State.PlayerTurn == 2 {
				r = r2
			}
			v := state.State.Available()
			if len(v) == 0 {
				break
			}
			n := r.Intn(len(v))
			t.Logf(
				"player %d available %d, choosing move %d (%d, %d)",
				state.State.PlayerTurn, len(v), n, v[n][0], v[n][1],
			)
			if err := cl.Move(ctx, v[n][0], v[n][1]); err != nil {
				return err
			}
		}
		if state := cl.State(); state != nil && res != nil {
			res.draw = state.State.Draw
			res.winner = state.State.Winner.Int()
			res.cells = make([]int, 9)
			copy(res.cells[0:3], state.State.Cells[0][:])
			copy(res.cells[3:6], state.State.Cells[1][:])
			copy(res.cells[6:9], state.State.Cells[2][:])
		}
		if err := cl.Leave(ctx); err != nil {
			return err
		}
		return nil
	}
}

type matchResult struct {
	draw   bool
	winner int
	cells  []int
}

type cellTest struct {
	seed   int64
	winner int
	draw   bool
	cells  []int
}

func cellTests() []cellTest {
	return []cellTest{
		{102, 1, false, []int{ // p1 bottom to top
			-1, 2, 1,
			2, 1, -1,
			1, 1, 2,
		}},
		{200, 0, true, []int{ // draw
			2, 1, 2,
			1, 1, 2,
			1, 2, 1,
		}},
		{1048, 1, false, []int{ // p1 top to bottom
			1, -1, -1,
			-1, 1, -1,
			2, 2, 1,
		}},
		{6093, 2, false, []int{ // p2 col 1
			-1, 2, 1,
			1, 2, 1,
			-1, 2, -1,
		}},
		{9004, 1, false, []int{ // p1 col 2
			2, 2, 1,
			1, 2, 1,
			2, 1, 1,
		}},
	}
}
