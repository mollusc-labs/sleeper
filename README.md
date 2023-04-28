# Sleeper

A blazingly <i>lazy</i> (let's call it simply beautiful) CouchDB client implementation in Go.

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
response, err := s.Find(`
    "selector": {
        "title": "Live And Let Die"
    },
    "fields": [
        "title"
    ]
`)
```

## License

Sleeper is licensed under the `Apache-2.0` license, the same as Apache CouchDB.
