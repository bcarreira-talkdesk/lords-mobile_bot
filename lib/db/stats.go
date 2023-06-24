package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const sqlTimeLayout string = "2006-01-02 15:04:05"
const dbSchema string = `
CREATE TABLE IF NOT EXISTS gifts(
    account_id INT,
	user_id INT, 
	date DATE,
	name TEXT, 
	hunt_l1 INT,
	hunt_l2 INT,
	hunt_l3 INT,
	hunt_l4 INT,
	hunt_l5 INT,
	purchase_l1 INT,
	purchase_l2 INT,
	purchase_l3 INT,
	purchase_l4 INT,
	purchase_l5 INT,
	hunt_points INT,
	purchase_points INT,
	goal_perc_hunt INT,
	goal_perc_purchase INT,
	PRIMARY KEY (account_id, user_id, date)
	);

CREATE TABLE IF NOT EXISTS files(
    account_id INT,
	filename TEXT,
	PRIMARY KEY (account_id, filename)
);

CREATE VIEW IF NOT EXISTS players AS
SELECT g2.account_id, g2.user_id , g2.name, g.updated_at, g.created_at FROM gifts g2 INNER JOIN (
	SELECT  account_id, user_id, max(date) as updated_at, min(date) as created_at FROM gifts GROUP BY 1, 2
) g ON g.updated_at = g2.date AND g.user_id = g2.user_id AND g.account_id = g2.account_id
ORDER BY created_at desc
`

type StatsDB struct {
	db *sql.DB
}

type StatsTx struct {
	tx *sql.Tx
}

type Stats struct {
	UserID         int    `json:"user_id"`
	UserName       string `json:"user_name"`
	Date           string `json:"date"`
	HuntL1         int    `json:"hunt_l1"`
	HuntL2         int    `json:"hunt_l2"`
	HuntL3         int    `json:"hunt_l3"`
	HuntL4         int    `json:"hunt_l4"`
	HuntL5         int    `json:"hunt_l5"`
	PurchaseL1     int    `json:"purchase_l1"`
	PurchaseL2     int    `json:"purchase_l2"`
	PurchaseL3     int    `json:"purchase_l3"`
	PurchaseL4     int    `json:"purchase_l4"`
	PurchaseL5     int    `json:"purchase_l5"`
	HuntPoints     int    `json:"hunt_points"`
	PurchasePoints int    `json:"purchase_points"`
	TotalPoints    int    `json:"total_points"`
}

type Players struct {
	UserID    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

func NewStats(filename string) (*StatsDB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(dbSchema); err != nil {
		return nil, err
	}
	return &StatsDB{
		db: db,
	}, nil
}

func (c *StatsDB) Close() {
	c.db.Close()
}

func (c *StatsDB) BeginTx() (*StatsTx, error) {
	tx, err := c.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return &StatsTx{
		tx: tx,
	}, nil
}

func (c *StatsDB) IsFileProcessed(accountId int, filename string) (bool, error) {
	res, err := c.db.Query("Select filename from files where account_id = ? and filename = ?", accountId, filename)
	if err != nil {
		return false, err
	}
	defer res.Close()
	return res.Next(), nil
}

func (c *StatsDB) FindGroupByUserDate(accountId int, begin time.Time, end time.Time, userId int) ([]Stats, error) {
	rows, err := c.db.Query(`
	SELECT 
		p.user_id,
		p.name, 
		STRFTIME('%Y/%m/%d', g.date) as date ,
		SUM(g.hunt_l1) AS hunt_l1,
		SUM(g.hunt_l2) AS hunt_l2,
		SUM(g.hunt_l3) AS hunt_l3,
		SUM(g.hunt_l4) AS hunt_l4,
		SUM(g.hunt_l5) AS hunt_l5,
		SUM(g.purchase_l1) AS purchase_l1,
		SUM(g.purchase_l2) AS purchase_l2,
		SUM(g.purchase_l3) AS purchase_l3,
		SUM(g.purchase_l4) AS purchase_l4,
		SUM(g.purchase_l5) AS purchase_l5,
		SUM(g.hunt_points) AS hunt_points ,
		SUM(g.purchase_points) AS purchase_points ,
		SUM(g.hunt_points) + SUM(g.purchase_points)  AS total_points
	FROM gifts g INNER JOIN players p ON p.user_id = g.user_id AND p.account_id = g.account_id
	WHERE g.date BETWEEN ? AND ?
		AND (p.user_id = ? OR 0 = ?)
		AND g.account_id = ?
	GROUP BY 1, 3
	ORDER BY date, total_points DESC
	`, begin.Format(sqlTimeLayout), end.Format(sqlTimeLayout), userId, userId, accountId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []Stats
	for rows.Next() {
		i := Stats{}
		err = rows.Scan(
			&i.UserID,
			&i.UserName,
			&i.Date,
			&i.HuntL1,
			&i.HuntL2,
			&i.HuntL3,
			&i.HuntL4,
			&i.HuntL5,
			&i.PurchaseL1,
			&i.PurchaseL2,
			&i.PurchaseL3,
			&i.PurchaseL4,
			&i.PurchaseL5,
			&i.HuntPoints,
			&i.PurchasePoints,
			&i.TotalPoints)
		if err != nil {
			return nil, err
		}
		data = append(data, i)
	}
	return data, nil

}

func (c *StatsDB) ListPlayers(accountId int, begin time.Time, end time.Time) ([]Players, error) {
	rows, err := c.db.Query(`
	SELECT p.account_id, p.user_id, p.name, p.updated_at, p.created_at
	FROM players p
	WHERE ? BETWEEN p.created_at AND p.updated_at 
	   OR ? BETWEEN p.created_at AND p.updated_at 
	   AND p.account_id = ?
	`, begin.Format(sqlTimeLayout), end.Format(sqlTimeLayout), accountId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []Players
	for rows.Next() {
		i := Players{}
		err = rows.Scan(
			&i.UserID,
			&i.UserName,
			&i.UpdatedAt,
			&i.CreatedAt)
		if err != nil {
			return nil, err
		}
		data = append(data, i)
	}
	return data, nil

}

func (c *StatsTx) Insert(accountId int, row []string, date *string) error {
	var t time.Time
	if date == nil {
		t, _ = time.Parse("2006/02/01 15:04:05", row[25])
	} else {
		t, _ = time.Parse("2006-02-1", *date)
	}

	var _, err = c.tx.Exec(`
	INSERT INTO gifts (
	    account_id,
		user_id, 
		date,
		name, 
		hunt_l1,
		hunt_l2,
		hunt_l3,
		hunt_l4,
		hunt_l5,
		purchase_l1,
		purchase_l2,
		purchase_l3,
		purchase_l4,
		purchase_l5,
		hunt_points,
		purchase_points,
		goal_perc_hunt,
		goal_perc_purchase
	) VALUES(
		?,
		?, 
		?,
		?, 
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?
		);
	`,
		accountId,
		row[0],
		t.Format(sqlTimeLayout),
		row[1],
		row[6],
		row[7],
		row[8],
		row[9],
		row[10],
		row[12],
		row[13],
		row[14],
		row[15],
		row[16],
		row[18],
		row[21],
		row[19],
		row[22])
	return err
}

func (c *StatsTx) AddFileProcessed(accountId int, filename string) error {
	var _, err = c.tx.Exec("INSERT INTO files VALUES(?, ?)", accountId, filename)
	return err
}

func (c *StatsTx) Commit() error {
	return c.tx.Commit()
}

func (c *StatsTx) Rollback() error {
	return c.tx.Rollback()
}
