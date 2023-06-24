package main

import (
	db "example.com/lords/stats/db"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	excelize "github.com/xuri/excelize/v2"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func main() {
	file := flag.String("db", "test.db", "location of SQLite DB file")
	botLocation := flag.String("bot", ".", "bot configuration")
	flag.Parse()

	accounts := listAccounts(*botLocation)
	stats, err := db.NewStats(*file)
	if err != nil {
		log.Fatal(err)
	}
	defer stats.Close()

	for _, account := range accounts {
		log.Printf("Process account %s", account)
		excelFilesLocation := fmt.Sprintf("%s/%s/stats/exported/", *botLocation, account)

		accountId, err := strconv.Atoi(account)
		if err != nil {
			fmt.Println("Error during conversion account to accountId")
			return
		}

		var files []string = findFiles(excelFilesLocation, ".xlsx")
		for _, filename := range files {
			log.Printf("- check if file %s is already processed", filename)
			isFileProcessed, err := stats.IsFileProcessed(accountId, filename)
			if err != nil {
				log.Fatal(err, filename)
			}
			if isFileProcessed {
				continue
			}

			log.Printf("- start processing %s file", filename)
			statsTx, err := stats.BeginTx()
			defer statsTx.Rollback()
			if err != nil {
				log.Fatal(err, filename)
			}

			statsTx.AddFileProcessed(accountId, filename)
			xlsx, err := loadExcel(filename)
			if err != nil {
				log.Println("fail process file ", filename, err)
				statsTx.Rollback()
				continue
			}

			var re = regexp.MustCompile(`(?m)(\d{4}-\d{2}-\d{1,2})`)
			var match = re.FindAllString(filename, -1)

			for _, row := range xlsx.rows {
				if len(row) < 25 {
					log.Fatalln("corrupted data ", row, filename)
				}
				if row[0] == "0" || row[25] == "0" || row[0] == "User ID" {
					continue
				}
				if err = statsTx.Insert(accountId, row, &match[0]); err != nil {
					log.Println("fail process row ", row, err)
				}
			}

			if err := statsTx.Commit(); err != nil {
				log.Fatal(err, filename)
			}
		}

	}
}

func listAccounts(root string) []string {
	var accounts []string
	entries, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(`\d+`)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if re.MatchString(e.Name()) {
			accounts = append(accounts, e.Name())
		}
	}
	return accounts
}

type ExcelFile struct {
	rows [][]string
}

func loadExcel(filename string) (*ExcelFile, error) {
	var rows [][]string

	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if rows, err = f.GetRows("Sheet"); err != nil {
		return nil, err
	}
	return &ExcelFile{
		rows: rows,
	}, nil
}

func findFiles(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		re := regexp.MustCompile(`^\d{4}-\d{2}-\d{1,2} \d{2}-\d{2} \w{3,}\.`)
		if filepath.Ext(d.Name()) == ext && re.MatchString(d.Name()) {
			a = append(a, s)
		}
		return nil
	})
	return a
}
