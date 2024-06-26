package sqlitedb

import (
	"context"
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Milad75Rasouli/portfolio/internal/model"
	"github.com/Milad75Rasouli/portfolio/internal/store"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type UserSqlite struct {
	dbPool *sqlitex.Pool
	logger *zap.Logger
}

func NewUserSqlite(dbPool *sqlitex.Pool, logger *zap.Logger) *UserSqlite {
	return &UserSqlite{
		dbPool: dbPool,
		logger: logger,
	}
}
func (u UserSqlite) parseToUser(stmt *sqlite.Stmt) (model.User, error) {
	var (
		usr model.User
		err error
	)
	usr.ID = stmt.GetInt64("id")
	usr.FullName = stmt.GetText("full_name")
	usr.Email = stmt.GetText("email")
	usr.Password = stmt.GetText("password")
	usr.IsGithub = stmt.GetInt64("is_github")
	usr.OnlineAt, err = time.Parse(timeLayout, stmt.GetText("online_at"))
	if err != nil {
		return usr, err
	}
	usr.CreatedAt, err = time.Parse(timeLayout, stmt.GetText("created_at"))
	if err != nil {
		return usr, err
	}
	usr.ModifiedAt, err = time.Parse(timeLayout, stmt.GetText("modified_at"))
	return usr, err
}
func (u *UserSqlite) CreateUser(ctx context.Context, usr model.User) (int64, error) {
	var rowID int64
	conn := u.dbPool.Get(ctx)
	defer u.dbPool.Put(conn)

	stmt, err := conn.Prepare(`INSERT INTO user (full_name, email, password, is_github,online_at, modified_at, created_at)
	VALUES($1,$2,$3,$4,$5,$6,$7);`)
	if err != nil {
		return rowID, errors.Errorf("unable to create the new user %s", err.Error())
	}
	defer stmt.Finalize()

	stmtSelect, err := conn.Prepare(`SELECT last_insert_rowid();`)
	if err != nil {
		return 0, errors.Errorf("unable to prepare the select statement: %s", err.Error())
	}
	defer stmtSelect.Finalize()

	stmt.SetText("$1", usr.FullName)
	stmt.SetText("$2", usr.Email)
	stmt.SetText("$3", usr.Password)
	stmt.SetInt64("$4", usr.IsGithub)
	stmt.SetText("$5", usr.OnlineAt.Format(timeLayout))
	stmt.SetText("$6", usr.ModifiedAt.Format(timeLayout))
	stmt.SetText("$7", usr.CreatedAt.Format(timeLayout))

	_, err = stmt.Step()
	if err != nil {
		e := err.Error()[18:42]
		if e == "SQLITE_CONSTRAINT_UNIQUE" {
			return rowID, store.DuplicateUserError
		}
		return rowID, err
	}
	hasRow, err := stmtSelect.Step()
	if err != nil {
		return rowID, err
	}

	if hasRow {
		rowID = conn.LastInsertRowID()
	}
	return rowID, err
}
func (u *UserSqlite) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	var usr model.User
	conn := u.dbPool.Get(ctx)
	defer u.dbPool.Put(conn)
	stmt, err := conn.Prepare(`SELECT * FROM user WHERE email=$1 LIMIT 1;`)
	if err != nil {
		return usr, errors.Errorf("unable to get the user %s from email", err.Error())
	}
	defer stmt.Finalize()
	stmt.SetText("$1", email)

	var hasRow bool
	hasRow, err = stmt.Step()
	if hasRow == false {
		return usr, store.UserNotFountError
	}
	if err != nil {
		return usr, err
	}
	usr, err = u.parseToUser(stmt)
	return usr, err
}
func (u *UserSqlite) GetUserByID(ctx context.Context, id int64) (model.User, error) {
	var usr model.User
	conn := u.dbPool.Get(ctx)
	defer u.dbPool.Put(conn)
	stmt, err := conn.Prepare(`SELECT * FROM user WHERE id=$1 LIMIT 1;`)
	if err != nil {
		return usr, errors.Errorf("unable to get the user %s from id", err.Error())
	}
	defer stmt.Finalize()
	stmt.SetInt64("$1", id)

	var hasRow bool
	hasRow, err = stmt.Step()
	if hasRow == false {
		return usr, store.UserNotFountError
	}
	if err != nil {
		return usr, err
	}
	usr, err = u.parseToUser(stmt)
	return usr, err
}
func (u *UserSqlite) GetAllUser(ctx context.Context) ([]model.User, error) {
	var usr []model.User
	conn := u.dbPool.Get(ctx)
	defer u.dbPool.Put(conn)
	stmt, err := conn.Prepare(`SELECT * FROM user;`)
	if err != nil {
		return usr, errors.Errorf("unable to get all users %s", err.Error())
	}
	defer stmt.Finalize()

	for {
		var (
			swapUser model.User
			hasRow   bool
		)
		hasRow, err = stmt.Step()
		if hasRow == false {
			break
		}
		swapUser, err = u.parseToUser(stmt)
		if err != nil {
			return usr, errors.Errorf("getting the users from database error %s", err.Error())
		}
		usr = append(usr, swapUser)
	}
	return usr, err
}
func (u *UserSqlite) DeleteUserByID(ctx context.Context, id int64) error {
	conn := u.dbPool.Get(ctx)
	defer u.dbPool.Put(conn)
	stmt, err := conn.Prepare(`DELETE FROM user WHERE id=$1;`)
	if err != nil {
		return err
	}
	defer stmt.Finalize()
	stmt.SetInt64("$1", id)
	_, err = stmt.Step()
	return err
}
func (u *UserSqlite) UpdateUserByPasswordFullName(ctx context.Context, id int64, password string, fullname string) error {
	conn := u.dbPool.Get(ctx)
	defer u.dbPool.Put(conn)
	var s string
	if len(password) != 0 && len(fullname) != 0 {
		s = fmt.Sprintf(`UPDATE user
		SET password='%s', full_name='%s'
		WHERE id=%d;`, password, fullname, id)
	} else if len(password) != 0 {
		s = fmt.Sprintf(`UPDATE user
		SET password='%s'
		WHERE id=%d;`, password, id)
	} else if len(fullname) != 0 {
		s = fmt.Sprintf(`UPDATE user
		SET full_name='%s'
		WHERE id=%d;`, fullname, id)
	}
	stmt, err := conn.Prepare(s)
	if err != nil {
		return err
	}
	defer stmt.Finalize()
	var hasRow bool
	hasRow, err = stmt.Step()
	if hasRow {
		return store.UserNotFountError
	}
	return err
}
