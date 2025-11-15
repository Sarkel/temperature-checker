package db

import (
	"database/sql"
	"log/slog"
	"temperature-checker/internal/config"
	sqlc "temperature-checker/internal/db/gen"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
)

type Dependencies struct {
	Logger *slog.Logger
	Config *config.DatabaseConfig
}

type ConManager struct {
	db  *sql.DB
	log *slog.Logger
}

func NewConManager(deps Dependencies) (*ConManager, error) {
	cfg := deps.Config
	log := deps.Logger

	pgxCfg, err := pgx.ParseConfig(cfg.URL)

	if err != nil {
		log.Error("Failed to parse database URL", slog.String("error", err.Error()))
		return nil, err
	}

	if cfg.Debug {
		pgxCfg.Tracer = &tracelog.TraceLog{
			Logger:   tracelog.LoggerFunc(traceDBLogs(log)),
			LogLevel: tracelog.LogLevelDebug,
		}
	}

	connector := stdlib.GetConnector(*pgxCfg)

	db := sql.OpenDB(connector)

	db.SetMaxOpenConns(cfg.ConPool)

	return &ConManager{
		db:  db,
		log: deps.Logger,
	}, nil
}

func (c *ConManager) WithQ() *sqlc.Queries {
	return sqlc.New(c.db)
}

func (c *ConManager) Close() error {
	return c.db.Close()
}
