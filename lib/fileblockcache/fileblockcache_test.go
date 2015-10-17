package fileblockcache

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/protocol"
)

var (
	folder = "fileblockcache_test"
)

func GetSetGet(t *testing.T) {
	cfg, db := setup(t)
	fbc := NewFileBlockCache(cfg, db, folder, 1)

	hash := []byte("teh hash")

	// check empty get
	assertUnavailable(t, fbc, hash)

	// add data
	expectedData := []byte("dead beef")
	block := protocol.BlockInfo{
		Hash: hash,
		Size: int32(len(expectedData)),
	}
	fbc.AddCachedFileData(block, expectedData)

	// check full get
	assertAvailable(t, fbc, hash, expectedData)
}

func BlockGetsEvicted1(t *testing.T) {
	cfg, db := setup(t)
	fbc := NewFileBlockCache(cfg, db, folder, 2)

	data1 := []byte("data1")
	block1 := protocol.BlockInfo{
		Hash: []byte("hash1"),
		Size: 1,
	}
	fbc.AddCachedFileData(block1, data1)
	assertAvailable(t, fbc, block1.Hash, data1)

	data2 := []byte("data2")
	block2 := protocol.BlockInfo{
		Hash: []byte("hash2"),
		Size: 1,
	}
	fbc.AddCachedFileData(block2, data2)
	assertAvailable(t, fbc, block1.Hash, data1)
	assertAvailable(t, fbc, block2.Hash, data2)

	data3 := []byte("data3")
	block3 := protocol.BlockInfo{
		Hash: []byte("hash3"),
		Size: 1,
	}
	fbc.AddCachedFileData(block3, data3)

	assertAvailable(t, fbc, block2.Hash, data2)
	assertAvailable(t, fbc, block3.Hash, data3)
	assertUnavailable(t, fbc, block1.Hash)
}

func BlockGetsEvicted2(t *testing.T) {
	cfg, db := setup(t)
	fbc := NewFileBlockCache(cfg, db, folder, 2)

	data1 := []byte("data1")
	block1 := protocol.BlockInfo{
		Hash: []byte("hash1"),
		Size: 1,
	}
	fbc.AddCachedFileData(block1, data1)

	data2 := []byte("data2")
	block2 := protocol.BlockInfo{
		Hash: []byte("hash2"),
		Size: 1,
	}
	fbc.AddCachedFileData(block2, data2)

	assertAvailable(t, fbc, block1.Hash, data1)
	assertAvailable(t, fbc, block2.Hash, data2)

	data3 := []byte("data3")
	block3 := protocol.BlockInfo{
		Hash: []byte("hash3"),
		Size: 1,
	}
	fbc.AddCachedFileData(block3, data3)

	assertUnavailable(t, fbc, block1.Hash)
	assertAvailable(t, fbc, block2.Hash, data2)
	assertAvailable(t, fbc, block3.Hash, data3)
}

func EvictMultipleBlocks(t *testing.T) {
	cfg, db := setup(t)
	fbc := NewFileBlockCache(cfg, db, folder, 2)

	data1 := []byte("data1")
	block1 := protocol.BlockInfo{
		Hash: []byte("hash1"),
		Size: 1,
	}
	fbc.AddCachedFileData(block1, data1)

	data2 := []byte("data2")
	block2 := protocol.BlockInfo{
		Hash: []byte("hash2"),
		Size: 1,
	}
	fbc.AddCachedFileData(block2, data2)

	assertAvailable(t, fbc, block1.Hash, data1)
	assertAvailable(t, fbc, block2.Hash, data2)

	data3 := []byte("data3")
	block3 := protocol.BlockInfo{
		Hash: []byte("hash3"),
		Size: 2,
	}
	fbc.AddCachedFileData(block3, data3)

	assertUnavailable(t, fbc, block1.Hash)
	assertUnavailable(t, fbc, block2.Hash)
	assertAvailable(t, fbc, block3.Hash, data3)
}

func assertAvailable(t *testing.T, fbc *FileBlockCache, hash []byte, expectedData []byte) {
	actualData, found := fbc.GetCachedBlockData(hash)
	if false == found {
		t.Error("entry should exist")
	}
	if len(actualData) != len(expectedData) {
		t.Error("actual data", len(actualData), "and expected data", len(expectedData), "sizes differ")
	}
	for i := 0; i < len(expectedData); i++ {
		if actualData[i] != expectedData[i] {
			t.Error("actual data mismatches expected data at index", i)
		}
	}
}

func assertUnavailable(t *testing.T, fbc *FileBlockCache, hash []byte) {
	_, found := fbc.GetCachedBlockData(hash)
	if found {
		t.Error("entry should not exist, but does. hash", string(hash))
	}
}

func setup(t *testing.T) (*config.Wrapper, *bolt.DB) {
	dir, _ := ioutil.TempDir("", "stf-mt")
	defer os.RemoveAll(dir)
	configFile, _ := ioutil.TempFile(dir, "config")
	deviceID, _ := protocol.DeviceIDFromString("FFR6LPZ-7K4PTTV-UXQSMUU-CPQ5YWH-OEDFIIQ-JUG777G-2YQXXR5-YD6AWQR")
	realCfg := config.New(deviceID)
	cfg := config.Wrap(configFile.Name(), realCfg)
	t.Logf("config path %s\n", configFile.Name())

	databasePath := path.Join(path.Dir(cfg.ConfigPath()), "boltdb")
	database, _ := bolt.Open(databasePath, 0600, nil)

	folderCfg := config.FolderConfiguration{
		ID: folder,
	}
	cfg.SetFolder(folderCfg)

	return cfg, database
}