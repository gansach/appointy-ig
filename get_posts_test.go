package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestGetPostFound(t *testing.T) {
	request, err := http.NewRequest("GET", "/posts/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := request.URL.Query()
	q.Add("page", "1")
	request.URL.RawQuery = q.Encode()
	response := httptest.NewRecorder()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	reqH := &requestHandler{}
	mux := http.NewServeMux()
	mux.Handle("/posts/", reqH)

	reqH.ServeHTTP(response, request)

	if status := response.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
func TestGetPostNotFound(t *testing.T) {
	request, err := http.NewRequest("GET", "/posts/100", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := request.URL.Query()
	q.Add("page", "1")
	request.URL.RawQuery = q.Encode()
	response := httptest.NewRecorder()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	reqH := &requestHandler{}
	mux := http.NewServeMux()
	mux.Handle("/posts/", reqH)

	reqH.ServeHTTP(response, request)

	if status := response.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
