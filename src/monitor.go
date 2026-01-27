package main

import (
	"fmt"
	"math"
	"net/http"
	"time"
)

type Monitor struct {
	Url         string
	Interval    float64
	Timeout     float64
	ServiceName string
	Group       string

	running     bool
	http_client http.Client
}

type CreateMonitorData struct {
	Url      string  `json:"url"`
	Name     string  `json:"name"`
	Group    string  `json:"group"`
	Interval float64 `json:"interval"`
	Timeout  float64 `json:"timeout"`
}

type UpdateMonitorData struct {
	Cname    string  `json:"cname"`
	Url      string  `json:"url"`
	Name     string  `json:"name"`
	Group    string  `json:"group"`
	Interval float64 `json:"interval"`
	Timeout  float64 `json:"timeout"`
}

type Check struct {
	Timestamp int64 `json:"timestamp"`
	Ok        int64 `json:"ok"`
}

type MonitorInfo struct {
	Uptime float32 `json:"uptime"`
	Name   string  `json:"name"`
	Group  string  `json:"group"`

	Checks []Check `json:"checks"`
}

func new_monitor(Url string, Interval float64, Timeout float64, ServiceName string, Group string) Monitor {

	client := http.Client{Timeout: time.Duration(Timeout) * time.Second}

	sqlite_exec("DELETE FROM checks ORDER BY timestamp DESC LIMIT -1 OFFSET 30 ")

	inst := Monitor{Url: Url, Interval: Interval, Timeout: Timeout, ServiceName: ServiceName, Group: Group, http_client: client}

	return inst
}

func (monitor *Monitor) ping() {

	res, err := monitor.http_client.Get(monitor.Url)

	var st = 0

	if err != nil {
		fmt.Println(err, monitor.ServiceName, monitor.Interval)
	}

	if err == nil {
		res.Body.Close()
	}

	if err != nil || res.StatusCode > 499 {

	} else {
		st = 1
	}

	timestamp := time.Now().Unix()

	//tx, s_err := db.Begin()
	/*
		if s_err != nil {
			//log.Fatal(s_err)
		}
	*/
	queries := make([]string, 4)
	params := make([][]any, 4)

	queries = append(queries, "INSERT INTO checks (url,service_name, timestamp, ok) VALUES(?,?,?,?)")
	queries = append(queries, "UPDATE monitors SET all_requests=all_requests+1, success=success+? WHERE service_name=?")
	queries = append(queries, "UPDATE monitors SET all_requests=0, success=0, timestamp=? WHERE ?-timestamp>86400")
	queries = append(queries, "DELETE FROM checks WHERE service_name = ? AND rowid NOT IN (SELECT rowid FROM checks WHERE service_name = ? ORDER BY timestamp DESC LIMIT 30);")

	params = append(params, []any{monitor.Url, monitor.ServiceName, timestamp, st})
	params = append(params, []any{st, monitor.ServiceName})
	params = append(params, []any{timestamp, timestamp})
	params = append(params, []any{monitor.ServiceName, monitor.ServiceName})
	/*
		tx.Exec("INSERT INTO checks (url,service_name, timestamp, ok) VALUES(?,?,?,?)", monitor.Url, monitor.ServiceName, timestamp, st)
		tx.Exec("UPDATE monitors SET all_requests=all_requests+1, success=success+? WHERE service_name=?", st, monitor.ServiceName)
		tx.Exec("UPDATE monitors SET all_requests=0, success=0, timestamp=? WHERE ?-timestamp>86400", timestamp, timestamp)

		_, err = tx.Exec("DELETE FROM checks WHERE service_name = ? AND rowid NOT IN (SELECT rowid FROM checks WHERE service_name = ? ORDER BY timestamp DESC LIMIT 30);", monitor.ServiceName, monitor.ServiceName)
	*/

	err = sqlite_exec_tx(queries, params)

	if err != nil {
		//log.Fatal(err)
	}

	//	tx.Commit()
}

func (monitor *Monitor) getchecks() []Check {

	result := []Check{}

	rows, e := sqlite_query("SELECT timestamp, ok FROM checks WHERE service_name=? ORDER BY timestamp DESC LIMIT 30", monitor.ServiceName)

	if e != nil {
		return result
	}

	for _, row := range rows {
		var check Check

		check.Timestamp = row["timestamp"].(int64)
		check.Ok = row["ok"].(int64)

		//err := rows.Scan(&check.Timestamp, &check.Ok)

		result = append(result, check)

	}

	return result

}

func (monitor *Monitor) getchecksfrom(time int) []Check {

	result := []Check{}

	rows, e := sqlite_query("SELECT timestamp, ok FROM checks WHERE timestamp > ? AND service_name=? ORDER BY timestamp DESC LIMIT 30", time, monitor.ServiceName)

	if e != nil {
		fmt.Println(e)
		return result
	}

	for _, row := range rows {
		var check Check

		check.Timestamp = row["timestamp"].(int64)
		check.Ok = row["ok"].(int64)

		//err := rows.Scan(&check.Timestamp, &check.Ok)

		result = append(result, check)
	}
	return result

}

func (monitor *Monitor) getUptime() float32 {

	rows, e := sqlite_query("SELECT success, all_requests FROM monitors WHERE service_name=?", monitor.ServiceName)

	if e != nil {
		return 0
	}
	var s float64 = 0.0
	var a float64 = 1.0

	for _, row := range rows {

		s = float64(row["success"].(int64))
		a = float64(row["all_requests"].(int64))
	}
	return float32(math.Round(s/a*100) / 100)

}

func (monitor *Monitor) delete() error {
	monitor.stop()
	q := []string{"DELETE FROM monitors WHERE service_name =?", "DELETE FROM checks WHERE service_name =?"}
	p := [][]any{{monitor.ServiceName}, {monitor.ServiceName}}

	return sqlite_exec_tx(q, p)
}

func (monitor *Monitor) run() {
	monitor.running = true

	for monitor.running {

		monitor.ping()

		time.Sleep(time.Duration(monitor.Interval) * time.Second)
	}
}
func (monitor *Monitor) stop() {
	monitor.running = false
}
