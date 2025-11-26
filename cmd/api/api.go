package main

import "go.uber.org/zap"

type application struct {
	config config
	logger *zap.SugaredLogger
}

type config struct {
	db dbConfig
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}
