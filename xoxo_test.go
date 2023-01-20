package xoxo

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"reflect"
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

func TestMatch(t *testing.T) {
	tests := []struct {
		seed   int64
		draw   bool
		winner int
		cells  []int
	}{
		{102, false, 1, []int{
			-1, 2, 1,
			2, 1, -1,
			1, 1, 2,
		}},
		{200, true, 0, []int{
			2, 1, 2,
			1, 1, 2,
			1, 2, 1,
		}},
		{1048, false, 1, []int{
			1, -1, -1,
			-1, 1, -1,
			2, 2, 1,
		}},
		{6093, false, 2, []int{
			-1, 2, 1,
			1, 2, 1,
			-1, 2, -1,
		}},
		{9004, true, 1, []int{
			2, 2, 1,
			1, 2, 1,
			2, 1, 1,
		}},
	}
	for i, test := range tests {
		n := i
		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			matchTest(t, test.seed, test.draw, test.winner, test.cells)
		})
	}
}

func matchTest(t *testing.T, seed int64, draw bool, winner int, cells []int) {
	t.Logf("seed: %d draw: %t winner: %d", seed, draw, winner)
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
	eg.Go(runTestMatch(t, ctx, s1, s2, urlstr, res))
	eg.Go(runTestMatch(t, ctx, s1, s2, urlstr, nil))
	if err := eg.Wait(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if res.draw != draw {
		t.Errorf("expected draw: %t, got: %t", draw, res.draw)
	}
	if res.winner != winner {
		t.Errorf("expected winner: %d, got: %d", winner, res.winner)
	}
	if !reflect.DeepEqual(res.cells, cells) {
		t.Errorf("expected cells:\n%v\ngot:\n%v", cells, res.cells)
	}
	<-time.After(1500 * time.Millisecond)
}

func runTestMatch(t *testing.T, ctx context.Context, s1, s2 int64, urlstr string, res *matchResult) func() error {
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
			var available [][]int
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					if state.State.Cells[i][j] == -1 {
						available = append(available, []int{i, j})
					}
				}
			}
			n := r.Intn(len(available))
			t.Logf(
				"player %d available %d, choosing move %d (%d, %d)",
				state.State.PlayerTurn, len(available), n, available[n][0], available[n][1],
			)
			if err := cl.Move(ctx, available[n][0], available[n][1]); err != nil {
				return err
			}
		}
		if state := cl.State(); state != nil && res != nil {
			res.draw = state.State.Draw
			res.winner = state.State.Winner
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
