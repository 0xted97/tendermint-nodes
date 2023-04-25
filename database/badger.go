package database

import (
	"github.com/dgraph-io/badger"
)

type BadgerDatabase struct {
	db *badger.DB
}

func NewBadgerDatabase(path string) (*BadgerDatabase, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	return &BadgerDatabase{
		db: db,
	}, nil
}

func (d *BadgerDatabase) Close() error {
	return d.db.Close()
}

func (d *BadgerDatabase) GetDB() *badger.DB {
	return d.db
}
