package migrator

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

var db *pgx.Conn

func Migrate(connectionString string, migrationFilesFolder string) {
	setupDatabaseConnection(connectionString)
	defer db.Close(context.Background())

	if isFirstRun() {
		createMigrationTable()
	}

	files := getFilesInFolder(migrationFilesFolder)

	for _, file := range files {
		if isAlreadyApplied(file) {
			assertFileNotChanged(file)
		} else {
			fmt.Println("Start applying migration " + file + " on database")

			apply(file)

			fmt.Println("Applied migration " + file + " on database")
		}
	}
}

func isFirstRun() bool {
	var migrationTableExists bool

	err := db.QueryRow(context.Background(),
		"SELECT EXISTS ("+
			"	SELECT FROM information_schema.tables"+
			"	WHERE  table_schema = 'public'"+ // TODO: get configuration from pgx
			"	AND    table_name   = 'migration'"+
			"	);").Scan(&migrationTableExists)

	if err != nil {
		panic(err)
	}

	return !migrationTableExists
}

func createMigrationTable() {
	fmt.Println("Start creating migration table")

	_, err := db.Exec(context.Background(),
		"CREATE TABLE migration ("+
			"	ID BIGINT NOT NULL PRIMARY KEY,"+
			"	FILE_NAME TEXT NOT NULL,"+
			"	EXECUTED_AT BIGINT,"+
			"	HASH TEXT"+
			")")

	if err != nil {
		panic(err)
	}

	fmt.Println("Migration table created")
}

func setupDatabaseConnection(connectionString string) {
	conn, err := pgx.Connect(context.Background(), connectionString)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	db = conn
}

func getFilesInFolder(migrationFilesFolder string) []string {
	pattern := filepath.Join(migrationFilesFolder, "*.sql")

	fmt.Println(pattern)

	files, err := filepath.Glob(pattern)
	fmt.Println(files)

	if err != nil {
		panic(err)
	}

	sort.Strings(files)

	return files
}

func isAlreadyApplied(file string) bool {


	return true
}


func assertFileNotChanged(file string) {
	var hash string

	err := db.QueryRow(context.Background(),
		"SELECT hash FROM migration WHERE id = $1", getId(file)).Scan(&hash)

	if err != nil {
		panic(err)
	}

	if calculateHashForFile(file) != hash {
		panic("File " + file + " has been changed since change is applied")
	}
}

// TODO: strip out all newlines, tabs, spaces and other invisible characters before calculating the hash
func calculateHashForFile(file string) string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func getId(file string) int64 {
	id, err := strconv.ParseInt(strings.Split(filepath.Base(file), "_")[0], 10, 32)

	if err != nil {
		panic(err)
	}

	return id
}

func apply(file string) {

	fileName := filepath.Base(file)
	id := getId(fileName)


	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	text := string(content)

	tx, err := db.Begin(context.Background())
	if err != nil {
		panic(err)
	}

	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), text)

	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(context.Background(),
		"INSERT INTO migration (id, file_name, executed_at, hash) VALUES ($1, $2, $3, $4);",
		id, fileName, time.Now().Unix(), calculateHashForFile(file))

	if err != nil {
		panic(err)
	}

	tx.Commit(context.Background())
}
