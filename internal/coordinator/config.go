package coordinator

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"

	"github.com/Assifar-Karim/apollo/internal/utils"
)

var lock = &sync.Mutex{}

type Config struct {
	devMode              bool
	artifactsPath        string
	splitSize            int64
	kubeConfigPath       string
	workerNS             string
	workerImg            string
	intermediateFilesLoc string
}

var configInstance *Config

func GetConfig() *Config {
	if configInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		args := os.Args[1:]
		devMode := false
		if slices.Contains(args, "--dev") {
			devMode = true
		}
		artifactsPath, exists := os.LookupEnv("ARTIFACTS_PATH")
		if !exists {
			artifactsPath = "/coordinator/artifacts"
		}
		if artifactsPath[len(artifactsPath)-1] == '/' {
			artifactsPath = artifactsPath[:len(artifactsPath)-1]
		}
		splitSizeStr, exists := os.LookupEnv("SPLIT_SIZE")
		var splitSize int64
		if !exists {
			splitSize = 67108864
		} else {
			conv, err := strconv.Atoi(splitSizeStr)
			if err != nil {
				splitSize = 67108864
				logger := utils.GetLogger()
				logger.Warn("can't read split size from SPLIT_SIZE environment variable, size will default to 67108864 bytes")
			} else {
				splitSize = int64(conv)
			}
		}
		kubeConfigPath, exists := os.LookupEnv("KUBECONFIG_PATH")
		if !exists {
			home, err := os.UserHomeDir()
			if err != nil {
				// In case of an error we suppose that the home can be found using ~
				home = "~"
			}
			kubeConfigPath = filepath.Join(home, ".kube/config")
		}

		workerNS, exists := os.LookupEnv("WORKER_NS")
		if !exists {
			workerNS = "apollo-workers"
		}

		workerImg, exists := os.LookupEnv("WORKER_IMG")
		if !exists {
			workerImg = "localhost:5000/worker:log-4" // TODO: Change with the actual worker image that will be uploaded later down the line
		}

		intermediateFilesLoc, exists := os.LookupEnv("INT_FILES_LOC")
		if !exists {
			intermediateFilesLoc = "/apollo/intermediate-files"
		}
		if intermediateFilesLoc[len(intermediateFilesLoc)-1] == '/' {
			intermediateFilesLoc = intermediateFilesLoc[:len(intermediateFilesLoc)-1]
		}
		configInstance = &Config{
			devMode:              devMode,
			artifactsPath:        artifactsPath,
			splitSize:            splitSize,
			kubeConfigPath:       kubeConfigPath,
			workerNS:             workerNS,
			workerImg:            workerImg,
			intermediateFilesLoc: intermediateFilesLoc,
		}

	}
	return configInstance
}

func (c *Config) isInDevMode() bool {
	return c.devMode
}

func (c *Config) GetArtifactsPath() string {
	return c.artifactsPath
}

func (c *Config) GetSplitSize() int64 {
	return c.splitSize
}

func (c *Config) GetKubeConfigPath() string {
	return c.kubeConfigPath
}

func (c *Config) GetWorkerNS() string {
	return c.workerNS
}

func (c *Config) GetWorkerImg() string {
	return c.workerImg
}

func (c *Config) GetIntermediateFilesLoc() string {
	return c.intermediateFilesLoc
}
