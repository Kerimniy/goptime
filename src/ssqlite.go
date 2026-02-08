package main

import (
	"database/sql"
	"errors"
	"log"

	_ "modernc.org/sqlite"
)

type stmt struct {
	q      string
	params []any
}

type Request struct {
	t     int
	stmt  []stmt
	reply chan Response
}

type Response struct {
	data []map[string]any
	err  error
}

var input = make(chan Request, 200)

func sqlite_exec(query string, params ...any) error {
	reply := make(chan Response, 1)

	input <- Request{
		t:     0,
		stmt:  []stmt{{q: query, params: params}},
		reply: reply,
	}

	res := <-reply

	return res.err

}

func sqlite_exec_tx(query []string, params [][]any) error {

	if len(query) != len(params) {
		return errors.New("params length must be equal query length")
	}

	stmt_list := []stmt{}

	for i := range query {
		stmt_list = append(stmt_list, stmt{q: query[i], params: params[i]})
	}

	reply := make(chan Response, 1)

	input <- Request{
		t:     2,
		stmt:  stmt_list,
		reply: reply,
	}

	res := <-reply

	return res.err
}

func sqlite_query(query string, params ...any) ([]map[string]any, error) {
	reply := make(chan Response, 1)

	input <- Request{
		t:     1,
		stmt:  []stmt{{q: query, params: params}},
		reply: reply,
	}

	res := <-reply

	return res.data, res.err

}

func run_SQLite_server() {

	var db *sql.DB

	var err error

	db, err = sql.Open("sqlite", "data/db/main.db?_busy_timeout=5000&_journal_mode=WAL")

	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	for req := range input {

		if req.t == 0 {

			_, err := db.Exec(req.stmt[0].q, req.stmt[0].params...)

			resp := Response{data: nil, err: err}

			req.reply <- resp

			close(req.reply)

		} else if req.t == 2 {

			tx, err := db.Begin()

			if err != nil {
				resp := Response{data: nil, err: err}
				req.reply <- resp
				close(req.reply)
				continue
			}

			bk:=false

			for _, e := range req.stmt {
				_, err = tx.Exec(e.q, e.params...)

				if err != nil {
					if tx.Rollback() != nil {

					}
					resp := Response{data: nil, err: err}
					req.reply <- resp
					close(req.reply)
					bk=true
					break
				}
			}
			if bk{
				continue
			}
			err = tx.Commit()

			resp := Response{data: nil, err: err}

			req.reply <- resp

			close(req.reply)

		} else {

			result := []map[string]any{}

			rows, err := db.Query(req.stmt[0].q, req.stmt[0].params...)

			if err != nil {
				resp := Response{data: nil, err: err}

				req.reply <- resp

				close(req.reply)
				continue
			}
			defer rows.Close()

			cols, c_err := rows.Columns()

			if c_err != nil {
				resp := Response{data: nil, err: c_err}

				req.reply <- resp

				close(req.reply)
				continue
			}

			values := make([]any, len(cols))
			valuePtrs := make([]any, len(cols))

			for i := range values {
				valuePtrs[i] = &values[i]
			}

			for rows.Next() {
				if err = rows.Scan(valuePtrs...); err != nil {
					break
				}

				row := make(map[string]any)
				for i, col := range cols {
					row[col] = values[i]
				}

				result = append(result, row)
			}

			resp := Response{data: result, err: err}

			req.reply <- resp

			close(req.reply)
		}

	}
}
