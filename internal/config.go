package app

type AppConfig struct {
	StorageType   string `yaml:"storage_type"`
	StorageConfig any    `yaml:"storage_config"`
}
