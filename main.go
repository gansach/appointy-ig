package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RegEx for Routing
var (
	getUserRe    = regexp.MustCompile(`^\/users\/(\d+)$`)
	createUserRe = regexp.MustCompile(`^\/users[\/]*$`)

	listPostRe     = regexp.MustCompile(`^\/posts\/users\/(\d+)`)
	listPostPageRe = regexp.MustCompile(`^\/posts\/users\/\d+\?page=(\d+)`)
	getPostRe      = regexp.MustCompile(`^\/posts\/(\d+)$`)
	createPostRe   = regexp.MustCompile(`^\/posts[\/]*$`)
)

// Models
type user struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type post struct {
	ID       string `json:"id"`
	Author   string `json:"author"`
	Caption  string `json:"caption"`
	ImageURL string `json:"image"`
	Posted   string `json:"time"`
}

var client *mongo.Client

type requestHandler struct{}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	mux := http.NewServeMux()
	reqH := &requestHandler{}

	mux.Handle("/users", reqH)
	mux.Handle("/users/", reqH)
	mux.Handle("/posts", reqH)
	mux.Handle("/posts/", reqH)

	http.ListenAndServe(":8080", mux)
}

// Routing using RegEx
func (h *requestHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	switch {
	case request.Method == http.MethodGet && getUserRe.MatchString(request.URL.Path):
		h.GetUser(response, request)
		return
	case request.Method == http.MethodPost && createUserRe.MatchString(request.URL.Path):
		h.CreateUser(response, request)
		return
	case request.Method == http.MethodGet && listPostRe.MatchString(request.URL.Path):
		h.ListPosts(response, request)
		return
	case request.Method == http.MethodGet && getPostRe.MatchString(request.URL.Path):
		h.GetPost(response, request)
		return
	case request.Method == http.MethodPost && createPostRe.MatchString(request.URL.Path):
		h.CreatePost(response, request)
		return
	default:
		notFound(response, request)
		return
	}
}

// ListPosts : Get All Posts by a User
// URL : /posts/users/<id>
// Parameters: int id
// Method: GET
// Output: Array of JSON Encoded Post objects
func (h *requestHandler) ListPosts(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	matches := listPostRe.FindStringSubmatch(request.URL.Path)

	if len(matches) < 2 {
		notFound(response, request)
		return
	}

	page := "1"
	limit := 5

	// Extract page parameter
	pageMatch := listPostPageRe.FindStringSubmatch(request.URL.RequestURI())
	if len(pageMatch) >= 2 {
		page = pageMatch[1]
	}

	pg, err := strconv.Atoi(page)
	if err != nil {
		internalServerError(response, request)
	}

	var posts []post
	collection := client.Database("appointy").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		internalServerError(response, request)
		return
	}
	defer cursor.Close(ctx)

	// Pagination using offset computed using page number
	// Responding only with valid posts between range [low, high]
	curr := 0
	low := (pg - 1) * limit
	high := pg*limit - 1

	for cursor.Next(ctx) {
		var p post
		cursor.Decode(&p)
		if p.Author == matches[1] {
			if curr >= low && curr <= high {
				posts = append(posts, p)
			}
			curr++
		}
	}

	if err := cursor.Err(); err != nil {
		internalServerError(response, request)
		return
	}
	json.NewEncoder(response).Encode(posts)
}

// GetUser : Get a User with id
// URL : /users/<id>
// Parameters: int id
// Method: GET
// Output: JSON Encoded User object if found else JSON Encoded Exception.
func (h *requestHandler) GetUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	matches := getUserRe.FindStringSubmatch(request.URL.Path)
	if len(matches) < 2 {
		notFound(response, request)
		return
	}
	id := matches[1]

	var u user
	collection := client.Database("appointy").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&u)
	if err != nil {
		internalServerError(response, request)
		return
	}
	json.NewEncoder(response).Encode(u)
}

// GetUser : Get a Post with id
// URL : /posts/<id>
// Parameters: int id
// Method: GET
// Output: JSON Encoded Post object if found else JSON Encoded Exception.
func (h *requestHandler) GetPost(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	matches := getPostRe.FindStringSubmatch(request.URL.Path)
	if len(matches) < 2 {
		notFound(response, request)
		return
	}
	id := matches[1]

	var p post
	collection := client.Database("appointy").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&p)
	if err != nil {
		internalServerError(response, request)
		return
	}
	json.NewEncoder(response).Encode(p)
}

// CreateUser - Create User
// URL : /users
// Method: POST
// Body:
/*
 * {
	"id": "1"
 *	"name":"gandharv",
	"email": "...",
 *	"password": "...",
   }
*/
// Output: JSON Encoded User object if created else JSON Encoded Exception.
func (h *requestHandler) CreateUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var u user
	_ = json.NewDecoder(request.Body).Decode(&u)

	// Ensuring none of the fields is empty
	v := reflect.ValueOf(u)
	for i := 0; i < v.NumField(); i++ {
		field := fmt.Sprintf("%v", v.Field(i).Interface())
		if len(field) == 0 {
			BadRequest(response, request)
			return
		}
	}

	// Using SHA256 checksum to hash password
	encrypted := sha256.Sum256([]byte(u.Password))
	u.Password = hex.EncodeToString(encrypted[:])

	collection := client.Database("appointy").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, _ := collection.InsertOne(ctx, u)
	json.NewEncoder(response).Encode(result)
}

// CreatePost - Create Post
// URL : /posts
// Method: POST
// Body:
/*
 * {
	"id": "1"
 *	"author":"1",
	"caption": "...",
 *	"image": "...",
	"time": "..."
   }
*/
// Output: JSON Encoded Post object if created else JSON Encoded Exception.
func (h *requestHandler) CreatePost(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var p post
	_ = json.NewDecoder(request.Body).Decode(&p)

	// Ensuring none of the fields is empty
	v := reflect.ValueOf(p)
	for i := 0; i < v.NumField(); i++ {
		field := fmt.Sprintf("%v", v.Field(i).Interface())
		if len(field) == 0 {
			BadRequest(response, request)
			return
		}
	}
	collection := client.Database("appointy").Collection("posts")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, _ := collection.InsertOne(ctx, p)
	json.NewEncoder(response).Encode(result)
}

func internalServerError(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte(`{"error": "Internal server error"}`))
}

func notFound(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(http.StatusNotFound)
	response.Write([]byte(`{"error": "Not found"}`))
}

func BadRequest(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(http.StatusBadRequest)
	response.Write([]byte(`{"error": "Bad Request"}`))
}
