package sqlitedb

import (
	"context"
	"log"
	"os"
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/Milad75Rasouli/portfolio/internal/db"
	"github.com/Milad75Rasouli/portfolio/internal/store"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const timeLayout = "2006-01-02 15:04:05"

type StoreSqlite struct {
	*UserSqlite
	*HomeSqlite
	*BlogSqlite
	*ContactSqlite
	*AboutMeSqlite
}

func NewStoreSqlite(userDB *UserSqlite, homeDB *HomeSqlite, blogDB *BlogSqlite, contactDB *ContactSqlite, aboutMe *AboutMeSqlite) *StoreSqlite {
	return &StoreSqlite{
		UserSqlite:    userDB,
		HomeSqlite:    homeDB,
		BlogSqlite:    blogDB,
		ContactSqlite: contactDB,
		AboutMeSqlite: aboutMe,
	}
}

type SqliteInit struct {
	Folder string
}

func (d *SqliteInit) Init(isTestMode bool, config db.Config, logger *zap.Logger) (*StoreSqlite, func(), error) {
	var (
		userDB    *UserSqlite
		homeDB    *HomeSqlite
		blogDB    *BlogSqlite
		aboutMeDB *AboutMeSqlite
		contactDB *ContactSqlite
	)
	os.Mkdir(d.Folder, 0777)
	cfg := config
	if isTestMode == true {
		cfg = db.Config{
			IsSqlite:          true,
			ConnectionTimeout: time.Millisecond * 200,
		}
		defer func() {
			err := os.RemoveAll(d.Folder)
			if err != nil {
				log.Printf("Error removing database folder: %v", err)
			}
		}()
	}
	dbPool, err := db.New(cfg)
	if err != nil {
		return nil, nil, err
	}

	err = CreateSqliteTable(dbPool, cfg)
	if err != nil {
		return nil, nil, err
	}

	userDB = NewUserSqlite(dbPool, logger)
	homeDB = NewHomeSqlite(dbPool, logger)
	blogDB = NewBlogSqlite(dbPool, logger)
	contactDB = NewContactSqlite(dbPool, logger)
	aboutMeDB = NewAboutMeSqlite(dbPool, logger)
	store := NewStoreSqlite(userDB, homeDB, blogDB, contactDB, aboutMeDB)

	return store, func() {
		err := dbPool.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}, nil
}

func CreateSqliteTable(dbPool *sqlitex.Pool, cfg db.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectionTimeout)
	defer cancel()
	conn := dbPool.Get(ctx)
	defer dbPool.Put(conn)

	tables := []string{
		`CREATE TABLE IF NOT EXISTS user (
			id INTEGER,
			full_name TEXT NOT NULL,
			email INTEGER NOT NULL UNIQUE,
			password TEXT NOT NULL,
			is_github INTEGER DEFAULT 0,
			online_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified_at DATETIME,
			PRIMARY KEY(id)
		)`,
		`CREATE TABLE IF NOT EXISTS post (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			body TEXT NOT NULL,
			caption TEXT NOT NULL,
			image_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS category (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS post_category_relation (
			category_id INTEGER,
			post_id INTEGER,
			PRIMARY KEY(category_id, post_id),
			FOREIGN KEY(category_id) REFERENCES category (id),
			FOREIGN KEY(post_id) REFERENCES post (id)
		)`,
		`CREATE TABLE IF NOT EXISTS contact (
			id INTEGER,
			subject TEXT NOT NULL,
			email INTEGER NOT NULL,
			message TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(id)
		);`,
		`CREATE TABLE IF NOT EXISTS about_me (
			id INTEGER,
			content TEXT NOT NULL,
			PRIMARY KEY(id)
		);`,
		`CREATE TABLE IF NOT EXISTS home (
			id INTEGER,
			name TEXT NOT NULL,
			slogan TEXT NOT NULL,
			short_intro TEXT NOT NULL,
			github_url TEXT NOT NULL,
			PRIMARY KEY(id)
		);`,
	}

	for _, table := range tables {
		stmt, err := conn.Prepare(table)
		if err != nil {
			return errors.Wrap(err, store.CannotCreateTableError.Error())
		}
		defer stmt.Finalize()
		_, err = stmt.Step()
		if err != nil {
			return err
		}
	}

	return nil
}
