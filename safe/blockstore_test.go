package safe

import (
	"io/ioutil"
	"math/big"
	"os"
	"reflect"
	"testing"
)

func TestStoreAndRead(t *testing.T) {
	t.Run("will store", func(t *testing.T) {
		dir, err := ioutil.TempDir(os.TempDir(), "blockstore")
		checkError(t, err)
		defer os.RemoveAll(dir)

		createBlockStore := func(RelayType) *Blockstore {
			blockstore, err := NewBlockstore(dir, Eth)
			checkError(t, err)

			return blockstore
		}

		var storeReadTests = []struct {
			name  string
			store *Blockstore
			value *big.Int
		}{
			{"for eth blockstore", createBlockStore(Eth), big.NewInt(42)},
			{"for elr blockstore", createBlockStore(Elrond), big.NewInt(42)},
		}

		for _, tt := range storeReadTests {
			t.Run(tt.name, func(t *testing.T) {
				err = tt.store.StoreBlockIndex(tt.value)
				checkError(t, err)

				got, err := tt.store.LoadBlockIndex()
				checkError(t, err)

				if !reflect.DeepEqual(got, tt.value) {
					t.Errorf("Expected %v, got %v", tt.value, got)
				}
			})
		}
	})

	t.Run("will read a default value when there is no file yet", func(t *testing.T) {
		dir, err := ioutil.TempDir(os.TempDir(), "blockstore")
		checkError(t, err)
		defer os.RemoveAll(dir)

		blockstore, err := NewBlockstore(dir, Eth)
		checkError(t, err)

		expected := big.NewInt(0)
		got, err := blockstore.LoadBlockIndex()
		checkError(t, err)

		if !reflect.DeepEqual(expected, got) {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	})
}

func checkError(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
