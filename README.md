# Sleeper

A blazingly <i>lazy</i> CouchDB client implementation in Go.

## Install

```bash
go get github.com/mollusc-labs/sleeper
```

## How to use

Sleeper is incredibly simple. First, make an auth and config struct,
then inject it.

```go
a := sleeper.NewAuth("foo", "bar") // username, password
c := sleeper.NewConfig("http", 5984, 5000, "127.0.0.1") // protocol, port, timeout, host
s, err := sleeper.New("posts", c, a) // posts is the DB for this sleeper instance
```

Then you can query the database or create your database, etc:

```go
_, err := s.CreateDatabase()

if err != nil {
    log.Fatal("Couldn't create database :(")
}
```

```go
type Book struct {
    Title string  `json:"title"`
    Author string `json:"author"`
}

response, err := s.Mango(`{
    "selector": {
        "title": "Live And Let Die"
    },
    "fields": [
        "title",
        "author"
    ]
}`)

b := Book{}
err := json.Unmarshal(response.Body, &b)

fmt.Printf("Book is %v by %v\n", b.Title, b.Author)
```

## License

Sleeper is licensed under the `Apache-2.0` license, the same as Apache CouchDB.
