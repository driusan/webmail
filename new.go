package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type sendData struct {
	Subject, To, CC, BCC, Body string
}

func NewHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/new" {
		NotFound(w, r)
		return
	}

	if !isAuthenticated(r) {
		PermissionDenied(w, r)
		return
	}

	switch r.Method {
	case "GET":
		var defaultData sendData
		if reply := r.URL.Query().Get("replyto"); reply != "" {
			rid, err := strconv.Atoi(reply)
			if err != nil {
				BadRequest(w, r)
				return
			}
			if _, err := os.Stat(fmt.Sprintf("/mail/fs/mbox/%d", rid)); err != nil {
				BadRequest(w, r)
				return
			}
			mid := MessageID(reply)
			fmt.Printf("MessageID: '%v'\n", mid)
			defaultData.Subject = "Re: " + strings.TrimPrefix(mid.Subject(), "Re: ")
			defaultData.To = mid.ReplyTo()
			defaultData.Body = replyToBody(mid)
		}
		if err := templates.ExecuteTemplate(w, "new.html", defaultData); err != nil {
			log.Fatalln(err)
		}
	default:
		InvalidMethodHandler(w, r, "GET")
	}

}

func replyToBody(m MessageID) string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "On %s <%s> wrote:\n", m.readfile("unixdate"), m.readfile("from"))

	body := strings.NewReader(m.Body())

	linescanner := bufio.NewScanner(body)
	for linescanner.Scan() {
		fmt.Fprintf(w, "> %s\n", linescanner.Text())
	}
	return w.String()
}
