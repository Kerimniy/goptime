<div align="center" width="100%">
    <img src="./frontend/static/icon" width="128" alt=" Logo" />
</div>

---

>![IMPORTANT]
>You have to have 2 repo:
> with data.json
> with frontend (github pages)

# Goptime-s

A lightweight serverless uptime monitoring service written in **Go**.
It periodically push repo with data.json, provides a web UI and SVG uptime badges.

## Features

* 🌐 Server status monitoring
* ⏱ Configurable check interval
* 📊 Uptime calculation based on recent checks
* 🖥 Web dashboard
* 🏷 Dynamic SVG uptime badges

## Screens & UI

* **Public page** — list of monitored services and their status
* **Badge endpoint** — embed uptime badges anywhere via script (badge.js)

## Requirements
* github access token
* Go **1.21+** (recommended)
* Linux / Windows


## data.json
* you have to adjust interval
* you have to add api url of data.json (e.g. `https://api.github.com/repos/Kerimniy-Uptime/test/contents/data.json`) to url list in index.html

On first run:

* create `token` file in workdir with your token
* crate `url` file in workdir with git repo's url



## Badge Example

```js
<div data-badge-url="https://api.github.com/repos/Kerimniy-Uptime/test/contents/data.json"></div>
```


## Security Notes

* Cookies are `HttpOnly` and signed
* Passwords are never stored in plaintext
* CSRF protection is minimal — **not recommended for hostile environments**
* Designed for **self-hosted / private use**

## Limitations

* No role system
* No service info
* No https/http

