package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var r_id int = -1
var r_id_mutex = sync.Mutex{}
var code_exp = -1

type Token struct {
	Data string `json:"data"`
	Id   int    `json:"id"`
	Exp  int    `json:"exp"`
	Hmac string `json:"hmac"`
}

func createHMAC(key []byte, data string, exp int, id int) string {

	v := key
	v = append(v, []byte(fmt.Sprintf("%v", exp))...)
	v = append(v, []byte(fmt.Sprintf("%v", id))...)

	h := hmac.New(sha256.New, v)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func createToken(email string) (string, error) {
	r_id_mutex.Lock()
	r_id = rand.Int()
	r_id_mutex.Unlock()
	expires := int(time.Now().Unix()) + 300

	hmac := createHMAC(SECRET_KEY, email, expires, r_id)

	t := Token{Data: email, Id: r_id, Exp: expires, Hmac: hmac}

	_json, err := json.Marshal(t)

	return hex.EncodeToString(_json), err
}

func verifyToken(token_str string) (bool, string) {

	var token Token

	dec, err := hex.DecodeString(token_str)

	if err != nil {
		fmt.Println(err)
		return false, "d"
	}

	err = json.Unmarshal(dec, &token)

	if err != nil {
		fmt.Println(err)
		return false, "u"
	}

	r_id_mutex.Lock()

	if r_id != token.Id {
		return false, "i"
	}

	r_id_mutex.Unlock()

	if time.Now().Unix() > int64(token.Exp) {
		return false, "t"
	}

	expected := createHMAC(SECRET_KEY, token.Data, token.Exp, token.Id)

	r := hmac.Equal([]byte(expected), []byte(token.Hmac))

	if r {
		r_id_mutex.Lock()
		r_id = -1
		r_id_mutex.Unlock()
		return true, token.Data
	}

	return false, "ue"

}

func createCode() string {

	r_id_mutex.Lock()
	r_id = rand.Intn(1000000)
	code_exp = int(time.Now().Unix()) + 600
	r_id_mutex.Unlock()

	return fmt.Sprintf("%06d", r_id)

}

func verifyCode(code string) bool {
	v, err := strconv.Atoi(code)
	if err != nil {
		return false
	}
	r_id_mutex.Lock()
	if time.Now().Unix() > int64(code_exp) {
		return false
	}
	defer r_id_mutex.Unlock()
	if v == r_id {
		r_id = -1
		return true
	}
	return false

}
