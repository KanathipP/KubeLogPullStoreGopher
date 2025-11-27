CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS fl_trainings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fl_training_id TEXT NOT NULL UNIQUE,
  current_server_round INT NOT NULL DEFAULT 1,
  total_server_round INT NOT NULL DEFAULT -1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS training_clients (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fl_training_id TEXT NOT NULL REFERENCES fl_trainings(fl_training_id) ON DELETE CASCADE,
  partition_id INT NOT NULL,
  node_name VARCHAR(255) NOT NULL,
  pod_name VARCHAR(255) NOT NULL,
  state VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_log_read TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (fl_training_id, partition_id)
);

CREATE TABLE IF NOT EXISTS client_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id UUID NOT NULL REFERENCES training_clients(id) ON DELETE CASCADE,
  text TEXT NOT NULL,
  client_output_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS training_graphs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id UUID NOT NULL REFERENCES training_clients(id) ON DELETE CASCADE,
  server_round INT NOT NULL,
  optimizer TEXT,
  learning_rate DOUBLE PRECISION,
  num_epochs INT NOT NULL,
  batch_size INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (client_id, server_round)
);

CREATE TABLE IF NOT EXISTS training_graph_points (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  graph_id UUID NOT NULL REFERENCES training_graphs(id) ON DELETE CASCADE,
  current_epoch INT NOT NULL,
  trained_batch INT NOT NULL,
  train_loss DOUBLE PRECISION,
  val_loss DOUBLE PRECISION,
  accuracy DOUBLE PRECISION,
  epoch_elapsed_time DOUBLE PRECISION,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (graph_id, current_epoch)
);

CREATE TABLE IF NOT EXISTS testing_graphs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id UUID NOT NULL REFERENCES training_clients(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (client_id)
);

CREATE TABLE IF NOT EXISTS testing_graph_points (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  graph_id UUID NOT NULL REFERENCES testing_graphs(id) ON DELETE CASCADE,
  server_round INT NOT NULL,
  criterion TEXT,
  batch_size INT,
  test_loss DOUBLE PRECISION,
  accuracy DOUBLE PRECISION,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (graph_id, server_round)
);

CREATE TABLE IF NOT EXISTS fl_model_weights (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fl_training_id TEXT NOT NULL REFERENCES fl_trainings(fl_training_id) ON DELETE CASCADE,
  server_round INT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (fl_training_id, server_round)
);

CREATE TABLE IF NOT EXISTS training_servers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fl_training_id TEXT NOT NULL REFERENCES fl_trainings(fl_training_id) ON DELETE CASCADE,
  node_name VARCHAR(255) NOT NULL,
  pod_name VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_log_read TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (fl_training_id)
);

