package stores

import "errors"

// StoreFactory - generate Store instance given user's setting
func StoreFactory(storeType string) (Store, error) {
	switch storeType {
	case "json":
		return JSONStore{}, nil
	case "sqlite":
		return SQLiteStore{}, nil
	}
	return nil, errors.New("Unknown Store type")
}
