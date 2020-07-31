package mgopool

import (
	"math/rand"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	conntrack "github.com/eaglerayp/go-conntrack"
	"github.com/globalsign/mgo"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/cake/goctx"
	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/gotrace/v2"
	"gitlab.com/cake/m800log"
)

var (
	poolWaitingDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "mgopool",
			Subsystem:  "pool",
			Name:       "get_duration",
			Help:       "pool.get latencies in seconds.",
			Objectives: map[float64]float64{0.8: 0.05, 0.95: 0.005},
		}, []string{"pool_name"})
)

func init() {
	prometheus.MustRegister(poolWaitingDuration)
}

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
	name        string
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
	conntrackDialer :=
		func(addr *mgo.ServerAddr) (net.Conn, error) {
			conntrackDialer := conntrack.NewDialFunc(
				conntrack.DialWithName("mgopool"),
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

	for attempts := 1; attempts <= 10; attempts++ {
		newSession, err = mgo.DialWithInfo(&dialInfo)
		if err == nil {
			break
		}
		errLogf(systemCtx, addrs, "[mongo] NewSession error: %v", err)
		time.Sleep(time.Duration(attempts) * time.Second)
	}
	if err != nil {
		errLogf(systemCtx, addrs, "[mongo] NewSession no reachable server error: %v", err)
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
	// mask password for security concern
	password := dbi.Password
	var pb strings.Builder
	for i := 0; i < len(password); i++ {
		pb.WriteString("*")
	}
	dbi.Password = pb.String()
	m800log.Infof(systemCtx, "[mgopool] init with config: %+v", dbi)

	// recover password
	dbi.Password = password
	c := make(chan *Session, dbi.MaxConn)
	addrAllocations := make(map[string]int)

	// get LiveServers
	rootSession, dialErr := newSession(dbi, dbi.Addrs, mgo.Primary)
	if dialErr != nil {
		errLogf(systemCtx, dbi.Addrs, "unable to connect to mongoDB errpr: %v", dialErr)
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
	p.name = dbi.Name
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

	// prevent ping, no root span case
	t := ctx.GetSpan()
	if t != nil {
		sp := gotrace.CreateChildOfSpan(ctx, FuncPoolWaiting)
		defer sp.Finish()
	}

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		poolWaitingDuration.WithLabelValues(p.name).Observe(v)
	}))
	defer timer.ObserveDuration()

	select {
	case session = <-p.c:
		p.wg.Add(1)
	case <-ctx.Done():
		err = gopkg.NewCarrierCodeError(ContextTimeout, "context timeout:"+ctx.Err().Error())
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
	m800log.Info(systemCtx, "[mgopool] start session reconnect")

	for {
		newS, err := newSession(p.config, s.Addr(), p.mode)
		if err == nil {
			s.s = newS
			p.put(s)
			return nil
		}
		m800log.Errorf(systemCtx, "[mgopool] session still try reconnect: %v", err)
	}
}

// Recover close and re-create the pool sessions
func (p *Pool) Recover() error {
	m800log.Info(systemCtx, "[mgopool] start recover")
	p.Close()
	for {
		err := p.Init(p.config)
		if err == nil {
			return nil
		}
		m800log.Errorf(systemCtx, "[mgopool] still try recover error:%v", err)
	}
}
