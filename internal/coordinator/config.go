package coordinator

import (
	"os"
	"sync"
)

var lock = &sync.Mutex{}

type Config struct {
	artifactsPath string
}

var configInstance *Config

func GetConfig() *Config {
	if configInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		artifactsPath, exists := os.LookupEnv("ARTIFACTS_PATH")
		if !exists {
			artifactsPath = "/coordinator/artifacts"
		}
		if artifactsPath[len(artifactsPath)-1] == '/' {
			artifactsPath = artifactsPath[:len(artifactsPath)-1]
		}
		configInstance = &Config{
			artifactsPath: artifactsPath,
		}
	}
	return configInstance
}

func (c *Config) GetArtifactsPath() string {
	return c.artifactsPath
}
