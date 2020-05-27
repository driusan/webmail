package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

var mu sync.RWMutex
var cachedmail []MessageID

type indexdata struct {
	NumMessages int
	Messages    []MessageID

	StartID, EndID     int
	NextPage, PrevPage int

	Banner string
}

const messagesPerPage = 50

func isAuthenticated(r *http.Request) bool {
	authtoken, err := r.Cookie("token")
	if err != nil {
		return false
	}
	_, ok := logins[loginCookie(authtoken.Value)]
	return ok
}

func getSessionInfo(r *http.Request) sessionInfo {
	authtoken, err := r.Cookie("token")
	if err != nil {
		return sessionInfo{}
	}
	i, ok := logins[loginCookie(authtoken.Value)]
	if ok {
		return i
	}
	return sessionInfo{}
}

func setBanner(r *http.Request, v string) {
	authtoken, err := r.Cookie("token")
	if err != nil {
		return
	}

	av := loginCookie(authtoken.Value)
	i, ok := logins[av]
	if !ok {
		return
	}
	i.banner = v
	logins[av] = i

}
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		w.WriteHeader(404)
		return
	}

	if !isAuthenticated(r) {
		w.Header().Set("Location", "/login")
		w.WriteHeader(303)
	}
	loadcacheifmodified()
	mu.RLock()
	defer mu.RUnlock()

	var messages []MessageID
	startMessage := 0
	endMessage := 0
	nextPage := 0
	prevPage := 0
	page := 1
	queryparams := r.URL.Query()
	if pnum := queryparams.Get("page"); pnum != "" {
		p, err := strconv.Atoi(pnum)
		if err != nil {
			log.Fatalln(err)
		}
		page = p

	}

	startMessage = (page - 1) * messagesPerPage
	if startMessage+messagesPerPage < len(cachedmail) {
		nextPage = page + 1
	}
	if startMessage > 0 {
		prevPage = page - 1
	}

	if startMessage > len(cachedmail) {
		messages = nil
		startMessage = 0
		endMessage = 0
	} else if len(cachedmail) > startMessage+messagesPerPage {
		messages = cachedmail[startMessage : startMessage+messagesPerPage]
		endMessage = startMessage + messagesPerPage
	} else {
		messages = cachedmail[startMessage:]
		endMessage = len(cachedmail)
	}

	templateData := indexdata{
		NumMessages: len(cachedmail),
		Messages:    messages,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		StartID:     startMessage + 1,
		EndID:       endMessage,
	}
	si := getSessionInfo(r)
	if si.banner != "" {
		templateData.Banner = si.banner
		setBanner(r, "")
	}
	if err := templates.ExecuteTemplate(w, "index.html", templateData); err != nil {
		log.Fatalln(err)
	}
}

func sortcache() {
	newcache := cachedmail
	mu.RLock()
	sort.Slice(newcache, func(i, j int) bool {
		// Unread come first
		iunread := newcache[i].IsUnread()
		junread := newcache[j].IsUnread()
		if iunread && !junread {
			return true
		} else if junread && !iunread {
			return false
		}

		// Then sort by date
		idate := newcache[i].UnixDate()
		jdate := newcache[j].UnixDate()
		if idate != nil && jdate != nil {
			return idate.Unix() > jdate.Unix()
		}

		// If there's no date, use the message id
		return newcache[i] < newcache[j]
	})
	mu.RUnlock()
	mu.Lock()
	cachedmail = newcache
	mu.Unlock()
}

var lastmTime time.Time

func loadcacheifmodified() {
	mtime, err := os.Stat("/mail/fs/mbox")
	if err != nil {
		return
	}
	mt := mtime.ModTime()
	// FIXME: Check if this is a valid assumption. The
	// meaning of mtime on a directory is undefined.
	if mt.Unix() > lastmTime.Unix() {
		loadcache()
	} else {
	}
}

func loadcache() {
	preloadmail, err := ioutil.ReadDir("/mail/fs/mbox")
	if err != nil {
		log.Fatalln(err)
	}

	// -1 for ctl
	newcachedmail := make([]MessageID, 0, len(preloadmail)-1)
	for _, file := range preloadmail {
		if file.Name() == "ctl" {
			continue
		}
		newcachedmail = append(newcachedmail, MessageID(file.Name()))
	}

	// Sort re-acquires the lock, so release it before calling to
	// prevent deadlock
	mu.Lock()
	cachedmail = newcachedmail
	mu.Unlock()
	sortcache()

	mtime, err := os.Stat("/mail/fs/mbox")
	if err != nil {
		return
	}
	lastmTime = mtime.ModTime()
}
