package main

import (
	"context"

	"github.com/KanathipP/KubeLogPullStoreGopher/internal/db"
	"github.com/KanathipP/KubeLogPullStoreGopher/internal/env"
	"github.com/KanathipP/KubeLogPullStoreGopher/internal/kubeclient"
	"go.uber.org/zap"
)

func main() {
	cfg := config{
		env: env.GetStr("ENV", "development"),
		db: dbConfig{
			addr:         env.GetStr("DB_ADDR", "postgres://admin:adminpassword@localhost:5432/fl_store?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetStr("DB_MAX_IDLE_TIME", "15m"),
		},
		kubeconfig: env.GetStr("KUBECONFIG", ""),
		podFilter: podFilterConfig{
			namespace:     env.GetStr("POD_FILTER_NAMESPACE", "flwr"),
			labelSelector: env.GetStr("POD_FILTER_LABEL_SELECTOR", "name=superexec"),
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Info("Database connection pool established")

	kube := kubeclient.New(cfg.kubeconfig)
	if kube == nil {
		logger.Fatal("Failed to initialized kubernetes client")
	}
	logger.Info("Kubernetes client initialized")

	app := &application{
		config: cfg,
		logger: logger,
		kube:   kube,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
}
