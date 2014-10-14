package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"

	"gateway/config"
	"gateway/db"
	"github.com/goraft/raft"
	"github.com/gorilla/mux"
)

// Server encapsulates the Raft server.
type Server struct {
	name       string
	conf       config.RaftServer
	router     *mux.Router
	RaftServer raft.Server
	httpServer *http.Server
}

// NewServer builds a new Raft server.
func NewServer(conf config.RaftServer) *Server {
	s := &Server{
		conf:   conf,
		router: mux.NewRouter(),
	}

	// Read existing name or generate a new one.
	nameFilePath := filepath.Join(conf.DataPath, "name")
	if b, err := ioutil.ReadFile(nameFilePath); err == nil {
		s.name = string(b)
	} else {
		s.name = fmt.Sprintf("%07x", rand.Int())[0:7]
		if err = ioutil.WriteFile(nameFilePath, []byte(s.name), 0644); err != nil {
			panic(err)
		}
	}

	return s
}

// Setup sets up the server.
func (s *Server) Setup(db db.DB) {
	var err error

	log.Printf("Initializing Raft Server: %s", s.conf.DataPath)

	// Initialize and start Raft server.
	transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
	s.RaftServer, err = raft.NewServer(s.name, s.conf.DataPath, transporter, nil, db, "")
	if err != nil {
		log.Fatal(err)
	}

	transporter.Install(s.RaftServer, s)
	s.RaftServer.Start()

	leader := s.conf.Leader
	if leader != "" {
		// Join to leader if specified.

		log.Println("Attempting to join leader:", leader)

		if !s.RaftServer.IsLogEmpty() {
			log.Fatal("Cannot join with an existing log")
		}
		if err := s.join(leader); err != nil {
			log.Fatal(err)
		}

	} else if s.RaftServer.IsLogEmpty() {
		// Initialize the server by joining itself.

		log.Println("Initializing new cluster")

		_, err := s.RaftServer.Do(&raft.DefaultJoinCommand{
			Name:             s.RaftServer.Name(),
			ConnectionString: s.connectionString(),
		})
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Println("Recovered from log")
	}

	log.Println("Initializing HTTP server")

	// Initialize and start HTTP server.
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.conf.Port),
		Handler: s.router,
	}

	s.router.HandleFunc("/join", s.joinHandler).Methods("POST")
}

// Run runs the server.
func (s *Server) Run() error {
	log.Println("Raft server listening at:", s.connectionString())
	return s.httpServer.ListenAndServe()
}

func (s *Server) connectionString() string {
	return fmt.Sprintf("http://%s:%d", s.conf.Host, s.conf.Port)
}

// HandleFunc is a hack around Gorilla mux not providing the correct net/http
// HandleFunc() interface.
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(pattern, handler)
}

// Joins to the leader of an existing cluster.
func (s *Server) join(leader string) error {
	command := &raft.DefaultJoinCommand{
		Name:             s.RaftServer.Name(),
		ConnectionString: s.connectionString(),
	}

	var b bytes.Buffer
	json.NewEncoder(&b).Encode(command)
	resp, err := http.Post(fmt.Sprintf("http://%s/join", leader), "application/json", &b)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func (s *Server) joinHandler(w http.ResponseWriter, req *http.Request) {
	command := &raft.DefaultJoinCommand{}

	if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := s.RaftServer.Do(command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
