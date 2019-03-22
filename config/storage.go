package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spikeekips/cvc"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spikeekips/naru/common"
)

type Storage struct {
	cvc.BaseGroup
	LevelDB *LevelDBStorage `flag:"leveldb"`
	Mongo   *MongoStorage
}

func NewStorage() *Storage {
	return &Storage{
		LevelDB: NewLevelDBStorage(),
		Mongo:   NewMongoStorage(),
	}
}

type LevelDBStorage struct {
	cvc.BaseGroup
	Path     string `flag-help:"storage path"`
	Scheme   string `flag:"-"`
	RealPath string `flag:"-"`
}

func NewLevelDBStorage() *LevelDBStorage {
	p := filepath.Join(common.CurrentDirectory, common.DefaultStoragePath)
	return &LevelDBStorage{
		Path:     "file://" + p,
		Scheme:   "file",
		RealPath: p,
	}
}

func (l LevelDBStorage) Type() string {
	return "leveldb"
}

func (l LevelDBStorage) ParsePath(s string) (string, error) {
	var scheme, path string
	{
		u, err := url.Parse(s)
		if err != nil {
			return "", err
		}
		scheme = u.Scheme
		path = u.Path
	}

	switch strings.ToLower(scheme) {
	case "memory":
		return "memory://", nil
	case "", "file":
		if !strings.HasPrefix(path, "/") {
			path = filepath.Join(common.CurrentDirectory, path)
		}
	default:
		return "", fmt.Errorf("unknown storage type")
	}

	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return "", fmt.Errorf("storage path is not directory")
		}
	}

	return s, nil
}

func (l *LevelDBStorage) Validate() error {
	var scheme, path string
	{
		u, err := url.Parse(l.Path)
		if err != nil {
			return err
		}
		scheme = u.Scheme
		path = u.Path
	}

	switch strings.ToLower(scheme) {
	case "memory":
		l.Scheme = "memory"
		l.RealPath = ""
	case "", "file":
		l.Scheme = "file"
		l.RealPath = path
		if !strings.HasPrefix(path, "/") {
			l.RealPath = filepath.Join(common.CurrentDirectory, path)
		}
	default:
		return fmt.Errorf("unknown storage type")
	}

	return nil
}

// MongoStorage is based on 'Connection String Options' of mongodb. See
// https://docs.mongodb.com/manual/reference/connection-string/#connections-connection-options
type MongoStorage struct {
	cvc.BaseGroup
	URI *mongooptions.ClientOptions
	DB  string
}

func NewMongoStorage() *MongoStorage {
	return &MongoStorage{
		URI: mongooptions.Client().ApplyURI("mongodb://localhost:27017"),
		DB:  "naru",
	}
}

func (m MongoStorage) Type() string {
	return "mongo"
}

func (m MongoStorage) FlagValueURI() string {
	return ""
}

func (m MongoStorage) ParseURI(s string) (*mongooptions.ClientOptions, error) {
	if _, err := url.Parse(s); err != nil {
		return nil, err
	}

	options := mongooptions.Client().ApplyURI(s)
	if err := options.Validate(); err != nil {
		return nil, err
	}

	return options, nil
}
