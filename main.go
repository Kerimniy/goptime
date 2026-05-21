package main

import (
	"fmt"
	"io"
	"os"

	"encoding/json"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	_ "github.com/go-git/go-git/v5/plumbing/transport"
	_ "github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"

	_ "github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type State struct {
	Name          string     `json:"name"`
	Group         string     `json:"group"`
	Uptime        uint64     `json:"uptime"`
	Last_changed  uint16     `json:"last_changed"`
	Started       uint64     `json:"started"`
	Interval      uint64     `json:"interval"`
	Checks        [30]uint64 `json:"checks"`
	Reset         uint64     `json:"reset"`
	Real_interval uint64     `json:"real_interval"`
}

const DayUnix = 86400

func main() {

	access_file, err := os.Open("token")

	if err != nil {
		panic(err)
	}

	ba, err := io.ReadAll(access_file)
	access_token := string(ba)

	if err != nil {
		panic(err)
	}
	
	url_file, err := os.Open("url")

	if err != nil {
		panic(err)
	}

	bu, err := io.ReadAll(url_file)
	url := string(bu)

	fmt.Println(url)

	if err != nil {
		panic(err)
	}

	fs := memfs.New()
	storer := memory.NewStorage()

	repo, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: url,
	})

	if err != nil {
		panic(err)
	}

	f, err := fs.Open("data.json")
	if err != nil {
		panic(err)
	}

	b, err := io.ReadAll(f)

	state := State{}
	err = json.Unmarshal(b, &state)
	if err != nil {
		panic(err)
	}

	var index = uint64(state.Last_changed) + 1
	state.Uptime = 0
	last_time := time.Now().Unix()

	if state.Reset == 0 {
		state.Reset = uint64(time.Now().Unix())
	}
	var count = 0
	for {
		time.Sleep(time.Second * time.Duration(state.Interval))

		if uint64(time.Now().Unix())-state.Reset >= DayUnix {
			state.Reset = uint64(time.Now().Unix())

			state.Uptime = 0
		}
		count++
		state.Checks[index] = uint64(time.Now().Unix())
		state.Last_changed = uint16(index)
		state.Uptime += (uint64(time.Now().Unix() - last_time))

		state.Real_interval = state.Uptime / uint64(count)
		last_time = time.Now().Unix()

		push_state(&state, repo, fs, access_token)
		index += 1
		if index == 30 {
			index = 0
		}

	}

}

func push_state(state *State, repo *git.Repository, fs billy.Filesystem, access_token string) {

	w, err := repo.Worktree()

	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(state)

	if err != nil {
		panic(err)
	}

	f, _ := fs.Create("data.json")
	f.Write(b)
	f.Close()

	_, err = w.Add("data.json")
	if err != nil {
		panic(err)
	}

	_, err = w.Commit("commit3", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Goptime-s",
			Email: "test@example.com",
			When:  time.Now().AddDate(2025, 01, 14),
		},
	})

	if err != nil {
		panic(err)
	}

	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "git",
			Password: access_token,
		},
	})

	if err != nil {
		panic(err)
	}
}
