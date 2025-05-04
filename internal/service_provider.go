package app

import (
	"log/slog"

	gghyt "github.com/dmi-feo/gogohandlers-yt-batteries"
)

type ExampleAppServiceProviderInterface interface {
	GetStorage() Storage
	GetLogger() *slog.Logger
}

type ExampleAppServiceProvider struct {
	logger  *slog.Logger
	storage Storage
}

func (sp *ExampleAppServiceProvider) GetLogger() *slog.Logger {
	return sp.logger
}

func (sp *ExampleAppServiceProvider) GetStorage() Storage {
	return sp.storage
}

func NewExampleAppServiceProvider(appConfig AppConfig, logger *slog.Logger) (ExampleAppServiceProviderInterface, error) {
	easp := &ExampleAppServiceProvider{logger: logger}
	var err error

	switch appConfig.StorageType {
	case "sqlite":
		filePath := appConfig.StorageConfig.(map[string]any)["file_path"].(string)
		easp.storage, err = NewSQLiteStorage(filePath, logger)
	case "yt":
		ytConfig := appConfig.StorageConfig.(map[string]any)
		ytcFactory := gghyt.NewYtClientFactory(ytConfig["cluster"].(string), ytConfig["token"].(string), logger)
		ytc, err := ytcFactory.FromSettings()
		if err != nil {
			return nil, err
		}
		easp.storage, err = NewYtStorage(ytc, ytConfig["path"].(string), logger)
	}

	if err != nil {
		return nil, err
	}
	return easp, nil
}
