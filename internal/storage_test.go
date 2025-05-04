package app

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	yttc "github.com/tractoai/testcontainers-ytsaurus"
)

// TODO: test sqlite storage

func TestNewYtStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	t.Run("NewYtStorage", func(t *testing.T) {
		container, err := yttc.RunContainer(ctx)
		require.NoError(t, err)

		ytc, err := container.NewClient(ctx)
		require.NoError(t, err)

		nodePath := "//tmp/test_node"

		storage, err := NewYtStorage(ytc, nodePath, logger)
		require.NoError(t, err)

		require.NotNil(t, storage)
		require.Equal(t, nodePath, storage.nodePath)
	})

	t.Run("Set and get", func(t *testing.T) {
		container, err := yttc.RunContainer(ctx)
		require.NoError(t, err)

		ytc, err := container.NewClient(ctx)
		require.NoError(t, err)

		nodePath := "//tmp/test_node"

		storage, err := NewYtStorage(ytc, nodePath, logger)
		require.NoError(t, err)

		err = storage.Set("key1", "value1")
		require.NoError(t, err)

		value, err := storage.Get("key1")
		require.NoError(t, err)
		require.Equal(t, "value1", *value)
	})
}
