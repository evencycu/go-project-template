package mgopool

import (
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	conntrack "github.com/mwitkow/go-conntrack"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gopkg"
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
	addr string
	s    *mgo.Session
}

// Session returns mgo.Session
func (s *Session) Session() *mgo.Session {
	return s.s
}

// Addr returns target mongos addr
func (s *Session) Addr() string {
	return s.addr
}

// Pool is the mgo session pool
type Pool struct {
	cap         int
	config      *DBInfo
	mode        mgo.Mode
	c           chan *Session
	available   bool
	liveServers []string
}

func newSession(dbi *DBInfo, addr []string, mode mgo.Mode) (newSession *mgo.Session, err error) {
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
		Addrs:      addr,
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
		errStr := "[mongo] NewSession no reachable server"
		errLog(systemCtx, strings.Join(addr, ","), errStr)
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
	c := make(chan *Session, dbi.MaxConn)

	addrAllocations := make(map[string]int)
	// get LiveServers
	rootSession, dialErr := newSession(dbi, dbi.Addrs, mgo.Primary)
	if dialErr != nil {
		errLog(systemCtx, strings.Join(dbi.Addrs, ","), "unable to connect to mongoDB:"+dialErr.Error())
		return dialErr
	}
	defer rootSession.Close()
	liveServers := rootSession.LiveServers()
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
		addrs := liveServers
		if dbi.Mongos {
			addrs = []string{liveServers[i%lengthOfMongos]}
		}
		newSession, err := newSession(dbi, addrs, dbi.ReadMode)
		if err == nil {
			addr := strings.Join(addrs, ",")
			addrAllocations[addr] = addrAllocations[addr] + 1
			c <- &Session{addr: addr, s: newSession}
			sessionCount++
		}
	}

	for k, v := range addrAllocations {
		log.Println("[mongo] mongo host: ", k, ": connnections:", v)
	}
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
	return p.available
}

func (p *Pool) get(ctx goctx.Context) (session *Session, err gopkg.CodeError) {
	if !p.available {
		err = ErrMongoPoolClosed
		return
	}

	select {
	case session = <-p.c:
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
}

// Close gracefull shutdown conns and Pool status
func (p *Pool) Close() {
	// wait all session come back to pool
	p.available = false
	for i := 0; i < 5; i++ {
		// len(p.c) should not > than p.cap, but use >= to get along with error
		if len(p.c) >= p.cap {
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}

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
	return config
}

func (p *Pool) backgroundReconnect(s *Session) error {
	s.s.Close()

	retry := 3
	for i := 0; i < retry; i++ {
		// newSession timeout in 15s
		newS, err := newSession(p.config, []string{s.addr}, p.mode)
		if err == nil {
			s.s = newS
			p.put(s)
			return nil
		}
	}

	// still cannot connet after 45 secs
	errRetryTotalFailed := gopkg.NewError("[mongo] Reconnect failed")
	errLog(systemCtx, s.addr, errRetryTotalFailed.Error())
	p.cap--
	if p.cap == 0 || p.cap < p.config.MaxConn/2 {
		log.Println("start recover")
		// should do error handling, since session drops to half of expectation
		return p.Recover()
	}
	select {
	case AlertChannel <- errRetryTotalFailed:
		log.Println("set alert")
	default:
		log.Println("ignore alert")
		// just pass, no spam our alert and non-blocking
	}
	return errRetryTotalFailed
}

// Recover close and re-create the pool sessions
func (p *Pool) Recover() error {
	log.Println("Start Pool Recover")
	p.Close()
	return p.Init(p.config)
}
