package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	ggh "github.com/dmi-feo/gogohandlers"
	yttc "github.com/tractoai/testcontainers-ytsaurus"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHandlePing(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbFilePath := fmt.Sprintf("/tmp/test_%s", uuid.New().String())
	defer func() { _ = os.Remove(dbFilePath) }()

	appConfig := AppConfig{StorageType: "sqlite", StorageConfig: map[string]any{"file_path": dbFilePath}}
	sp, err := NewExampleAppServiceProvider(appConfig, logger)
	require.NoError(t, err)

	handler := ggh.GGHandler[ExampleAppServiceProvider, struct{}, PingGetParams, PingResponse, ExampleAppErrorData]{
		ServiceProvider: sp.(*ExampleAppServiceProvider),
		HandlerFunc:     HandlePing,
		Middlewares:     getDefaultMiddlewares[struct{}, PingGetParams, PingResponse](),
		Logger:          logger,
	}

	t.Run("works without get params", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/ping?mayfail=0", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusOK, response.Code)

		var responseBody PingResponse
		err = json.Unmarshal(response.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		require.Equal(t, "pong", responseBody.Message)
	})

	t.Run("returns custom message", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/ping?mayfail=0&msg=test-message", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusOK, response.Code)

		var responseBody PingResponse
		err = json.Unmarshal(response.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		require.Equal(t, "test-message", responseBody.Message)
	})

	t.Run("returns custom error data", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/ping?mustfail=1", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusTeapot, response.Code)

		var errorData ExampleAppErrorData
		err = json.Unmarshal(response.Body.Bytes(), &errorData)
		require.NoError(t, err)
		require.Equal(t, "TEAPOT", errorData.Code)
	})
}

func TestHandleGetSetValue(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Set up SQLite storage configurations
	dbFilePath := fmt.Sprintf("/tmp/test_%s", uuid.New().String())
	defer func() { _ = os.Remove(dbFilePath) }()
	sqliteAppConfig := AppConfig{
		StorageType: "sqlite",
		StorageConfig: map[string]any{
			"file_path": dbFilePath,
		},
	}

	// Set up YT storage configurations
	ctx := context.Background()
	container, err := yttc.RunContainer(ctx)
	require.NoError(t, err)

	proxy, err := container.GetProxy(ctx)
	require.NoError(t, err)

	ytAppConfig := AppConfig{
		StorageType: "yt",
		StorageConfig: map[string]any{
			"cluster": proxy,
			"token":   "fake",
			"path":    "//tmp/test_node",
		},
	}

	// Test cases for different storage types
	testCases := []struct {
		name      string
		appConfig AppConfig
	}{
		{name: "sqlite", appConfig: sqliteAppConfig},
		{name: "yt", appConfig: ytAppConfig},
	}

	for _, tc := range testCases {
		sp, err := NewExampleAppServiceProvider(tc.appConfig, logger)
		require.NoError(t, err)

		setValueHandler := ggh.GGHandler[ExampleAppServiceProvider, SetValueRequest, struct{}, SetValueResponse, ExampleAppErrorData]{
			ServiceProvider: sp.(*ExampleAppServiceProvider),
			HandlerFunc:     HandleSetValue,
			Middlewares:     getDefaultMiddlewares[SetValueRequest, struct{}, SetValueResponse](),
			Logger:          logger,
		}

		getValueHandler := ggh.GGHandler[ExampleAppServiceProvider, struct{}, struct{}, GetValueResponse, ExampleAppErrorData]{
			ServiceProvider: sp.(*ExampleAppServiceProvider),
			HandlerFunc:     HandleGetValue,
			Middlewares:     getDefaultMiddlewares[struct{}, struct{}, GetValueResponse](),
			Logger:          logger,
		}

		t.Run(fmt.Sprintf("set and get value %s", tc.name), func(t *testing.T) {
			reqData := SetValueRequest{Key: "test-key", Value: "test-value"}
			reqBody, err := json.Marshal(reqData)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, "/set_value", bytes.NewReader(reqBody))
			require.NoError(t, err)
			response := httptest.NewRecorder()

			setValueHandler.ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code)

			request, err = http.NewRequest(http.MethodGet, "/get_value/{key}", nil)
			require.NoError(t, err)
			request.SetPathValue("key", "test-key")
			response = httptest.NewRecorder()

			getValueHandler.ServeHTTP(response, request)
			require.Equal(t, http.StatusOK, response.Code)
			var responseBody GetValueResponse
			err = json.Unmarshal(response.Body.Bytes(), &responseBody)
			require.NoError(t, err)
			require.Equal(t, "test-value", responseBody.Value)
		})
	}
}
