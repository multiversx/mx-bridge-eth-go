package safe

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
)

type Blockstorer interface {
	StoreBlockIndex(*big.Int) error
}

type Blockreader interface {
	ReadBlockIndex() (*big.Int, error)
}

type Blockstore struct {
	dirPath  string
	fullPath string
}

type RelayType string

const (
	Eth    RelayType = "eth"
	Elrond RelayType = "elrond"
)

func NewBlockstore(dirPath string, relayType RelayType) (*Blockstore, error) {
	_, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	filename := getFileName(relayType)

	return &Blockstore{
		dirPath:  dirPath,
		fullPath: filepath.Join(dirPath, filename),
	}, nil
}

func (b *Blockstore) StoreBlockIndex(index *big.Int) error {
	if _, err := os.Stat(b.dirPath); os.IsNotExist(err) {
		er := os.MkdirAll(b.dirPath, os.ModePerm)
		if er != nil {
			return er
		}
	}

	err := ioutil.WriteFile(b.fullPath, []byte(index.String()), 0600)
	if err != nil {
		return err
	}

	return nil
}

func (b *Blockstore) LoadBlockIndex() (*big.Int, error) {
	if exists, _ := fileExists(b.fullPath); !exists {
		return big.NewInt(0), nil
	}

	data, err := ioutil.ReadFile(b.fullPath)
	if err != nil {
		return nil, err
	}

	index, _ := big.NewInt(0).SetString(string(data), 10)

	return index, nil
}

func getFileName(relayType RelayType) string {
	return fmt.Sprintf(".%v.block", relayType)
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, nil
	}

	return true, nil
}
