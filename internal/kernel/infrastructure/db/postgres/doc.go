// Package postgres implements postgres connection.
//
// Example usage with options:
//
//	import (
//		"context"
//		"time"
//
//		"gct/config"
//		"gct/internal/kernel/infrastructure/logger"
//		"gct/internal/kernel/infrastructure/db/postgres"
//	)
//
//	func main() {
//		ctx := context.Background()
//		l := logger.New("info")
//		cfg := config.Postgres{
//			Host:     "localhost",
//			Port:     5432,
//			User:     "user",
//			Password: "password",
//			Name:     "database",
//			SSLMode:  "disable",
//		}
//
//		// Basic usage without options
//		pg, err := postgres.New(ctx, "dev", cfg, l)
//		if err != nil {
//			panic(err)
//		}
//		defer pg.Close()
//
//		// Advanced usage with custom options
//		pg, err = postgres.New(ctx, "dev", cfg, l,
//			postgres.WithMaxConns(100),
//			postgres.WithMinConns(10),
//			postgres.WithMaxConnLifetime(1*time.Hour),
//			postgres.WithMaxConnIdleTime(10*time.Minute),
//			postgres.WithStatementTimeout(30*time.Second),
//			postgres.WithApplicationName("my-service"),
//		)
//		if err != nil {
//			panic(err)
//		}
//		defer pg.Close()
//	}
package postgres
