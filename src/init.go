package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Title     string              `json:"title"`
	Md        string              `json:"md"`
	Password  string              `json:"password"`
	Mail_info Mail                `json:"mail"`
	Monitors  []CreateMonitorData `json:"monitors"`
}

func load_conf() {
	if user_exists {
		return
	}
	file, err := os.Open("data/CONF.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err = file.Close(); err != nil {

		}
	}()

	b, err1 := io.ReadAll(file)
	if err1 != nil {
		fmt.Println(err1)
		return
	}

	cfg := Config{}

	err = json.Unmarshal(b, &cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cfg)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcrypt.DefaultCost)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = sqlite_exec("INSERT INTO users (email,password, username,sender , host, mail_pwd) VALUES(?,?,?,?,?,?)", cfg.Mail_info.To, hashedPassword, cfg.Mail_info.Username, cfg.Mail_info.From, cfg.Mail_info.Host, cfg.Mail_info.Password)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = sqlite_exec("UPDATE server SET title=?, md=?", cfg.Title, cfg.Md)
	if err != nil {
		fmt.Println(err)
		return
	}
	title = cfg.Title
	md = cfg.Md

	user_exists = true

	for _, el := range cfg.Monitors {

		client := http.Client{Timeout: time.Duration(el.Timeout) * time.Second}
		monitor := Monitor{Url: el.Url, ServiceName: el.Name, Group: el.Group, Timeout: el.Timeout, Interval: el.Interval, http_client: client}

		err = add_monitor_to_db(&monitor)
		if err != nil {
			fmt.Println(err)
			return
		}

		monitors = append(monitors, monitor)
		go monitor.run()

	}

}
