package main

import (
	"database/sql"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hrashk/shorturl/internal/app"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/suite"
)

type mainServer struct {
	parentSuite   *suite.Suite
	origNewServer func(modifiers ...app.Configurator) (*http.Server, error)
	server        *http.Server
	baseURL       string
	ch            chan struct{}
}

func newServer(parentSuite *suite.Suite) *mainServer {
	ms := &mainServer{origNewServer: app.NewServer, parentSuite: parentSuite}
	app.NewServer = ms.spy

	ms.ch = make(chan struct{}, 1)
	ms.ch <- struct{}{}

	return ms
}

func (ms *mainServer) spy(modifiers ...app.Configurator) (*http.Server, error) {
	srv, err := ms.origNewServer(modifiers...)

	if err == nil && srv != nil {
		<-ms.ch
		ms.server = srv
		ms.baseURL = ms.inferBaseURL()
		ms.ch <- struct{}{}
	}

	return srv, err
}

func (ms *mainServer) inferBaseURL() string {
	if ms.server.Addr[0] == ':' {
		return "http://localhost" + ms.server.Addr
	} else if !strings.Contains(ms.server.Addr, "://") {
		return "http://" + ms.server.Addr
	} else {
		return ms.server.Addr
	}
}

func (ms *mainServer) addr() string {
	var addr string

	if ms.server != nil {
		addr = ms.server.Addr
	}

	return addr
}

func (ms *mainServer) wipeData() {
	ms.deleteFiles()

	ms.wipeDB()
}

func (ms *mainServer) deleteFiles() {
	ms.deleteFile(app.DefaultStoragePath)
	ms.deleteFile(samplePath)
	ms.deleteFile(anotherPath)
}

func (ms *mainServer) wipeDB() {
	db, err := sql.Open("pgx", app.DefaultDatabaseDsn)
	if err != nil {
		ms.parentSuite.T().Logf("Unable to connect to db: %v", err)
		return
	}

	_, err = db.Exec("drop table if exists urls")
	if err != nil {
		ms.parentSuite.T().Logf("Unable to drop table urls: %v", err)
		return
	}
}

func (ms *mainServer) deleteFile(path string) {
	if err := os.Remove(path); err != nil {
		ms.parentSuite.ErrorIs(err, os.ErrNotExist, "failed to delete file %s", path)
	}
}

func (ms *mainServer) stop() {
	if ms.server != nil {
		ms.server.Close()
		ms.server = nil
		ms.baseURL = ""
	}
}

func (ms *mainServer) start() {
	go main()

	ms.waitForPort()
}

func (ms *mainServer) waitForPort() {
	const timeout = time.Second
	const pollInterval = 50 * time.Millisecond

	var timer = time.NewTimer(timeout)
	var ticker = time.NewTicker(pollInterval)

	for {
		select {
		case <-timer.C:
			ms.parentSuite.Require().Fail("timed out connecting to server")
			return
		case <-ticker.C:
			if ms.portIsOpen(timeout) {
				return
			}
		}
	}
}

func (ms *mainServer) portIsOpen(timeout time.Duration) bool {
	addr := ms.addrSync()

	if addr == "" {
		return false
	}

	return ms.dialSuccessful(addr, timeout)
}

func (ms *mainServer) addrSync() string {
	var addr string

	<-ms.ch
	if ms.server != nil {
		addr = ms.server.Addr
	}
	ms.ch <- struct{}{}

	return addr
}

func (ms *mainServer) dialSuccessful(addr string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err == nil {
		conn.Close() // the port is open
		return true
	}
	return false
}
