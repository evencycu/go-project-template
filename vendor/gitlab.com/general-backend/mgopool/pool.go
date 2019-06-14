package mgopool

import (
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/globalsign/mgo"
	conntrack "github.com/mwitkow/go-conntrack"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gopkg"
	"gitlab.com/general-backend/m800log"
)

// AlertChannel put error message, wait for outer user (i.e., gobuster) pick and send.
var AlertChannel = make(chan error, 1)

// Default dial timeout value from https://github.com/globalsign/mgo/blob/v2/cluster.go
var syncSocketTimeout = 5 * time.Second

// DBInfo logs the required info for baas mongodb.
type DBInfo struct {
	Name         string
	User         string
	Password     string
	AuthDatabase string
	Addrs        []string
	MaxConn      int
	Timeout      time.Duration
	ReadMode     mgo.Mode
	Direct       bool
	Mongos       bool
}

// NewDBInfo
func NewDBInfo(name string, addrs []string, user, password, authdbName string,
	timeout time.Duration, maxConn int, direct, readSecondary, mongos bool) *DBInfo {
	readMode := mgo.Primary
	if readSecondary {
		readMode = mgo.SecondaryPreferred
	}
	return &DBInfo{
		MaxConn:      maxConn,
		Name:         name,
		Addrs:        addrs,
		User:         user,
		Password:     password,
		AuthDatabase: authdbName,
		Timeout:      timeout,
		Direct:       direct,
		ReadMode:     readMode,
		Mongos:       mongos,
	}
}

// Session is wrapper for mgo.Session with logging mongos addr
type Session struct {
	addrs []string
	s     *mgo.Session
}

// Session returns mgo.Session
func (s *Session) Session() *mgo.Session {
	return s.s
}

// Addr returns target mongos addr
func (s *Session) Addr() []string {
	return s.addrs
}

// Pool is the mgo session pool
type Pool struct {
	cap         int
	config      *DBInfo
	mode        mgo.Mode
	c           chan *Session
	available   bool
	rwLock      sync.RWMutex
	wg          sync.WaitGroup
	liveServers []string
}

func newSession(dbi *DBInfo, addrs []string, mode mgo.Mode) (newSession *mgo.Session, err error) {
	podName, err := os.Hostname()
	if err != nil {
		panic("Get hostname failed")
	}
	conntrackDialer :=
		func(addr *mgo.ServerAddr) (net.Conn, error) {
			conntrackDialer := conntrack.NewDialFunc(
				conntrack.DialWithName(podName),
				conntrack.DialWithTracing(),
				conntrack.DialWithDialer(&net.Dialer{
					Timeout: syncSocketTimeout,
				}),
			)
			return conntrackDialer(addr.TCPAddr().Network(), addr.String())
		}

	dialInfo := mgo.DialInfo{
		Addrs:      addrs,
		Direct:     dbi.Direct,
		FailFast:   true,
		Source:     dbi.AuthDatabase,
		Username:   dbi.User,
		Password:   dbi.Password,
		Timeout:    dbi.Timeout,
		PoolLimit:  1,
		DialServer: conntrackDialer,
	}

	for attempts := 1; attempts <= 5; attempts++ {
		newSession, err = mgo.DialWithInfo(&dialInfo)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * time.Second)
	}
	if err != nil {
		errLogf(systemCtx, addrs, "[mongo] NewSession no reachable server")
		return
	}
	newSession.SetMode(mode, true)
	err = newSession.Ping()
	if err != nil {
		newSession.Close()
	}

	return
}

// NewSessionPool construct connection pool
func NewSessionPool(dbi *DBInfo) (*Pool, error) {
	p := &Pool{}
	err := p.Init(dbi)
	return p, err
}

// Init returns whether Pool availalbe
func (p *Pool) Init(dbi *DBInfo) error {
	m800log.Infof(systemCtx, "[mgopool] init with config: %+v", dbi)
	c := make(chan *Session, dbi.MaxConn)

	addrAllocations := make(map[string]int)
	// get LiveServers
	rootSession, dialErr := newSession(dbi, dbi.Addrs, mgo.Primary)
	if dialErr != nil {
		errLog(systemCtx, dbi.Addrs, "unable to connect to mongoDB:", dialErr.Error())
		return dialErr
	}
	defer rootSession.Close()
	liveServers := rootSession.LiveServers()
	m800log.Infof(systemCtx, "[mgopool] liveservers: %s", liveServers)
	lengthOfMongos := len(liveServers)
	if dbi.Mongos {
		// shuffle the server lists
		for i := range liveServers {
			j := rand.Intn(i + 1)
			liveServers[i], liveServers[j] = liveServers[j], liveServers[i]
		}
	}

	sessionCount := 0
	for i := 0; i < dbi.MaxConn; i++ {
		var addrs []string
		sort.Strings(liveServers)
		if dbi.Mongos {
			addrs = []string{liveServers[i%lengthOfMongos]}
		} else {
			addrs = dbi.Addrs
		}
		newSession, err := newSession(dbi, addrs, dbi.ReadMode)
		if err == nil {
			addr := strings.Join(addrs, ",")
			addrAllocations[addr] = addrAllocations[addr] + 1
			c <- &Session{addrs: addrs, s: newSession}
			sessionCount++
		}
	}

	m800log.Infof(systemCtx, "[mgopool] mongo addrAllocations: %+v", addrAllocations)
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	p.liveServers = liveServers
	p.c = c
	p.config = dbi
	p.cap = sessionCount
	p.available = true
	p.mode = dbi.ReadMode
	return nil
}

// IsAvailable returns whether Pool availalbe
func (p *Pool) IsAvailable() bool {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	return p.available
}

func (p *Pool) get(ctx goctx.Context) (session *Session, err gopkg.CodeError) {
	if !p.IsAvailable() {
		err = ErrMongoPoolClosed
		return
	}

	select {
	case session = <-p.c:
		p.wg.Add(1)
	case <-ctx.Done():
		err = gopkg.NewCarrierCodeError(APIFullResource, "mongo resource not enough:"+ctx.Err().Error())
		return
	}
	return
}

// Len returns current Pool availalbe connections
func (p *Pool) Len() int {
	return len(p.c)
}

// LiveServers returns current Pool live servers list
func (p *Pool) LiveServers() []string {
	return p.liveServers
}

// Cap returns Pool capacity
func (p *Pool) Cap() int {
	return p.cap
}

// Mode returns mgo.Mode settings of Pool
func (p *Pool) Mode() mgo.Mode {
	return p.mode
}

// Config returns DBInfo of Pool
func (p *Pool) Config() *DBInfo {
	return p.config
}

func (p *Pool) put(session *Session) {
	p.c <- session
	p.wg.Done()
}

// Close gracefull shutdown conns and Pool status
func (p *Pool) Close() {
	// wait all session come back to pool
	p.rwLock.Lock()
	p.available = false
	p.rwLock.Unlock()
	p.wg.Wait()
	// since Close() is seldom called,
	// just share the rwlock with available()
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	close(p.c)
	for s := range p.c {
		s.s.Close()
	}
	p.cap = 0
	p.c = nil
}

// ShowConfig returns debug config info
func (p *Pool) ShowConfig() map[string]interface{} {
	config := make(map[string]interface{})
	config["MaxConn"] = p.config.MaxConn
	config["Addrs"] = p.config.Addrs
	config["Timeout"] = p.config.Timeout
	config["Direct"] = p.config.Direct
	config["Mongos"] = p.config.Mongos
	config["ReadMode"] = p.config.ReadMode
	config["User"] = p.config.User
	config["AuthDatabase"] = p.config.AuthDatabase
	return config
}

func (p *Pool) backgroundReconnect(s *Session) error {
	s.s.Close()
	m800log.Info(systemCtx, "[mgopool] start reconnect")

	retry := 3
	for i := 0; i < retry; i++ {
		// newSession timeout in 15s
		newS, err := newSession(p.config, s.Addr(), p.mode)
		if err == nil {
			s.s = newS
			p.put(s)
			return nil
		}
	}

	// still cannot connet after 45 secs
	errRetryTotalFailed := gopkg.NewError("[mongo] Reconnect failed")
	errLog(systemCtx, s.Addr(), errRetryTotalFailed.Error())
	p.cap--
	if p.cap == 0 || p.cap < p.config.MaxConn/2 {
		// should do error handling, since session drops to half of expectation
		return p.Recover()
	}
	select {
	case AlertChannel <- errRetryTotalFailed:
		m800log.Debug(systemCtx, "[mgopool] send retry alert")
	default:
		// just pass, no spam our alert and non-blocking
	}
	return errRetryTotalFailed
}

// Recover close and re-create the pool sessions
func (p *Pool) Recover() error {
	m800log.Info(systemCtx, "[mgopool] start recover")
	p.Close()
	return p.Init(p.config)
}
