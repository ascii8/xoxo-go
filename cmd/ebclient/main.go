package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ascii8/xoxo-go/ebxoxo"
	"github.com/rs/zerolog"
)

func main() {
	debug := flag.Bool("debug", true, "enable debug")
	urlstr := flag.String("url", "http://127.0.0.1:7350", "xoxo host")
	key := flag.String("key", "xoxo-go_server", "server key")
	flag.Parse()
	if err := run(context.Background(), *debug, *urlstr, *key); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, debug bool, urlstr, key string) error {
	level := zerolog.Disabled
	if s := os.Getenv("LEVEL"); s != "" {
		if l, err := zerolog.ParseLevel(s); err == nil {
			level = l
		}
	}
	if s := os.Getenv("DEBUG"); s != "" && s != "0" && s != "off" && s != "false" {
		level = zerolog.DebugLevel
	}
	if s := os.Getenv("TRACE"); s != "" && s != "0" && s != "off" && s != "false" {
		level = zerolog.TraceLevel
	}
	if level > zerolog.DebugLevel && debug {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
	w := zerolog.NewConsoleWriter(func(cw *zerolog.ConsoleWriter) {
		cw.Out = os.Stdout
		cw.TimeFormat = "2006-01-02 15:04:05"
		cw.PartsOrder = []string{zerolog.TimestampFieldName, zerolog.LevelFieldName, zerolog.CallerFieldName, zerolog.MessageFieldName}
		cw.FieldsExclude = cw.PartsOrder
	})
	logger := zerolog.New(w).With().Timestamp().Logger()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		// catch signals, canceling context to cause cleanup
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case sig := <-ch:
			logger.Trace().Str("sig", sig.String()).Msg("caught signal")
			ebxoxo.Shutdown()
			cancel()
		}
	}()
	if err := ebxoxo.Run(ctx, logger, debug, urlstr, key); err != nil {
		return err
	}
	return nil
}
