package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/ascii8/xoxo-go/xoxo"
)

func main() {
	urlstr := flag.String("url", "http://127.0.0.1:7350", "xoxo host")
	key := flag.String("key", "xoxo-go_server", "server key")
	seed := flag.Int64("seed", 0, "seed")
	count := flag.Int("count", 3, "game count")
	flag.Parse()
	if err := run(context.Background(), *urlstr, *key, *seed, *count); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, urlstr, key string, seed int64, count int) error {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	cl, err := xoxo.Dial(ctx, xoxo.WithURL(urlstr), xoxo.WithServerKey(key), xoxo.WithLogf(log.Printf), xoxo.WithDebug())
	if err != nil {
		return err
	}
	if err := cl.Join(ctx); err != nil {
		return err
	}
	for i := 0; count < 1 || i < count; i++ {
		for cl.Ready(ctx) && cl.Next(ctx) {
			state := cl.State()
			log.Printf("player turn %q (%d)", state.ActivePlayer.UserId, state.State.PlayerTurn)
			v := state.State.Available()
			n := r.Intn(len(v))
			log.Printf(
				"player %d available %d, choosing move %d (%d, %d)",
				state.State.PlayerTurn, len(v), n, v[n][0], v[n][1],
			)
			if err := cl.Move(ctx, v[n][0], v[n][1]); err != nil {
				return err
			}
		}
		switch state := cl.State(); {
		case state.State.Draw:
			log.Printf("game %d: was a draw!", i+1)
		default:
			log.Printf("game %d: player %d won!", i+1, state.State.Winner)
		}
	}
	<-time.After(2 * time.Second)
	return cl.Leave(ctx)
}
