package ngorm

import (
	"errors"
	"sync"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v2"
)

type Service struct {
	Addr string `mapstructure:"addr"`
	Port int    `mapstructure:"port"`
}

type Config struct {
	Servers  []Service
	Username string
	Password string
	PoolSize int
	LogLevel LogLevel
}

type NGDB struct {
	sync.Mutex
	space    string
	pool     *nebula.ConnectionPool
	username string
	password string
	sessChan chan *nebula.Session
}

func (db *NGDB) getSession(duration ...time.Duration) (*nebula.Session, error) {
	if len(duration) > 0 {
		select {
		case <-time.After(duration[0]):
			return nil, errors.New("get session timeout")
		case session := <-db.sessChan:
			return session, nil
		}
	}

	return <-db.sessChan, nil
}

func (db *NGDB) release(session *nebula.Session) {
	session.Release()
	go func() {
		for {
			if db.newSession(true) {
				break
			}
		}
	}()
}

func (db *NGDB) initSess(size int) {
	for idx := 0; idx < size; idx++ {
		db.newSession(true)
	}
}

func (db *NGDB) newSession(must bool) bool {
	sess, err := db.pool.GetSession(db.username, db.password)
	if err != nil {
		if must {
			log.Fatalf("pool get session err: %v", err)
		}
		log.Errorf("ngdb get session err: %v", err)
		return false
	}

	db.sessChan <- sess

	return true
}

func (db *NGDB) prepare() (sess *nebula.Session, err error) {
	sess, err = db.getSession()
	if err != nil {
		return sess, err
	}

	if sess == nil {
		return sess, errors.New("invalid session")
	}

	_, err = sess.Execute("USE " + db.space)
	if err != nil {
		db.release(sess)
		return nil, err
	}

	return sess, nil
}


var (
	err   error
	nglog = nebula.DefaultLogger{}
)

func NewNGDB(space string, config ...Config) (*NGDB, error) {
	var (
		cfg      Config
		hostList = make([]nebula.HostAddress, 0)
		db       = new(NGDB)
	)

	if len(config) > 0 {
		cfg = config[0]
	}

	if len(cfg.Servers) == 0 {
		cfg.Servers = []Service{{Addr: "127.0.0.1", Port: 9669}}
	}

	if cfg.Username == "" {
		cfg.Username = "user"
	}

	if cfg.PoolSize == 0 {
		cfg.PoolSize = 10
	}

	SetLogLevel(cfg.LogLevel)

	for _, s := range cfg.Servers {
		hostAddr := nebula.HostAddress{
			Host: s.Addr,
			Port: s.Port,
		}
		hostList = append(hostList, hostAddr)
	}

	defaultPoolConfig := nebula.GetDefaultConf()
	defaultPoolConfig.MaxConnPoolSize = cfg.PoolSize
	defaultPoolConfig.MinConnPoolSize = cfg.PoolSize
	db.space = space

	db.pool, err = nebula.NewConnectionPool(hostList, defaultPoolConfig, nglog)
	if err != nil {
		log.Fatalf("can't constructs new connection pool, err: %v", err)
	}

	db.sessChan = make(chan *nebula.Session, cfg.PoolSize)
	db.username = cfg.Username
	db.password = cfg.Password

	db.initSess(cfg.PoolSize)

	return db, nil
}
