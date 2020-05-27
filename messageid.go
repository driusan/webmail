package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/driusan/dkim"
)

type MessageID string

func (m MessageID) IsUnread() bool {
	flags := m.readfile("flags")
	return flags[5] != 's'
}

func (m MessageID) From() string {
	ffrom := m.readfile("ffrom")
	if ffrom != "" {
		return ffrom + " <" + m.readfile("from") + ">"
	}
	return m.readfile("from")
}

func (m MessageID) UnixDate() *time.Time {
	datestr := m.readfile("unixdate")
	t, err := time.Parse(time.UnixDate, datestr)
	if err != nil {
		return nil
	}
	return &t
}

func (m MessageID) Date() string {
	datestr := m.readfile("date")
	return datestr
}

func (m MessageID) To() string {
	return m.readfile("to")
}

func (m MessageID) CC() string {
	return m.readfile("cc")
}

func (m MessageID) ReplyTo() string {
	if r := m.readfile("replyto"); r != "" {
		return r
	}
	return m.From()
}

func (m MessageID) Subject() string {
	return m.readfile("subject")
}

func (m MessageID) bodyType(typ string) string {
	t := m.readfile("type")
	switch t {
	case typ:
		return m.readfile("body")

	case "multipart/mixed":
		fallthrough
	case "multipart/alternative":
		for i := 1; ; i++ {
			t := MessageID(
				fmt.Sprintf("%s/%d", m, i),
			).readfile("type")

			if t == "" {
				break
			}
			if t == typ {
				return MessageID(fmt.Sprintf("%s/%d", m, i)).readfile("body")
			}
		}
	}
	return ""
}

func (m MessageID) Body() string {
	return m.bodyType("text/plain")
}
func (m MessageID) HTML() string {
	return m.bodyType("text/html")
}

var dkimcache map[MessageID]string

func (m MessageID) DKIMStatus() string {
	if dkimcache == nil {
		dkimcache = make(map[MessageID]string)
	}
	if s, ok := dkimcache[m]; ok {
		return s
	}
	raw := m.readfile("raw")
	r, err := dkim.FileBuffer(dkim.NormalizeReader(strings.NewReader(raw)))
	if err != nil {
		return "Error checking status"
	}
	defer os.Remove(r.Name())
	if err := dkim.Verify(r); err != nil {
		if s := err.Error(); s == "Permanent failure: no DKIM signature" {
			dkimcache[m] = ""
			return ""
		}
		dkimcache[m] = err.Error()
		return err.Error()
	}
	dkimcache[m] = "Verified"
	return "Verified"
}

type Attachment struct {
	// The ID of the message with the attachment.
	MessageID MessageID

	// The ID of the attachment itself.
	AttachID      MessageID
	Type, Content string
	Filename      string
	// Helper to use readfile on the attachment
	// It's value is "MessageID/Attachid"
	attachId MessageID
}

func (m MessageID) Attachments() []Attachment {
	t := m.readfile("type")
	switch t {
	case "multipart/mixed":
		fallthrough
	case "multipart/alternative":
		var attachments []Attachment
		for i := 1; ; i++ {
			attachid := MessageID(
				fmt.Sprintf("%s/%d", m, i),
			)
			subid := MessageID(
				fmt.Sprintf("%d", i),
			)

			t := attachid.readfile("type")
			if t == "text/plain" || t == "text/html" {
				continue
			}
			if t == "" {
				break
			}
			filename := attachid.readfile("filename")
			if filename == "" {
				filename = string(subid)
			}
			attachments = append(attachments, Attachment{
				MessageID: m,
				AttachID:  subid,
				Type:      t,
				Filename:  filename,
				attachId:  attachid,
			})
		}
		return attachments
	case "text/html", "text/plain":
		return nil
	default:
		// Consider the message itself an attachment if
		// it's not text/plain, text/html, or multipart.
		filename := m.readfile("filename")
		if filename == "" {
			filename = string(m)
		}

		return []Attachment{
			Attachment{
				MessageID: m,
				Type:      t,
				Filename:  filename,
				AttachID:  "1",
				attachId:  m,
			},
		}
	}
}
func (m MessageID) Prev() MessageID {
	mu.RLock()
	defer mu.RUnlock()

	for i, val := range cachedmail {
		if val == m {
			if i == len(cachedmail)-1 {
				return ""
			}
			// cachedmail is in reverse chronological
			// order
			return cachedmail[i+1]
		}
	}
	return ""
}
func (m MessageID) Next() MessageID {
	mu.RLock()
	defer mu.RUnlock()

	for i, val := range cachedmail {
		if val == m {
			if i == 0 {
				return ""
			}
			// cachedmail is in reverse chronological
			// order
			return cachedmail[i-1]
		}
	}
	return ""
}

func (m MessageID) readfile(file string) string {
	content, err := ioutil.ReadFile(fmt.Sprintf("/mail/fs/mbox/%s/%s", m, file))
	if err != nil {
		return ""
	}
	return string(content)
}
