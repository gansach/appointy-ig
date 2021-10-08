package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
)

var (
	listUserRe   = regexp.MustCompile(`^\/users[\/]*$`)
	getUserRe    = regexp.MustCompile(`^\/users\/(\d+)$`)
	createUserRe = regexp.MustCompile(`^\/users[\/]*$`)

	listPostRe   = regexp.MustCompile(`^\/posts\/users\/(\d+)$`)
	getPostRe    = regexp.MustCompile(`^\/posts\/(\d+)$`)
	createPostRe = regexp.MustCompile(`^\/posts[\/]*$`)
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type post struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Caption string `json:"caption"`
}

type usersDatastore struct {
	m map[string]user
	*sync.RWMutex
}

type postsDatastore struct {
	m map[string]post
	*sync.RWMutex
}

type requestHandler struct {
	usersDB *usersDatastore
	postsDB *postsDatastore
}

func (h *requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodGet && listUserRe.MatchString(r.URL.Path):
		h.ListUsers(w, r)
		return
	case r.Method == http.MethodGet && getUserRe.MatchString(r.URL.Path):
		h.GetUser(w, r)
		return
	case r.Method == http.MethodPost && createUserRe.MatchString(r.URL.Path):
		h.CreateUser(w, r)
		return
	case r.Method == http.MethodGet && listPostRe.MatchString(r.URL.Path):
		h.ListPosts(w, r)
		return
	case r.Method == http.MethodGet && getPostRe.MatchString(r.URL.Path):
		h.GetPost(w, r)
		return
	case r.Method == http.MethodPost && createPostRe.MatchString(r.URL.Path):
		h.CreatePost(w, r)
		return
	default:
		notFound(w, r)
		return
	}
}

func (h *requestHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	h.usersDB.RLock()
	users := make([]user, 0, len(h.usersDB.m))
	for _, v := range h.usersDB.m {
		users = append(users, v)
	}
	h.usersDB.RUnlock()
	jsonBytes, err := json.Marshal(users)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *requestHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	matches := listPostRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}

	h.usersDB.RLock()
	u, ok := h.usersDB.m[matches[1]]
	h.usersDB.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`"error": "User not found"`))
		return
	}

	h.postsDB.RLock()
	posts := make([]post, 0, len(h.postsDB.m))
	for _, p := range h.postsDB.m {
		if p.Author == u.ID {
			posts = append(posts, p)
		}
	}
	h.postsDB.RUnlock()

	jsonBytes, err := json.Marshal(posts)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *requestHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	matches := getUserRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}
	h.usersDB.RLock()
	u, ok := h.usersDB.m[matches[1]]
	h.usersDB.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`"error": "User not found"`))
		return
	}
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *requestHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	matches := getPostRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		notFound(w, r)
		return
	}
	h.postsDB.RLock()
	p, ok := h.postsDB.m[matches[1]]
	h.postsDB.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`"error": "Post not found"`))
		return
	}
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *requestHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u user
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		internalServerError(w, r)
		return
	}
	h.usersDB.Lock()
	h.usersDB.m[u.ID] = u
	h.usersDB.Unlock()
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *requestHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var p post
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		internalServerError(w, r)
		return
	}

	h.postsDB.Lock()
	h.postsDB.m[p.ID] = p
	h.postsDB.Unlock()

	jsonBytes, err := json.Marshal(p)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`"error": "Internal server error"`))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`"error": "Not found"`))
}

func main() {
	mux := http.NewServeMux()
	reqH := &requestHandler{
		usersDB: &usersDatastore{
			m: map[string]user{
				"1": user{ID: "1", Name: "gandharv"},
			},
			RWMutex: &sync.RWMutex{},
		},
		postsDB: &postsDatastore{
			m: map[string]post{
				"1": post{ID: "1", Author: "1", Caption: "First post"},
			},
			RWMutex: &sync.RWMutex{},
		},
	}
	mux.Handle("/users", reqH)
	mux.Handle("/users/", reqH)
	mux.Handle("/posts", reqH)
	mux.Handle("/posts/", reqH)

	http.ListenAndServe("localhost:8080", mux)
}
