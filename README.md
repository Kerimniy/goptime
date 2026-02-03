
---

# Goptime

A lightweight self-hosted uptime monitoring service written in **Go**.
It periodically checks configured URLs, stores results in SQLite, and provides a web UI, JSON API, and SVG uptime badges.

## Features

* 🌐 HTTP/HTTPS uptime monitoring
* ⏱ Configurable check interval and timeout per service
* 📊 Uptime calculation based on recent checks
* 🗄 SQLite database (no external dependencies)
* 👤 Single admin account with authentication
* ✉ Email notifications & password recovery
* 🖥 Web dashboard + admin panel
* 📡 JSON API
* 🏷 Dynamic SVG uptime badges

## Screens & UI

* **Public page** — list of monitored services and their status
* **Admin panel** — manage monitors and server info
* **Login / Registration / Recovery** pages
* **Badge endpoint** — embed uptime badges anywhere

## Requirements

* Go **1.21+** (recommended)
* SQLite (embedded via Go driver)
* Linux / macOS / Windows

## Installation

```bash
git clone https://github.com/yourname/uptime-monitor.git
cd uptime-monitor
go build -o uptime
```

## Running

```bash
./Goptime
```

On first run:

* A random `SECRET_KEY` will be generated in `data/SECRET_KEY`
* The server will bind to `0.0.0.0:80` by default
* SQLite tables will be created automatically
* You will be redirected to **registration** to create the admin account

To change bind address:

```text
data/HOST
```

Example:

```text
127.0.0.1:8080
```

## Project Structure

```text
.
├── main.go
├── data/
│   ├── HOST
│   ├── SECRET_KEY
│   ├── static/
│   │   └── icon
│   └── templates/
│       ├── index.html
│       ├── admin.html
│       ├── login.html
│       ├── reg.html
│       ├── reset_pwd.html
│       └── badge.svg
```

## Authentication Model

* Only **one user** (admin) is supported
* First registered user becomes the admin
* Sessions are stored in **signed HTTP-only cookies**
* Passwords are hashed with **bcrypt**

## Monitors

Each monitor has:

* `url` — target endpoint
* `service_name` — unique name
* `interval` — check interval (seconds)
* `timeout` — request timeout (seconds)
* `group` — logical grouping

Checks:

* Run in separate goroutines
* Last **30 checks** per service are stored
* Uptime is calculated from recent results

## API Endpoints

### Public

| Method | Endpoint                      | Description                   |
| ------ | ----------------------------- | ----------------------------- |
| GET    | `/get-state`                  | Current state of all monitors |
| GET    | `/get_info_from?time=SECONDS` | History since timestamp       |
| GET    | `/api/badge/{id}`             | SVG uptime badge              |
| GET    | `/api/badge?name=SERVICE`     | Badge by service name         |


## Badge Example

```html
<img src="http://your-host/api/badge?name=MyService" />
<!-- or -->
<img src="http://your-host/api/badge/<monitor index e.g 0>" />
```


## Email Support

Used for:

* Password recovery

SMTP credentials are stored in the database and configurable via the UI.

## Security Notes

* Cookies are `HttpOnly` and signed
* Passwords are never stored in plaintext
* CSRF protection is minimal — **not recommended for hostile environments**
* Designed for **self-hosted / private use**

## Limitations

* Single admin user only
* No role system
* No TLS (use a reverse proxy like Nginx / Caddy)
* No rate limiting


