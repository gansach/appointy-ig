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

func TestGetUserFound(t *testing.T) {
	request, err := http.NewRequest("GET", "/users/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	reqH := &requestHandler{}
	mux := http.NewServeMux()
	mux.Handle("/users/", reqH)

	reqH.ServeHTTP(response, request)

	if status := response.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestGetUserNotFound(t *testing.T) {
	request, err := http.NewRequest("GET", "/users/1000", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	reqH := &requestHandler{}
	mux := http.NewServeMux()
	mux.Handle("/users/", reqH)

	reqH.ServeHTTP(response, request)

	if status := response.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
