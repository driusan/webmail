package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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
			defaultData.Subject = "Re: " + strings.TrimPrefix(mid.Subject(), "Re: ")
			defaultData.To = mid.ReplyTo()
			defaultData.Body = replyToBody(mid)
			defaultData.CC = mid.CC()
		}
		if err := templates.ExecuteTemplate(w, "new.html", defaultData); err != nil {
			log.Fatalln(err)
		}
		return
	case "POST":
		r.ParseForm()

		sb := &strings.Builder{}
		if to := r.PostForm.Get("to"); to != "" {
			fmt.Fprintf(sb, "To: %v\n", to)
		} else {
			BadRequest(w, r)
			return
		}
		if s := r.PostForm.Get("subject"); s != "" {
			fmt.Fprintf(sb, "Subject: %v\n", s)
		}
		if cc := r.PostForm.Get("cc"); cc != "" {
			fmt.Fprintf(sb, "Cc: %v\n", cc)
		}
		if bcc := r.PostForm.Get("bcc"); bcc != "" {
			fmt.Fprintf(sb, "Bcc: %v\n", bcc)
		}
		if body := r.PostForm.Get("body"); body != "" {
			fmt.Fprintf(sb, "\n%v\n", body)
		}

		// Write to /bin/upas/marshal -8 -R /mail/fs/mbox/$replyto
		args := []string{"-8"}
		if reply := r.URL.Query().Get("replyto"); reply != "" {
			args = append(args, "-R", "/mail/fs/mbox/"+reply)
		}

		cmd := exec.Command("/bin/upas/marshal", args...)
		cmd.Stdin = strings.NewReader(sb.String())
		if err := cmd.Run(); err != nil {
			log.Println(err)
		}
		w.Header().Set("Location", "/")

		setBanner(r, "Message Successfully sent")
		w.WriteHeader(303)
		return
	default:
		InvalidMethodHandler(w, r, "GET, POST")
	}

}

func replyToBody(m MessageID) string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "On %s <%s> wrote:\n", m.Date(), m.readfile("from"))

	body := strings.NewReader(m.Body())

	linescanner := bufio.NewScanner(body)
	for linescanner.Scan() {
		fmt.Fprintf(w, "> %s\n", linescanner.Text())
	}
	return w.String()
}
