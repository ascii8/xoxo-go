package main

import (
	"context"
	"database/sql"

	"github.com/ascii8/xoxo-go/nkxoxo"
	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	return nkxoxo.InitModule(ctx, logger, db, nk, initializer)
}
