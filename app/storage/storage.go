package storage

import (
	"fmt"

	"github.com/networkteam/filestore"
	"github.com/networkteam/filestore/local"
	"github.com/networkteam/filestore/memory"

	"github.com/tanema/dicom-microservice-api/config"
)

type Storage filestore.FileStore

func New(cfg config.Storage) (filestore.FileStore, error) {
	switch cfg.Kind {
	case "memory":
		return memory.NewFilestore(), nil
	case "local":
		return local.NewFilestore(cfg.TmpPath, cfg.AssetPath)
	case "s3":
		return nil, fmt.Errorf("s3 not implemented yet")
	default:
		return nil, fmt.Errorf("unknown store type %v", cfg.Kind)
	}
}
