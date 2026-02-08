package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

var title string = ""
var md string = ""

var monitors []Monitor

type AdminCnt struct {
	Monitors   string
	ServerInfo string
}

type IndexCnt struct {
	Monitors   string
	ServerInfo string
	User       bool
}

type BadgeCnt struct {
	Uptime string
	Color  string
}

type CreateAccountInfo struct {
	Password     string `json:"password"`
	New_password string `json:"new_password"`
	Mail_info    Mail   `json:"mail_info"`
}

type LoginInfo struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RecoveryInfo struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}

var user_exists = false
var user_exists_mutex = sync.Mutex{}

var SECRET_KEY = make([]byte, 64)
var s = securecookie.New(SECRET_KEY, nil)

var index_tmpl = template.Must(template.ParseFiles("data/templates/index.html"))
var login_tmpl = template.Must(template.ParseFiles("data/templates/login.html"))
var reg_tmpl = template.Must(template.ParseFiles("data/templates/reg.html"))
var reset_pwd_tmpl = template.Must(template.ParseFiles("data/templates/reset_pwd.html"))
var admin_tmpl = template.Must(template.ParseFiles("data/templates/admin.html"))
var badge_tmpl = template.Must(template.ParseFiles("data/templates/badge.svg"))

func load_monitors() {
	rows, e := sqlite_query("SELECT * FROM monitors")

	if e != nil {
		log.Fatal(e)
	}

	for _, row := range rows {
		var monitor = Monitor{}

		//rows.Scan(&monitor.Url, &monitor.ServiceName, &monitor.Interval, &monitor.Timeout, &int_dummy, &int_dummy, &int_dummy, &monitor.Group)

		monitor.Url = row["url"].(string)
		monitor.ServiceName = row["service_name"].(string)
		monitor.Interval = row["interval"].(float64)
		monitor.Timeout = row["timeout"].(float64)
		monitor.Group = row["mgroup"].(string)

		client := http.Client{Timeout: time.Duration(monitor.Timeout) * time.Second}
		sqlite_exec("DELETE FROM checks WHERE service_name = ? AND rowid NOT IN (SELECT rowid FROM checks WHERE service_name = ? ORDER BY timestamp DESC LIMIT 30);", monitor.ServiceName, monitor.ServiceName)

		monitor.http_client = client

		go monitor.run()

		monitors = append(monitors, monitor)
	}

}

func setSignedCookie(w http.ResponseWriter) {
	encoded, _ := s.Encode("session", "authorized")
	cookie := &http.Cookie{Name: "session", Value: encoded, HttpOnly: true, Path: "/", SameSite: http.SameSiteLaxMode, Expires: time.Unix(time.Now().Unix()+34560000, 0)}
	http.SetCookie(w, cookie)
}

func deleteCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

}

func getSignedCookie(r *http.Request, w http.ResponseWriter) string {
	if cookie, err := r.Cookie("session"); err == nil {

		var decoded string

		if err = s.Decode("session", cookie.Value, &decoded); err == nil {

			return decoded
		}
		deleteCookie(w)
	}
	return "none"
}

func is_logined(r *http.Request, w http.ResponseWriter) bool {
	return getSignedCookie(r, w) == "authorized"
}

func init_db() {
	var count int64 = 0

	var eerr error

	eerr = sqlite_exec("PRAGMA synchronous = NORMAL;")

	if eerr != nil {
		log.Fatal(eerr)
	}

	eerr = sqlite_exec("CREATE TABLE IF NOT EXISTS users (email TEXT NOT NULL UNIQUE,password TEXT NOT NULL, username TEXT, sender TEXT,host TEXT, mail_pwd TEXT,PRIMARY KEY (email))")

	if eerr != nil {
		log.Fatal(eerr)
	}

	eerr = sqlite_exec("CREATE TABLE IF NOT EXISTS server (title TEXT NOT NULL UNIQUE DEFAULT 'Uptime', md TEXT)")
	if eerr != nil {
		log.Fatal(eerr)
	}
	eerr = sqlite_exec("CREATE TABLE IF NOT EXISTS checks (url TEXT NOT NULL, service_name TEXT, timestamp INTEGER, ok INTEGER)")
	if eerr != nil {
		log.Fatal(eerr)
	}
	eerr = sqlite_exec("CREATE TABLE IF NOT EXISTS monitors ( url TEXT NOT NULL, service_name TEXT, interval REAL, timeout REAL, success INTEGER, all_requests INTEGER, timestamp TEXT NOT NULL, mgroup TEXT NOT NULL, PRIMARY KEY (url,service_name) )")
	if eerr != nil {
		log.Fatal(eerr)
	}
	rows, e := sqlite_query("SELECT Count(*) FROM server")

	if e != nil {
		log.Fatal(e)
	}
	count = 0
	for _, e := range rows {
		count = e["Count(*)"].(int64)
		break
	}

	if count < 1 {
		e = sqlite_exec(`INSERT INTO server (title,md) VALUES ("Uptime","")`)
		if e != nil {
			log.Fatal(e)
		}
	}

	q, err := sqlite_query("SELECT COUNT(*) FROM users")
	count = 0
	if err != nil {
		log.Fatal(err)
	}

	for _, el := range q {
		t := el["COUNT(*)"].(int64)

		count = t
	}

	if count > 0 {
		user_exists = true
	}

}

func main() {

	file, f_err := os.Open("data/SECRET_KEY")
	if f_err != nil {
		_, e := rand.Read(SECRET_KEY)
		f, err := os.Create("data/SECRET_KEY")
		_, e1 := f.Write(SECRET_KEY)
		if e != nil || err != nil || e1 != nil {
			log.Fatal(e)
		}

	} else {
		_, err2 := file.Read(SECRET_KEY)
		if err2 != nil {
			_, e := rand.Read(SECRET_KEY)
			f, err := os.Create("data/SECRET_KEY")
			_, e1 := f.Write(SECRET_KEY)
			if e != nil || err != nil || e1 != nil {
				log.Fatal(e)
			}

		}
	}

	var host = read_file_as_str("data/HOST")

	if host == "" {
		host = "0.0.0.0:80"
		f, err := os.Create("data/HOST")
		_, e1 := f.Write([]byte(`0.0.0.0:80`))
		if err != nil || e1 != nil {
			log.Fatal(err, e1)
		}
	}

	s = securecookie.New(SECRET_KEY, nil)

	file.Close()

	go run_SQLite_server()

	init_db()

	http.HandleFunc("/", index)
	http.HandleFunc("/get-state", get_monitors_state)
	http.HandleFunc("/get_info_from", get_info_from)
	http.HandleFunc("/create-monitor/", create_monitor)
	http.HandleFunc("/update-monitor/", update_monitor)
	http.HandleFunc("/delete-monitor/", delete_monitor)
	http.HandleFunc("/update-server/", update_server_info)
	http.HandleFunc("/admin/", render_admin)
	http.HandleFunc("/reg/", render_reg)
	http.HandleFunc("/login/", render_login)
	http.HandleFunc("/logout/", logout)
	http.HandleFunc("/recovery/", render_recovery)
	http.HandleFunc("/api/recovery/", reset_password)
	http.HandleFunc("/api/reg/", create_account)
	http.HandleFunc("/api/login/", login)
	http.HandleFunc("/api/badge/{id}/", get_badge)
	http.HandleFunc("/api/badge/", get_badge)
	http.HandleFunc("/send-test/", send_test_email)

	fs := http.FileServer(http.Dir("./data"))
	http.Handle("/static/", fs)

	load_monitors()
	load_mail()
	load_conf()

	fmt.Printf("\n\nRunning at %s\n\n", host)
	log.Fatal(http.ListenAndServe(host, nil))
}

func index(w http.ResponseWriter, r *http.Request) {

	if user_exists == false {
		http.Redirect(w, r, "/reg/", 303)
	}

	var monitors string
	dt, err := json.Marshal(get_all_info())

	if err != nil {
		monitors = "{}"
	} else {
		monitors = string(dt)
	}

	cnt := IndexCnt{Monitors: monitors, ServerInfo: get_server_info(), User: is_logined(r, w)}

	index_tmpl.Execute(w, cnt)
}

func get_badge(w http.ResponseWriter, r *http.Request) {

	i_s := r.PathValue("id")
	i, err := strconv.Atoi(i_s)
	var uptime float32 = -1.0
	if err != nil {
		service_name := r.URL.Query().Get("name")

		for _, el := range monitors {
			if el.ServiceName == service_name {

				uptime = el.getUptime()
				break
			}
		}
	} else if i < len(monitors) {
		uptime = monitors[i].getUptime()
	}

	cnt := BadgeCnt{Uptime: fmt.Sprintf("%v", func() string {
		if uptime >= 0 {
			return fmt.Sprintf("%v", math.Round(float64(uptime)*1000)/10)
		} else {
			return "N/A"
		}
	}()), Color: p2rgb(uptime)}

	w.Header().Add("Content-Type", "image/svg+xml")

	badge_tmpl.Execute(w, cnt)

}

func render_login(w http.ResponseWriter, r *http.Request) {

	login_tmpl.Execute(w, nil)

}

func render_reg(w http.ResponseWriter, r *http.Request) {

	is_logined := is_logined(r, w)

	if user_exists && !is_logined {

		http.Redirect(w, r, "/login", 303)

	} else if is_logined {

		cnt := Mail{}

		rows, err := sqlite_query("SELECT email,username,sender,host,mail_pwd FROM users LIMIT 1")

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err)
			return
		}

		for _, row := range rows {
			cnt.From = row["sender"].(string)
			cnt.To = row["email"].(string)
			cnt.Username = row["username"].(string)
			cnt.Host = row["host"].(string)
			cnt.Password = row["mail_pwd"].(string)
		}

		v, e := json.Marshal(cnt)

		if e != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err)
			return
		}

		reg_tmpl.Execute(w, string(v))
		return
	} else {
		reg_tmpl.Execute(w, "{}")
	}
}

func render_recovery(w http.ResponseWriter, r *http.Request) {

	err := send_recovery(mail_info, createCode())
	if err != nil {
		fmt.Println(err)
		r_id_mutex.Lock()
		r_id = -1
		r_id_mutex.Unlock()
		w.WriteHeader(424)
		return
	}

	reset_pwd_tmpl.Execute(w, nil)

}

func render_admin(w http.ResponseWriter, r *http.Request) {

	if !is_logined(r, w) {
		http.Redirect(w, r, "/login", 303)
	}

	cnt := AdminCnt{Monitors: get_monitors_info(), ServerInfo: get_server_info()}

	admin_tmpl.Execute(w, cnt)

}

func create_account(w http.ResponseWriter, r *http.Request) {

	var data CreateAccountInfo
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(400)
		fmt.Println(err)
		fmt.Fprint(w, err)
		return
	}

	if !user_exists {

		setSignedCookie(w)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err)
			return
		}

		err = sqlite_exec("INSERT INTO users (email,password, username,sender , host, mail_pwd) VALUES(?,?,?,?,?,?)", data.Mail_info.To, hashedPassword, data.Mail_info.Username, data.Mail_info.From, data.Mail_info.Host, data.Mail_info.Password)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err)
			return
		}

		user_exists_mutex.Lock()
		user_exists = true
		user_exists_mutex.Unlock()
	} else {
		if is_logined(r, w) && verify_password(data.Mail_info.To, data.Password) {

			var hashedPassword []byte
			if data.New_password != "" {
				var err error
				hashedPassword, err = bcrypt.GenerateFromPassword([]byte(data.New_password), bcrypt.DefaultCost)
				if err != nil {
					w.WriteHeader(500)
					fmt.Fprint(w, err)
					return
				}
			}

			err = sqlite_exec("UPDATE users SET email=COALESCE(NULLIF(?, ''), email) ,password= COALESCE(NULLIF(?, ''), password), username= COALESCE(NULLIF(?, ''), username),sender= COALESCE(NULLIF(?, ''), sender) , host= COALESCE(NULLIF(?, ''), host), mail_pwd= COALESCE(NULLIF(?, ''), mail_pwd)", data.Mail_info.To, hashedPassword, data.Mail_info.Username, data.Mail_info.From, data.Mail_info.Host, data.Mail_info.Password)

			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, err)
				return
			}
		} else {
			w.WriteHeader(401)
			fmt.Fprint(w, "Unauthorized")
		}
	}

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		fmt.Fprint(w, err)
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) {

	var data LoginInfo
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(400)
		fmt.Println(err)
		fmt.Fprint(w, err)
		return
	}

	rows, err := sqlite_query("SELECT password FROM users WHERE email=?", data.Email)

	if err != nil {
		w.WriteHeader(400)
		fmt.Println(err)
		fmt.Fprint(w, err)
		return
	}

	var hashedPassword []uint8
	for _, row := range rows {
		hashedPassword = row["password"].([]uint8)
	}

	failure := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(data.Password))

	if failure != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, "incorrect login or password")
		return
	}

	setSignedCookie(w)
	w.WriteHeader(200)

}

func logout(w http.ResponseWriter, r *http.Request) {
	deleteCookie(w)
	w.WriteHeader(205)
}

func reset_password(w http.ResponseWriter, r *http.Request) {
	var data RecoveryInfo
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, "Invalid data")
		return
	}
	if verifyCode(data.Token) {

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "server error")
			return
		}

		err = sqlite_exec("UPDATE users SET password=?", hashedPassword)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "db server error")
			return
		}

		w.WriteHeader(205)
		return

	} else {
		w.WriteHeader(400)
		fmt.Fprint(w, "Invalid or expired code")
		return
	}

}

func send_test_email(w http.ResponseWriter, r *http.Request) {

	var data Mail
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}

	err = send_test(data)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(205)
	fmt.Fprint(w, "success")

}

func create_monitor(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	if !is_logined(r, w) {
		http.Redirect(w, r, "/login", 303)
	}

	var data CreateMonitorData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, "Invalid data")

		return
	}
	if data.Interval < 1 {
		w.WriteHeader(400)
		fmt.Fprint(w, "Interval too short")
		return
	}

	client := http.Client{Timeout: time.Duration(data.Timeout) * time.Second}
	sqlite_exec("DELETE FROM checks WHERE service_name=? ORDER BY timestamp DESC LIMIT -1 OFFSET 30", data.Name)
	monitor := Monitor{Url: data.Url, ServiceName: data.Name, Group: data.Group, Timeout: data.Timeout, Interval: data.Interval, http_client: client}

	err = add_monitor_to_db(&monitor)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	monitors = append(monitors, monitor)
	go monitor.run()

	w.WriteHeader(202)
}

func update_server_info(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	if !is_logined(r, w) {
		http.Redirect(w, r, "/login", 303)
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("image")

	if err != nil && err != http.ErrMissingFile {
		w.WriteHeader(500)
		fmt.Fprint(w, "file error")
		return
	} else if err != http.ErrMissingFile {

		localfile, err := os.Create("data/static/icon")

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "file creation")

			return
		}
		_, err = io.Copy(localfile, file)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "file write")

			return
		}

		defer localfile.Close()
		defer file.Close()
	}

	r_md := r.FormValue("md")
	r_title := r.FormValue("title")

	err = sqlite_exec("UPDATE server SET md=?, title=?", r_md, r_title)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "info update")

		return
	}

	md = r_md
	title = r_title

}

func delete_monitor(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	if !is_logined(r, w) {
		http.Redirect(w, r, "/login", 303)
	}

	buff, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, "parse error")
		return
	}

	cname := string(buff)

	for i, e := range monitors {
		if e.ServiceName == cname {
			err = e.delete()
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, "deletion failed")
				return
			}
			monitors = remove(monitors, i)
		}
	}

	w.WriteHeader(205)
}

func update_monitor(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	if !is_logined(r, w) {
		http.Redirect(w, r, "/login", 303)
	}

	var data UpdateMonitorData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "err: %s", err)
		return
	}
	if data.Interval < 1 {
		w.WriteHeader(400)
		fmt.Fprint(w, "to short interval")
		return
	}

	client := http.Client{Timeout: time.Duration(data.Timeout) * time.Second}

	//tx, o_err := db.Begin()
	/*
		if o_err != nil {
			w.WriteHeader(500)
			return
		}
	*/
	queries := make([]string, 2)
	params := make([][]any, 2)

	queries = append(queries, "UPDATE checks SET url=?, service_name=? WHERE service_name=?")
	queries = append(queries, "UPDATE monitors SET url=?, service_name=?, interval=?, timeout=?, mgroup=? WHERE service_name=?")

	params = append(params, []any{data.Url, data.Name, data.Cname})
	params = append(params, []any{data.Url, data.Name, data.Interval, data.Timeout, data.Group, data.Cname})

	//tx.Exec("UPDATE checks WHERE SET url=?, service_name=? WHERE service_name=?", data.Url, data.Name, data.Cname)
	//tx.Exec("UPDATE monitors WHERE SET url=?, service_name=?, interval=?, timeout=?, mgroup=? WHERE service_name=?", data.Url, data.Name, data.Interval, data.Interval, data.Group, data.Cname)

	err = sqlite_exec_tx(queries, params)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "err: %s", err)
		return
	}

	for i, e := range monitors {

		if e.ServiceName == data.Cname {

			monitors[i].stop()

			monitors[i].http_client = client
			monitors[i].Group = data.Group
			monitors[i].Url = data.Url
			monitors[i].ServiceName = data.Name
			monitors[i].Interval = data.Interval
			monitors[i].Timeout = data.Timeout

			go monitors[i].run()
			break
		}

	}

	w.WriteHeader(202)
}

func get_monitors_info() string {

	data := []CreateMonitorData{}

	for _, e := range monitors {
		mi := CreateMonitorData{}

		mi.Name = e.ServiceName
		mi.Group = e.Group
		mi.Interval = e.Interval
		mi.Timeout = e.Timeout
		mi.Url = e.Url

		data = append(data, mi)
	}

	r, e := json.Marshal(data)

	if e != nil {
		return "[]"
	}
	return string(r)
}

func get_monitors_state(w http.ResponseWriter, r *http.Request) {

	//w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(get_all_info())

}

func get_info_from(w http.ResponseWriter, r *http.Request) {

	time, err := strconv.Atoi(r.URL.Query().Get("time"))

	if err != nil {
		w.WriteHeader(400)
		return
	}

	res := []MonitorInfo{}
	for _, e := range monitors {

		mi := MonitorInfo{Name: e.ServiceName, Group: e.Group}

		mi.Checks = e.getchecksfrom(time)
		mi.Uptime = e.getUptime()

		res = append(res, mi)
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	json.NewEncoder(w).Encode(res)
}

func get_all_info() []MonitorInfo {

	res := []MonitorInfo{}
	for _, e := range monitors {

		mi := MonitorInfo{Name: e.ServiceName, Group: e.Group}

		mi.Checks = e.getchecks()
		mi.Uptime = e.getUptime()

		res = append(res, mi)
	}

	return res
}

func add_monitor_to_db(monitor *Monitor) error {

	err := sqlite_exec("INSERT INTO monitors (url, service_name, interval, timeout,timestamp,mgroup, success, all_requests) VALUES (?,?,?,?,?,?,0,0);", monitor.Url, monitor.ServiceName, monitor.Interval, monitor.Timeout, time.Now().Unix(), monitor.Group)
	return err
}

func get_server_info() string {

	if title == "" && md == "" {
		rows, err := sqlite_query("SELECT title, md FROM server")

		if err != nil {
			fmt.Println(err)
			return `{"title":"Uptime","md":""}`
		}

		for _, row := range rows {

			buff, e := json.Marshal(row)

			if e != nil {
				fmt.Println(e)
				title = "Uptime"
				md = ""
				return `{"title":"Uptime","md":""}`
			} else {
				title = row["title"].(string)
				md = row["md"].(string)
				return string(buff)
			}

		}
	}

	return fmt.Sprintf(`{"title":"%s","md":"%s"}`, title, md)
}

func verify_password(email string, pwd string) bool {
	rows, err := sqlite_query("SELECT password FROM users WHERE email=?", email)
	if err != nil {
		return false
	}

	var hashedPassword []uint8
	for _, row := range rows {
		hashedPassword = row["password"].([]uint8)
	}

	failure := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(pwd))

	if failure != nil {
		return false
	}
	return true
}

func remove(slice []Monitor, s int) []Monitor {
	return append(slice[:s], slice[s+1:]...)
}

func read_file_as_str(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() {
		if err = file.Close(); err != nil {

		}
	}()

	b, err1 := io.ReadAll(file)

	if err1 != nil {
		return ""
	}

	return string(b)
}

func p2rgb(v float32) string {
	var g float32
	var r float32

	if v == -1.0 {
		return `rgb(128,128,128)`
	}

	v -= 0.25
	if v < 0 {
		v = 0
	}

	if v < 0.5 {

		r = 255
		g = (v * 2) * 245
	} else {
		r = (1 - (v-0.5)*2) * 250
		if r < 0 {
			r = 0
		}
		g = 245
	}
	return fmt.Sprintf(`rgb(%v,%v,0)`, r*0.9, g*0.9)
}
