package main

import (
	"github.com/KanathipP/KubeLogPullStoreGopher/internal/kubeclient"
	"go.uber.org/zap"
)

type application struct {
	config config
	logger *zap.SugaredLogger
	kube   *kubeclient.Set
}

type config struct {
	env        string
	db         dbConfig
	kubeconfig string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}
