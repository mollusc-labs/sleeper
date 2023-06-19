package sleeper

/*
   Copyright 2023 Mollusc Labs

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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Auth struct {
	user string
	pass string
}

type Protocol string

const (
	HTTP  Protocol = "http"
	HTTPS          = "https"
)

type Auditable struct {
	_id  string
	_rev string
}

type CouchResponse struct {
	Body    *json.RawMessage
	Headers *http.Header
}

type ParsedCouchDocumentResponse[T any] struct {
	Docs     []T    `json:"docs"`
	Bookmark string `json:"bookmark"`
}

type Config struct {
	protocol Protocol
	port     uint16
	timeout  uint64
	host     string
}

type Sleeper struct {
	auth    *Auth
	config  *Config
	headers map[string]string
}

func NewConfig(protocol Protocol, port uint16, timeout uint64, host string) *Config {
	c := new(Config)
	c.protocol = protocol
	c.port = port
	c.timeout = timeout
	c.host = host

	return c
}

func NewAuth(user, pass string) *Auth {
	a := new(Auth)
	a.pass = pass
	a.user = user

	return a
}

func New(conf *Config, auth *Auth) (*Sleeper, error) {
	s := new(Sleeper)
	// Auth can be nil, if it is, we simply don't use it later
	s.auth = auth

	if conf == nil {
		s.config = NewConfig("http", 5984, 5000, "127.0.0.1")
	} else {
		s.config = conf
	}

	s.headers = make(map[string]string)

	s.headers["user-agent"] = "go-sleeper"
	s.headers["content-type"] = "application/json; charset=utf8"

	if s.auth != nil {
		auth_str := fmt.Sprintf("%s:%s", s.auth.user, s.auth.pass)
		base64_str := base64.StdEncoding.EncodeToString([]byte(auth_str))
		s.headers["authorization"] = fmt.Sprintf("Basic %s", base64_str)
	}

	return s, nil
}

/*
 * Attempts to parse a CouchDB response body to [ParsedCouchDocumentResponse]
 */
func Parse[T any](msg json.RawMessage) (*ParsedCouchDocumentResponse[T], error) {
	m := ParsedCouchDocumentResponse[T]{}
	err := json.Unmarshal(msg, &m)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (s *Sleeper) fetch(db, method, location string, body []byte, query map[string]string) (*CouchResponse, error) {
	var request *http.Request

	raw_uri := fmt.Sprintf("%s://%s:%d", s.config.protocol, s.config.host, s.config.port)

	if db != "" {
		raw_uri += fmt.Sprintf("/%s", db)
	}

	if location != "" {
		raw_uri += fmt.Sprintf("/%s", location)
	}

	uri, err := url.Parse(raw_uri)

	if err != nil {
		return nil, errors.New("Could not generate URI from configuration")
	}

	for k, v := range query {
		uri.Query().Add(k, v)
	}

	if body != nil {
		request, err = http.NewRequest(method, uri.String(), bytes.NewBuffer(body))
	} else {
		request, err = http.NewRequest(method, uri.String(), nil)
	}

	if err != nil {
		return nil, err
	}

	for k, v := range s.headers {
		request.Header.Add(k, v)
	}

	if _, t := os.LookupEnv("SLEEPER_TRACE"); t {
		log.Println(fmt.Sprintf("Sleeper log: \n %s %s \n %v \n %s", method, uri.String(), request.Header, string(body)))
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	t, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	response_body := json.RawMessage(t)

	if response.StatusCode >= 400 {
		// https://docs.couchdb.org/en/stable/api/basics.html#http-status-codes
		return &CouchResponse{Body: &response_body, Headers: &response.Header},
			errors.New(string(response_body)) // CouchDB provides fairly helpful errors
	}

	return &CouchResponse{Body: &response_body, Headers: &response.Header}, nil
}

/*
 * Saves a document to the database. If there is a "_id" field in the interface
 * CouchDB will update the document if it exists, otherwise it will generate a new
 * document and use that id or generate a new id if no id is provided.
 */
func (s *Sleeper) Save(db string, data interface{}) (*CouchResponse, error) {
	json, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	return s.fetch(db, "POST", "", json, nil)
}

/*
 * See [Save] but takes a list of documents to save
 */
func (s *Sleeper) SaveMany(db string, data []interface{}) (*CouchResponse, error) {
	v := make(map[string][]interface{})
	v["docs"] = data

	json, err := json.Marshal(v)

	if err != nil {
		return nil, err
	}

	return s.fetch(db, "POST", "bulk_docs", json, nil)
}

/*
 * A more explicit save, this call will make sure the "_id" and "_rev" fields are included
 */
func (s *Sleeper) Update(db string, data interface{}) (*CouchResponse, error) {
	u, ok := data.(Auditable)

	if !ok {
		return nil, errors.New("_id and _rev need to exist on the data you plan to update.")
	}

	json, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	return s.fetch(db, "PUT", fmt.Sprintf("%s", url.QueryEscape(u._id)), json, nil)
}

func (s *Sleeper) Delete(db, id, rev string) (*CouchResponse, error) {
	query := make(map[string]string)
	query["id"] = id
	query["rev"] = rev
	return s.fetch(db, "DELETE", "", nil, query)
}

/*
 *	Deletes many documents based on a map of ids to a list of revisions
 *  https://docs.couchdb.org/en/stable/api/database/misc.html
 */
func (s *Sleeper) DeleteMany(db string, ids_and_revisions map[string][]string) (*CouchResponse, error) {
	json, err := json.Marshal(ids_and_revisions)
	if err != nil {
		return nil, errors.New("Couldn't JSONify (this should never happen)")
	}

	return s.fetch(db, "POST", "_purge", json, nil)
}

/*
 * Attempts to create a database with a given name
 */
func (s *Sleeper) CreateDatabase(db string) (*CouchResponse, error) {
	return s.fetch(db, "PUT", "", nil, nil)
}

/*
 * Attempts to delete a database with a given name
 */
func (s *Sleeper) DropDatabase(db string) (*CouchResponse, error) {
	return s.fetch(db, "DELETE", "", nil, nil)
}

/*
 * Attempts to find a list of documents based on a query
 */
func (s *Sleeper) Find(db, view string, query map[string]interface{}) (*CouchResponse, error) {
	key_to_enclude := func(str string) bool {
		return str == "key" || str == "keys" || str == "startkey" || str == "endkey"
	}

	sanitized_query := make(map[string]string)
	if query != nil {
		for k, v := range query {
			if key_to_enclude(k) {
				json, err := json.Marshal(v)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Could not marshal value with key %s", k))
				}
				sanitized_query[k] = string(json)
			} else {
				sanitized_query[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	return s.fetch(db, "GET", fmt.Sprintf("%s", view), nil, sanitized_query)
}

/*
 * Uses a Mango query to find documents
 */
func (s *Sleeper) Mango(db, query string) (*CouchResponse, error) {
	return s.fetch(db, "POST", "_find", []byte(query), nil)
}

/*
 * Takes a query that can be represented by a struct or map instead of a string
 */
func (s *Sleeper) MangoStruct(db string, query interface{}) (*CouchResponse, error) {
	json, err := json.Marshal(query)

	if err != nil {
		return nil, errors.New("Could not parse JSON from interface, maybe try Mango(string) instead")
	}

	return s.Mango(db, string(json))
}

/*
 * Finds a list of [count] UUID's
 */
func (s *Sleeper) NewUUID(count uint) ([]string, error) {
	if count < 1 {
		return nil, errors.New("Can't find 0 uuids, count must be greater than 0")
	}

	uri, err := url.Parse(fmt.Sprintf("%s://%s:%d/%s", s.config.protocol, s.config.host, s.config.port, "_uuids"))

	if err != nil {
		return nil, errors.New("Could not parse URI")
	}

	uri.Query().Add("count", fmt.Sprint(count))

	response, err := http.DefaultClient.Get(uri.String())

	if err != nil {
		return nil, err
	}

	var body json.RawMessage
	_, err = response.Body.Read(body)

	if err != nil {
		return nil, err
	}

	var v map[string][]string

	err = json.Unmarshal(body, &v)

	if err != nil {
		return nil, err
	}

	return v["uuids"], nil
}

/*
 * Lists all databases on the CouchDB server
 */
func (s *Sleeper) ListDatabases() ([]string, error) {
	res, err := s.fetch("", "GET", "_all_dbs", nil, nil)

	if err != nil {
		return nil, err
	}

	var arr []string
	_ = json.Unmarshal(*res.Body, &arr)
	return arr, nil
}
