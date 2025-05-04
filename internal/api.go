package app

import (
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"

	ggh "github.com/dmi-feo/gogohandlers"

	_ "github.com/mattn/go-sqlite3"
)

type PingGetParams struct {
	Message  string `schema:"msg,default:pong"`
	MayFail  bool   `schema:"mayfail"`
	MustFail bool   `schema:"mustfail"`
}

type PingResponse struct {
	Message string `json:"msg"`
}

func HandlePing(ggreq *ggh.GGRequest[ExampleAppServiceProvider, struct{}, PingGetParams]) (*ggh.GGResponse[PingResponse, ExampleAppErrorData], error) {
	ggreq.Logger.Info("Preparing pong...")
	if ggreq.GetParams.MayFail && rand.Intn(2) == 1 || ggreq.GetParams.MustFail {
		return &ggh.GGResponse[PingResponse, ExampleAppErrorData]{}, RandomError{}
	}
	return &ggh.GGResponse[PingResponse, ExampleAppErrorData]{
		ResponseData: &PingResponse{
			Message: ggreq.GetParams.Message,
		},
	}, nil
}

type SetValueRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SetValueResponse struct {
	Message string `json:"message"`
}

func HandleSetValue(ggreq *ggh.GGRequest[ExampleAppServiceProvider, SetValueRequest, struct{}]) (*ggh.GGResponse[SetValueResponse, ExampleAppErrorData], error) {
	storage := ggreq.ServiceProvider.GetStorage()
	err := storage.Set(ggreq.RequestData.Key, ggreq.RequestData.Value)
	if err != nil {
		return &ggh.GGResponse[SetValueResponse, ExampleAppErrorData]{}, DatabaseError{DBMessage: err.Error()}
	}
	return &ggh.GGResponse[SetValueResponse, ExampleAppErrorData]{
		ResponseData: &SetValueResponse{Message: "ok"},
	}, nil
}

type GetValueResponse struct {
	Value string `json:"value"`
}

func HandleGetValue(ggreq *ggh.GGRequest[ExampleAppServiceProvider, struct{}, struct{}]) (*ggh.GGResponse[GetValueResponse, ExampleAppErrorData], error) {
	key := ggreq.Request.PathValue("key")
	storage := ggreq.ServiceProvider.GetStorage()
	value, err := storage.Get(key)
	if err != nil {
		return &ggh.GGResponse[GetValueResponse, ExampleAppErrorData]{}, DatabaseError{DBMessage: err.Error()}
	}
	var returnValue string
	if value != nil {
		returnValue = *value
	}
	return &ggh.GGResponse[GetValueResponse, ExampleAppErrorData]{
		ResponseData: &GetValueResponse{Value: returnValue},
	}, nil
}

func getDefaultMiddlewares[TReqBody, TGetParams, TRespBody any]() []func(hFunc func(*ggh.GGRequest[ExampleAppServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, ExampleAppErrorData], error)) func(*ggh.GGRequest[ExampleAppServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, ExampleAppErrorData], error) {
	return []func(func(*ggh.GGRequest[ExampleAppServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, ExampleAppErrorData], error)) func(*ggh.GGRequest[ExampleAppServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, ExampleAppErrorData], error){
		ggh.GetErrorHandlingMiddleware[ExampleAppServiceProvider, TReqBody, TGetParams, TRespBody, ExampleAppErrorData](HandleErrors),
		ggh.GetDataProcessingMiddleware[ExampleAppServiceProvider, TReqBody, TGetParams, TRespBody, ExampleAppErrorData](nil),
		ggh.RequestLoggingMiddleware[ExampleAppServiceProvider, TReqBody, TGetParams, TRespBody, ExampleAppErrorData],
		ggh.RequestIDMiddleware[ExampleAppServiceProvider, TReqBody, TGetParams, TRespBody, ExampleAppErrorData],
	}
}

func GetRouter() *http.ServeMux {
	loggingHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggingHandler).WithGroup("fields")

	mux := http.NewServeMux()

	appConfig := AppConfig{
		StorageType: "sqlite",
		StorageConfig: map[string]any{
			"file_path": "/tmp/foo",
		},
	}

	sp, err := NewExampleAppServiceProvider(appConfig, logger)
	if err != nil {
		log.Fatal(err)
	}

	mux.Handle("GET /ping", &ggh.GGHandler[ExampleAppServiceProvider, struct{}, PingGetParams, PingResponse, ExampleAppErrorData]{
		ServiceProvider: sp.(*ExampleAppServiceProvider),
		HandlerFunc:     HandlePing,
		Middlewares:     getDefaultMiddlewares[struct{}, PingGetParams, PingResponse](),
		Logger:          logger,
	})

	mux.Handle("POST /set_value", &ggh.GGHandler[ExampleAppServiceProvider, SetValueRequest, struct{}, SetValueResponse, ExampleAppErrorData]{
		ServiceProvider: sp.(*ExampleAppServiceProvider),
		HandlerFunc:     HandleSetValue,
		Middlewares:     getDefaultMiddlewares[SetValueRequest, struct{}, SetValueResponse](),
		Logger:          logger,
	})

	mux.Handle("GET /get_value/{key}", &ggh.GGHandler[ExampleAppServiceProvider, struct{}, struct{}, GetValueResponse, ExampleAppErrorData]{
		ServiceProvider: sp.(*ExampleAppServiceProvider),
		HandlerFunc:     HandleGetValue,
		Middlewares:     getDefaultMiddlewares[struct{}, struct{}, GetValueResponse](),
		Logger:          logger,
	})

	return mux
}
