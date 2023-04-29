package sleeper_test

/*
   Copyright 2023 Mollusc Labs Inc

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"net"
	"net/http"
	"testing"

	"github.com/mollusc-labs/sleeper"
)

func TestNewAuth(t *testing.T) {
	sleeper.NewAuth("foo", "bar")
}

func TestNewConfig(t *testing.T) {
	sleeper.NewConfig("http", 5984, 5000, "127.0.0.1")
}

func TestNew(t *testing.T) {
	a := sleeper.NewAuth("foo", "bar")
	c := sleeper.NewConfig("http", 5984, 5000, "127.0.0.1")
	_, err := sleeper.New("posts", c, a)

	if err != nil {
		t.Error("Error should be nil")
	}
}

type Foo struct {
	Name string `json:"name"`
}

func createListener() net.Listener {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	return l
}

type test_handler struct{}

const test_json string = `{"docs":[
        {"name":"foo"}
        ],
        "bookmark": "g1AAAABweJzLYWBgYMpgSmHgKy5JLCrJTq2MT8lPzkzJBYorJJsnmRqmpqWlGRgkp6YlphlbJiemJicZWpoZGBiZG5iB9HHA9BGlIwsA9pcf4w",
        "warning": "No matching index found, create an index to optimize query time."}`

func (t *test_handler) ServeHTTP(io http.ResponseWriter, req *http.Request) {
	io.Write([]byte(test_json))
	io.Header().Add("content-type", "application/json; charset=utf8")
}

func TestParseDocs(t *testing.T) {
	resp, _ := sleeper.ParseDocs[Foo]([]byte(test_json))
	t.Logf("%v\n", resp)
	for _, v := range resp.Docs {
		if v.Name != "foo" {
			t.Fatal("Name was not Foo, ParseDocs is not working.")
		}
	}
}

func TestFull(t *testing.T) {

	l := createListener()

	go http.Serve(l, &test_handler{})

	a := sleeper.NewAuth("foo", "password")                                                 // username, password
	c := sleeper.NewConfig("http", uint16(l.Addr().(*net.TCPAddr).Port), 5000, "127.0.0.1") // protocol, port, timeout, host
	s, _ := sleeper.New("posts", c, a)                                                      // posts is the DB for this sleeper instance

	response, err := s.Mango(`{
    "selector": {}, 
    "fields": [
        "name"
    ]}`)

	if err != nil {
		t.Fatal("You need a CouchDB instance to fully test this module")
	} else {
		t.Logf("%v\n", string(*response.Body))
		resp, _ := sleeper.ParseDocs[Foo](*response.Body)
		t.Logf("%v\n", resp)

		for _, v := range resp.Docs {
			if v.Name != "foo" {
				t.Fatal("Failed to decode response with ParseDocs")
			}
		}
	}

	l.Close()
}
