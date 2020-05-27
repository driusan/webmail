package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type messageBody struct {
	Message     MessageID
	Type        string
	Content     string
	HTMLContent template.HTML
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		w.WriteHeader(403)
		fmt.Fprintf(w, "Permission denied")
		return
	}
	pieces := strings.Split(strings.TrimPrefix(r.URL.Path, "/messages/"), "/")
	if len(pieces) < 1 {
		NotFound(w, r)
		return
	}

	if _, err := os.Stat(fmt.Sprintf("/mail/fs/mbox/%s", pieces[0])); err != nil {
		NotFound(w, r)
		return
	}

	message := MessageID(pieces[0])
	if len(pieces) == 1 {
		if err := templates.ExecuteTemplate(w, "message.html", messageBody{message, "text", message.Body(), ""}); err != nil {
			log.Fatalln(err)
		}
		if r.Method == "POST" {
			r.ParseForm()
			if r.PostForm.Get("unread") == "true" {
				ioutil.WriteFile(
					fmt.Sprintf("/mail/fs/mbox/%s/flags", pieces[0]),
					[]byte("-s"),
					0600,
				)
			}
			setSeen(message, false)
		} else {
			setSeen(message, true)
		}
		return
	}
	switch pieces[1] {
	case "raw":
		f, err := os.Open(
			fmt.Sprintf("/mail/fs/mbox/%s/raw", pieces[0]),
		)
		if err != nil {
			NotFound(w, r)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "text/plain")
		io.Copy(w, f)
		return
	case "html":
		if err := templates.ExecuteTemplate(w, "message.html", messageBody{message, "html", "", template.HTML(message.HTML())}); err != nil {
			log.Fatalln(err)
		}
		setSeen(message, true)
		return
	case "attachments":
		if len(pieces) < 3 {
			NotFound(w, r)
			return
		}
		attachments := message.Attachments()
		var aid MessageID
		for _, a := range attachments {
			if a.AttachID == MessageID(pieces[2]) {
				aid = a.attachId
				break
			}
		}
		if aid == "" {
			NotFound(w, r)
			return
		}
		f, err := os.Open(fmt.Sprintf("/mail/fs/mbox/%s/body", aid))
		if err != nil {
			NotFound(w, r)
			return
		}
		defer f.Close()

		if typ := aid.readfile("type"); typ != "" {
			w.Header().Set("Content-Type", typ)
		}

		if aname := aid.readfile("filename"); aname != "" {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", aname))
		} else {
			w.Header().Set("Content-Disposition", "attachment")
		}

		io.Copy(w, f)
		return
	}
	NotFound(w, r)
	return
}

func setSeen(m MessageID, val bool) {
	old := !m.IsUnread()

	if val {
		ioutil.WriteFile(
			fmt.Sprintf("/mail/fs/mbox/%s/flags", m),
			[]byte("+s"),
			0600,
		)
	} else {
		ioutil.WriteFile(
			fmt.Sprintf("/mail/fs/mbox/%s/flags", m),
			[]byte("-s"),
			0600,
		)
	}

	// Changing the seen flag changes the order
	// displayed.
	if old != val {
		go sortcache()
	}

}
