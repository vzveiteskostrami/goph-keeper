package pdb

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/goph-keeper/internal/server/config"
	"github.com/vzveiteskostrami/goph-keeper/internal/server/logging"
)

type PGStorage struct {
	db *sql.DB
}

// вот они, миграции!
func (d *PGStorage) tableInitData() error {
	if d.db == nil {
		return errors.New("база данных не инициализирована")
	}

	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		logging.S().Infoln("Ошибка 1:", err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		logging.S().Infoln("Ошибка 2:", err)
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		logging.S().Infoln("Ошибка 3:", err)
		return err
	}

	return nil
}

func (d *PGStorage) Init() {
	var err error

	d.db, err = sql.Open("postgres", *config.Get().DatabaseDSN)
	if err != nil {
		logging.S().Panic(err)
	}
	logging.S().Infof("Объявлено соединение с %s", *config.Get().DatabaseDSN)

	err = d.db.Ping()
	if err != nil {
		logging.S().Panic(err)
	}
	logging.S().Infof("Установлено соединение с %s", *config.Get().DatabaseDSN)
	err = d.tableInitData()
	if err != nil {
		logging.S().Panic(err)
	}
}

func (d *PGStorage) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
