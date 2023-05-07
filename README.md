# Sleeper

A blazingly <i>lazy</i> CouchDB client implementation in idiomatic Go. Sleeper
tries its best to follow Data-Oriented Programming by keeping data retrieved and sent to
CouchDB as `[]byte`, this allows for very simple and flexible usage and abstractions.

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
s, err := sleeper.New(c, a)
```

Then you can query the database or create your database, etc:

```go
_, err := s.CreateDatabase("books")

if err != nil {
    log.Fatal("Couldn't create 'books' database :(")
}
```

Sleeper's `CouchResponse` is simply a cut-down HTTP response, it has the raw
message as a `json.RawMessage` (`[]byte`) type, stored in `CouchResponse.Body` (which can be parsed via `sleeper.Parse[T]`).
It also has the headers of the response stored in `CouchResponse.Headers`.

Body is kept as a `[]byte` because it makes it much more flexible to leave the data as data.
If you really need to manipulate it you can use `sleeper.Parse[T]`.

```go
type Book struct {
    Title string  `json:"title"`
    Author string `json:"author"`
}

response, err := s.Mango("books", `{
    "selector": {
        "title": "Live And Let Die"
    },
    "fields": [
        "title",
        "author"
    ]
}`)

if err != nil {
    log.Fatal("Failed to query... :(")
} else {
    fmt.Printf("%s\n", string(*response.Body))
    parsed_response, _ := sleeper.Parse[Book](*response.Body)
    for _, v := range parsed_response.Docs {
        fmt.Printf("Found book: %v by %v", v.Title, v.Author)
    }
}
```

## License

Sleeper is licensed under the `Apache-2.0` license, the same as Apache CouchDB.
