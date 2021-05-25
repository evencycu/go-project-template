package mgopool

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/m800log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/eaglerayp/go-conntrack"
)

// AlertChannel put error message, wait for outer user (i.e., gobuster) pick and send.
var AlertChannel = make(chan error, 1)

// Default dial timeout value from https://gitlab.com/cake/mgo/blob/v2/cluster.go
var syncSocketTimeout = 5 * time.Second

// DBInfo logs the required info for baas mongodb.
type DBInfo struct {
	Name               string
	User               string
	Password           string
	AuthDatabase       string
	Addrs              []string
	MaxConn            int
	MaxConnectAttempts int
	Timeout            time.Duration
	ReadMode           readpref.Mode
	Direct             bool
	Mongos             bool
}

// NewDBInfo
func NewDBInfo(name string, addrs []string, user, password, authdbName string,
	timeout time.Duration, maxConn int, direct, readSecondary, mongos bool) *DBInfo {
	readMode := readpref.PrimaryPreferredMode
	if readSecondary {
		readMode = readpref.SecondaryPreferredMode
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

// Pool is the mgo session pool
type Pool struct {
	name      string
	config    *DBInfo
	mode      readpref.Mode
	available bool
	rwLock    sync.RWMutex
	client    *mongo.Client
}

func newClient(dbi *DBInfo, addrs []string) (newClient *mongo.Client, err error) {
	account := ""
	if dbi.User != "" && dbi.Password != "" {
		account = fmt.Sprintf("%s:%s@", url.QueryEscape(dbi.User), url.QueryEscape(dbi.Password))
	}
	uri := fmt.Sprintf("mongodb://%s%s/%s", account, strings.Join(addrs, ","), dbi.AuthDatabase)
	clientOpt := options.Client().ApplyURI(uri)

	conntrackDialer := conntrack.NewDialer(
		conntrack.DialWithName("mgopool"),
		conntrack.DialWithTracing(),
	)
	clientOpt.SetDialer(conntrackDialer)
	clientOpt.SetAppName(gopkg.GetAppName())
	clientOpt.SetConnectTimeout(dbi.Timeout)
	clientOpt.SetSocketTimeout(syncSocketTimeout)
	clientOpt.SetDirect(dbi.Direct)
	clientOpt.SetMaxPoolSize(uint64(dbi.MaxConn))
	clientOpt.SetMinPoolSize(uint64(dbi.MaxConn))
	readPref, _ := readpref.New(dbi.ReadMode)
	clientOpt.SetReadPreference(readPref)
	// The default read preference is primary

	maxAttempts := 10
	if dbi.MaxConnectAttempts > 0 {
		maxAttempts = dbi.MaxConnectAttempts
	}
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		newClient, err = mongo.Connect(context.Background(), clientOpt)
		if err == nil {
			break
		}
		errLogf(systemCtx, addrs, "[mongo] NewClient error: %v", err)
		time.Sleep(time.Duration(attempts) * time.Second)
	}
	if err != nil {
		errLogf(systemCtx, addrs, "[mongo] NewClient no reachable server error: %v", err)
		return
	}
	err = newClient.Ping(context.TODO(), nil)
	if err != nil {
		newClient.Disconnect(context.TODO())
	}

	return
}

// NewSessionPool construct connection pool
func NewSessionPool(dbi *DBInfo) (*Pool, error) {
	p := &Pool{}
	err := p.Init(dbi)
	return p, err
}

// Init returns whether Pool available
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

	// connection establish
	client, dialErr := newClient(dbi, dbi.Addrs)
	if dialErr != nil {
		errLogf(systemCtx, dbi.Addrs, "unable to connect to mongoDB error: %v", dialErr)
		return dialErr
	}
	p.name = dbi.Name
	p.config = dbi
	p.available = true
	p.mode = dbi.ReadMode
	p.client = client

	return nil
}

func (p *Pool) GetMongoClient() (*mongo.Client, error) {
	if p.client == nil {
		return nil, errors.New("mongo client empty")
	}

	return p.client, nil
}

// IsAvailable returns whether Pool availalbe
func (p *Pool) IsAvailable() bool {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	return p.available
}

// Len returns current Pool available connections
func (p *Pool) Len() int {
	if p.IsAvailable() {
		return p.config.MaxConn
	}
	return 0
}

// LiveServers returns current Pool live servers list
func (p *Pool) LiveServers() []string {
	if p.IsAvailable() {
		return p.config.Addrs
	}
	return []string{}
}

// Cap returns Pool capacity
func (p *Pool) Cap() int {
	if p.IsAvailable() {
		return p.config.MaxConn
	}
	return 0
}

// Mode returns mgo.Mode settings of Pool
func (p *Pool) Mode() readpref.Mode {
	return p.mode
}

// Config returns DBInfo of Pool
func (p *Pool) Config() *DBInfo {
	return p.config
}

// Close graceful shutdown conns and Pool status
func (p *Pool) Close() {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	p.available = false
	p.client.Disconnect(context.TODO())
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
