# Appointy-IG
## About API

Basic REST API implemented in Go.

- Based on Rest API Guidelines
- Using JSON body requests

## Built With
- Go
- MongoDB

## Features

- [x] Built using Go standard library
- [x] SHA256 hashed passwords
- [x] Pagination for posts/users/{userId} endpoint
- [x] Encpoints tested with unit tests
- [x] Mandatory fields in JSON request body


## Install and Run
- Clone project
- Install `Golang` 
- Fire up `MongoDB` at `mongodb://localhost:27017`
- Start the go server using 
```sh
go run main.go
```
- That's It, you can now send requests at `localhost:8080/{endpoint}`


## API Documentation


#### User 
> Every thing about Users
 
| API Route      | Functionality  |
| ------------- |:-------------:| 
| GET /user/{userId}     | User Detail      | 
| POST /user     |  Create new User      | 

#### Post 
 >Every thing about Posts
 
| API Route      | Functionality  |
| ------------- |:-------------:| 
| GET /posts/{postId}     | Post Details| 
| GET /posts/users/{userId}| All posts by a user | 
| POST /posts| Create new post| 

### Sample Requests
```sh
GET http://localhost:8080/users/1

GET http://localhost:8080/posts/users/1

GET http://localhost:8080/posts/users/1?page=2

GET http://localhost:8080/posts/2
```

## Testing
>Unit tests have been written for all the API endpoints

Tests can be run with
```sh
go test -v
```
### Output:
```sh
=== RUN   TestGetPostFound
--- PASS: TestGetPostFound (0.64s)
=== RUN   TestGetPostNotFound
--- PASS: TestGetPostNotFound (0.62s)
=== RUN   TestGetUserFound
--- PASS: TestGetUserFound (0.63s)
=== RUN   TestGetUserNotFound
--- PASS: TestGetUserNotFound (0.62s)
=== RUN   TestListPostsPaginated
--- PASS: TestListPostsPaginated (0.63s)
PASS
ok      ig-api  3.211s
```
