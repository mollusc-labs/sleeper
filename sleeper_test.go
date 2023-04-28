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
	"encoding/json"
	"testing"

	"github.com/mollusc-labs/sleeper"
)

func TestNew(t *testing.T) {
	a := sleeper.NewAuth("foo", "bar")
	c := sleeper.NewConfig("http", 5984, 5000, "127.0.0.1")
	_, err := sleeper.New("posts", c, a)

	if err != nil {
		t.Error("Error should be nil")
	}
}

func TestRunning(t *testing.T) {
	a := sleeper.NewAuth("foo", "bar")                      // username, password
	c := sleeper.NewConfig("http", 5984, 5000, "127.0.0.1") // protocol, port, timeout, host
	s, _ := sleeper.New("posts", c, a)                      // posts is the DB for this sleeper instance

	response, err := s.Find(`
    "selector": {
        "title": "Live And Let Die"
    },
    "fields": [
        "title",
        "author"
    ]`, nil)

	if err != nil {
		t.Logf("%v\n", err)
	}

	b := Book{}
	err = json.Unmarshal(*response.Body, &b)

	t.Logf("Book is %v by %v\n", b.Title, b.Author)

}

type Book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}
