package main

import (
	"fmt"
	"log"
	"net/http"
	osuser "os/user"

	"crypto/rand"

	"github.com/driusan/p9auth"
)

type loginCookie string
type user string

var logins map[loginCookie]user = make(map[loginCookie]user)

type templateArgs struct {
	Username string
	Error    string
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		w.WriteHeader(404)
		return
	}
	switch r.Method {
	case "GET":
		if err := templates.ExecuteTemplate(w, "login.html", templateArgs{}); err != nil {
			log.Fatalln(err)
		}
	case "POST":
		uname := r.FormValue("username")
		pass := r.FormValue("password")
		osuname, err := osuser.Current()
		if err != nil {
			panic(err)
		}

		_, err = p9auth.Userpasswd(uname, pass)
		// Validate that the user is the same user whose
		// mailbox is mounted at /mail/fs/mbox and that
		// their password is correct
		if err != nil || uname != osuname.Username {
			w.WriteHeader(403)
			if err := templates.ExecuteTemplate(w, "login.html", templateArgs{
				Error:    "Bad username/password",
				Username: uname,
			}); err != nil {
				log.Fatalln(err)
			}
			return
		}
		var buf [128]byte
		if _, err := rand.Read(buf[:]); err != nil {
			w.WriteHeader(500)
			return
		}

		cookie := loginCookie(fmt.Sprintf("%x", buf[:]))

		logins[cookie] = user(uname)
		w.Header().Set("Location", "/")
		w.Header().Set("Set-Cookie", "token="+string(cookie))
		w.WriteHeader(303)
		fmt.Fprintf(w, "Your browser should have redirected you to /")
	default:
		InvalidMethodHandler(w, r, "GET, POST")
		return
	}
	return
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Set-Cookie", "")
	w.Header().Set("Location", "/")
	cookie, err := r.Cookie("token")
	if err == nil {
		delete(logins, loginCookie(cookie.Value))
	}
	w.WriteHeader(303)
}
