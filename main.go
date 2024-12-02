package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/bluesky-social/jetstream/pkg/client"
	"github.com/bluesky-social/jetstream/pkg/models"

	//"github.com/bluesky-social/jetstream/pkg/client/schedulers"
	"github.com/bluesky-social/jetstream/pkg/client/schedulers/parallel"
	"github.com/google/uuid"
	pg "github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
)

var connlite *sql.DB
var connpsql *pg.Conn
var dbtype *string

func separator(s string) []string {
	s = strings.Trim(s, "[]")
	sr := strings.Split(s, "\"")
	var ret []string
	for i, v := range sr {
		if v == ", " || v == "" || v == "," || v == " " || v == `"` {
			i++
		} else {
			ret = append(ret, v)
		}
	}
	return ret
}
func eventHandler(ctx context.Context, e *models.Event) error {
	id, err := uuid.NewV7()
	if err != nil {
		log.Fatal(err)
	}

	query := `
        INSERT INTO Event (
            InternalID, Did, TimeUS, Kind, 
            CommitRev, CommitOperation, CommitCollection, CommitRKey, CommitRecord, CommitCID, 
            AccountActive, AccountDid, AccountSeq, AccountStatus, AccountTime, 
            IdentityDid, IdentityHandle, IdentitySeq, IdentityTime
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
    `

	var (
		commitRev, commitOperation, commitCollection, commitRKey, commitCID string
		commitRecord                                                        []byte
		accountActive                                                       bool
		accountDid, accountStatus, accountTime                              string
		accountSeq                                                          int64
		identityDid, identityHandle, identityTime                           string
		identitySeq                                                         int64
	)

	if e.Commit != nil {
		commitRev = e.Commit.Rev
		commitOperation = e.Commit.Operation
		commitCollection = e.Commit.Collection
		commitRKey = e.Commit.RKey
		commitRecord = e.Commit.Record
		commitCID = e.Commit.CID
	}

	if e.Account != nil {
		accountActive = e.Account.Active
		accountDid = e.Account.Did
		accountSeq = e.Account.Seq
		if e.Account.Status != nil {
			accountStatus = *e.Account.Status
		}
		accountTime = e.Account.Time
	}

	if e.Identity != nil {
		identityDid = e.Identity.Did
		if e.Identity.Handle != nil {
			identityHandle = *e.Identity.Handle
		}
		identitySeq = e.Identity.Seq
		identityTime = e.Identity.Time
	}

	args := []interface{}{
		id, e.Did, e.TimeUS, e.Kind,
		commitRev, commitOperation, commitCollection, commitRKey, commitRecord, commitCID,
		accountActive, accountDid, accountSeq, accountStatus, accountTime,
		identityDid, identityHandle, identitySeq, identityTime,
	}

	if *dbtype == "sqlite" {
		_, err := connlite.Exec(query, args...)
		if err != nil {
			log.Fatal(err)
		}
	} else if *dbtype == "postgres" {
		_, err := connpsql.Exec(ctx, query, args...)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
func main() {
	dbtype = flag.String("t", "sqlite", "database type (sqlite, postgres)")
	sqlite := flag.String("s", "output.sqlite", "sqlite db path")
	pgau := flag.String("p", "postgres://postgres:password@localhost:5432/postgres", "postgres connection string")
	params := flag.String("q", `["app.bsky.*"]`, "array of strings to filter the messages")
	wanteddids := flag.String("w", "", "array of dids to filter the messages")
	server := flag.String("r", "wss://jetstream1.us-east.bsky.network/subscribe", "wss uri to connect to")
	// qu := flag.String("q", "", "params to add to the wss uri")
	flag.Parse()
	// websocket.DefaultDialer.Dial("wss://jetstream1.us-east.bsky.network/subscribe?compress=true" + *qu, http.Header{})
	var wd []string
	pr := separator(*params)
	if *wanteddids != "" {
		wd = separator(*wanteddids)
	} else {
		wd = []string{}
	}
	com := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	sl := slog.New(com)
	//ctx := context.Background()
	sch := parallel.NewScheduler(1, "bgcat", sl, eventHandler)
	jsclient, err := client.NewClient(
		&client.ClientConfig{
			Compress:          true,
			WebsocketURL:      *server,
			WantedDids:        wd,
			WantedCollections: pr,
			MaxSize:           0,
			ExtraHeaders:      map[string]string{},
		},
		sl,
		sch,
	)
	if err != nil {
		log.Fatal(err)
	}
	if *dbtype == "sqlite" {
		var err error
		connlite, err = sql.Open("sqlite3", *sqlite)
		if err != nil {
			log.Fatal(err)
		}
		defer connlite.Close()
		_, err = connlite.Exec(
			`CREATE TABLE IF NOT EXISTS Event (
			InternalID TEXT PRIMARY KEY,
			Did TEXT NOT NULL,
			TimeUS INTEGER NOT NULL,
			Kind TEXT,
			CommitRev TEXT,
			CommitOperation TEXT,
			CommitCollection TEXT,
			CommitRKey TEXT,
			CommitRecord BLOB,
			CommitCID TEXT,
			AccountActive BOOLEAN,
			AccountDid TEXT,
			AccountSeq INTEGER,
			AccountStatus TEXT,
			AccountTime TEXT,
			IdentityDid TEXT,
			IdentityHandle TEXT,
			IdentitySeq INTEGER,
			IdentityTime TEXT
		);`)
		if err != nil {
			log.Fatal(err)
		}
	} else if *dbtype == "postgres" {
		var err error
		connpsql, err = pg.Connect(context.Background(), *pgau)
		if err != nil {
			log.Fatal(err)
		}
		defer connpsql.Close(context.Background())
	}
	ctx := context.Background()
	cursor := time.Now().UnixMicro()
	/* go func() {
		for {
			*cursor = time.Now().UnixMicro()
			time.Sleep(1 * time.Microsecond)
		}
	}() */
	jsclient.ConnectAndRead(ctx, &cursor)
}
