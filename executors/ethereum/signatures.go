package ethereum

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	logger "github.com/multiversx/mx-chain-logger-go"
)

const filesPattern = "0x*.json"

// LoadAllSignatures can load all valid signatures from the specified directory
func LoadAllSignatures(logger logger.Logger, path string) []SignatureInfo {
	filesContents, err := getAllFilesContents(path)
	if err != nil {
		logger.Warn(err.Error())
		return make([]SignatureInfo, 0)
	}

	signatures := make([]SignatureInfo, 0, len(filesContents))
	for _, buff := range filesContents {
		sigInfo := &SignatureInfo{}
		err = json.Unmarshal(buff, sigInfo)
		if err != nil {
			logger.Warn("error unmarshalling to json", "error", err)
			continue
		}

		signatures = append(signatures, *sigInfo)
	}

	return signatures
}

func getAllFilesContents(dirPath string) ([][]byte, error) {
	dirInfo, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("%w while fetching %s directory contents", err, dirPath)
	}

	data := make([][]byte, 0, len(dirInfo))
	for _, di := range dirInfo {
		if di.IsDir() {
			continue
		}
		matched, errMatched := filepath.Match(filesPattern, di.Name())
		if errMatched != nil || !matched {
			continue
		}

		buff, errRead := os.ReadFile(path.Join(dirPath, di.Name()))
		if errRead != nil {
			continue
		}

		data = append(data, buff)
	}

	return data, nil
}
