package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var AppVer = "0.1a"
var db *sql.DB

var filestore string
var fileperm os.FileMode = 0664

func main() {
	var err error
	var server, port, user, password, database string
	var pFlagCommand, pFlagEnvfile string

	log.Printf("App version: %s", AppVer)

	flag.StringVar(&pFlagCommand, "cmd", "", "Command to sql: pull/push")
	flag.StringVar(&pFlagCommand, "c", "", "Command to sql: pull/push (shorthand)")

	flag.StringVar(&pFlagEnvfile, "env", ".env", "Enveroment file")
	flag.StringVar(&pFlagEnvfile, "e", ".env", "Enveroment file  (shorthand)")

	flag.Parse()

	err = godotenv.Load(pFlagEnvfile) //Загрузить файл .env
	if err != nil {
		log.Fatal("Error loading env file", err.Error())
	}

	server = os.Getenv("SQLSERVER")
	port = os.Getenv("PORT")
	database = os.Getenv("DATABASE")
	user = os.Getenv("USERNAME")
	password = os.Getenv("PASSWORD")
	filestore = os.Getenv("FILESTORE")

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;", server, user, password, port, database)

	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}

	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Connected.")

	var VersionSQL string
	err = db.QueryRow("SELECT @@VERSION").Scan(&VersionSQL)
	if err != nil {
		log.Println(err)
	}

	log.Println(VersionSQL)

	if pFlagCommand == "pull" {
		PullProc()
		PullView()
	}
}

func PullProc() {
	ctx := context.Background()

	os.MkdirAll(filestore+"/PROC", 0777)

	tsql := "EXECUTE sp_stored_procedures"
	rows, err := db.QueryContext(ctx, tsql)
	if err != nil {
		log.Println(err)
	}

	// Бежим по списку (исключаем PROCEDURE_OWNER = sys)
	for rows.Next() {
		var qualifier, owner, name, inparams, outparams, resultsets, proctype string
		var remarks sql.NullString

		// Get values from row.
		err := rows.Scan(&qualifier, &owner, &name, &inparams, &outparams, &resultsets, &remarks, &proctype)
		if err != nil {
			log.Println(err)
		}
		name = strings.Split(name, ";")[0]

		if owner != "sys" {

			SpText := ""
			SpChunk := ""

			tsql = fmt.Sprintf("EXECUTE sp_helptext N'%s'", name)

			SpRows, err := db.QueryContext(ctx, tsql)
			if err != nil {
				log.Println(err)
			}

			for SpRows.Next() {
				err = SpRows.Scan(&SpChunk)
				SpText += SpChunk
			}

			// Получаем и сохраняем каждую процедуру в файл
			log.Printf("Qualifier: %s, Name: %s\n", qualifier, name)

			err = ioutil.WriteFile(filestore+"/PROC/"+name+".sql", []byte(SpText), fileperm)
			if err != nil {
				log.Println(err)
			}

		}
	}

}

func PullView() {
	ctx := context.Background()

	os.MkdirAll(filestore+"/VIEW", 0777)

	// Запрашиваем список вьюх
	// EXECUTE sp_stored_procedures
	tsql := "SELECT TABLE_NAME, VIEW_DEFINITION FROM INFORMATION_SCHEMA.VIEWS WHERE TABLE_SCHEMA = 'dbo'"
	rows, err := db.QueryContext(ctx, tsql)
	if err != nil {
		log.Println(err)
	}

	// Бежим по списку (исключаем PROCEDURE_OWNER = sys)
	for rows.Next() {
		var name, vText string

		// Get values from row.
		err := rows.Scan(&name, &vText)
		if err != nil {
			log.Println(err)
		}

		// Получаем и сохраняем каждую процедуру в файл
		log.Printf("Name: %s\n", name)

		err = ioutil.WriteFile(filestore+"/VIEW/"+name+".sql", []byte(vText), fileperm)
		if err != nil {
			fmt.Println(err)
		}

	}

}
