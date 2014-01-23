package backend

import (
	"container/list"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"log"
)

var (
	ErrUserNotExisting = errors.New("User does not exist")
	ErrUserExists      = errors.New("User already exists")
	globalDb           *db
)

func AddUser(id int64) error {
	return globalDb.userAdd(id)
}

func DelUser(id int64) error {
	return globalDb.userDel(id)
}

type db struct {
	conn                                                                      *sql.DB
	userAddStmt, userDelStmt, gcmRegAdd, gcmRegDel, gcmRegFetch, gcmUpdateReg *sql.Stmt
}

func ConnectDb(connstr string) (e error) {
	globalDb, e = dialDb(connstr)
	return
}

func CloseDb() error {
	return globalDb.close()
}

func dialDb(connstr string) (*db, error) {
	log.Println("Connecting to postgresql...")
	conn, e := sql.Open("postgres", connstr)

	if e != nil {
		return nil, e
	}

	if e = conn.Ping(); e != nil {
		return nil, e
	}

	dbInst := new(db)

	dbInst.conn = conn

	e = dbInst.gcmInitStmt()

	if e != nil {
		return nil, e
	}

	dbInst.userAddStmt, e = conn.Prepare("INSERT INTO USERS VALUES ($1)")

	if e != nil {
		return nil, e
	}

	dbInst.userDelStmt, e = conn.Prepare("DELETE FROM USERS WHERE ID = $1")

	if e != nil {
		return nil, e
	}

	return dbInst, nil
}

func (db *db) close() (e error) {

	if e = db.gcmCloseStmt(); e != nil {
		return
	}

	if e = db.userAddStmt.Close(); e != nil {
		return
	}

	if e = db.userDelStmt.Close(); e != nil {
		return
	}

	return db.conn.Close()

}

func (db *db) probe() error {
	return db.conn.Ping()
}

func (db *db) users() (*list.List, error) {

	people := list.New()

	rows, e := db.conn.Query("SELECT ID FROM USERS")

	if e != nil {
		return nil, e
	}

	var id int64

	for rows.Next() {
		if e = rows.Scan(&id); e != nil {
			return nil, e
		}

		people.PushBack(id)
	}

	if e = rows.Err(); e != nil {
		return nil, e
	}

	return people, nil

}

func (db *db) userAdd(id int64) error {
	log.Printf("Adding user %d...", id)

	_, e := db.userAddStmt.Exec(id)

	return e
}

func (db *db) userDel(id int64) error {
	log.Printf("Deleting user %d...", id)

	_, e := db.userDelStmt.Exec(id)

	return e

}

func InitDb(connstr string) error {

	log.Println("Connecting to postgresql...")
	conn, e := sql.Open("postgres", connstr)

	if e != nil {
		return e
	}

	if e = conn.Ping(); e != nil {
		return e
	}

	dbInst := new(db)

	dbInst.conn = conn

	if e != nil {
		return e
	}

	log.Println("Connected.\nCreating table USERS...")

	_, e = dbInst.conn.Exec("CREATE TABLE USERS (ID BIGINT PRIMARY KEY CHECK (ID > -1))")

	if e != nil {
		return e
	}

	log.Println("Done.\nCreating table GCM...")

	if e = dbInst.gcmInitTable(); e != nil {
		return e
	}

	log.Println("Done.")

	return nil

}