package api

import (
	"testing"

	"github.com/PSKP-95/scheduler/config"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := config.ServerConfig{
		ServerAddress: "0.0.0.0:8888",
	}

	server, err := NewServer(config, store, nil, nil)
	require.NoError(t, err)

	return server
}
