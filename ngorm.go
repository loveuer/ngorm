package ngorm

import (
	"context"
	"errors"
	"sync"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type Service struct {
	Addr string `mapstructure:"addr"`
	Port int    `mapstructure:"port"`
}

type Config struct {
	Ctx      context.Context
	Servers  []Service
	Username string
	Password string
	PoolSize int
	LogLevel LogLevel
}

type NGDB struct {
	sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
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
			// todo ErrorType
			return nil, errors.New("get session timeout")
		case session := <-db.sessChan:
			return session, nil
		}
	}

	return <-db.sessChan, nil
}

// Deprecated
func (db *NGDB) release(session *nebula.Session) {
	session.Release()
}

func (db *NGDB) initSess(size int) {
	sessionsPool := make([]*nebula.Session, 0, size)
	for idx := 0; idx < size; idx++ {
		if sess, err := db.newSession(); err == nil {
			sessionsPool = append(sessionsPool, sess)
		}
	}

	if len(sessionsPool) == 0 {
		log.Panic("session pool length == 0")
	}

	log.Infof("inited [%d] nebula sessions", len(sessionsPool))

	go func() {
		for {
			select {
			case <-db.ctx.Done():
				for idx := range sessionsPool {
					sessionsPool[idx].Release()
				}
				return
			default:
				for idx := range sessionsPool {
					db.sessChan <- sessionsPool[idx]
				}
			}
		}
	}()
}

func (db *NGDB) newSession() (*nebula.Session, error) {
	sess, err := db.pool.GetSession(db.username, db.password)
	if err != nil {
		log.Warnf("ngdb get session err: %v", err)
		return sess, err
	}

	return sess, nil
}

func (db *NGDB) prepare() (sess *nebula.Session, err error) {
	sess, err = db.getSession()
	if err != nil {
		return sess, err
	}

	if sess == nil {
		// todo ErrorType
		return sess, errors.New("invalid session")
	}

	_, err = sess.Execute("USE " + db.space)
	if err != nil {
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

	if cfg.Ctx == nil {
		db.ctx, db.cancel = context.WithCancel(context.Background())
	} else {
		db.ctx, db.cancel = context.WithCancel(cfg.Ctx)
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
	db.username = cfg.Username
	db.password = cfg.Password

	db.pool, err = nebula.NewConnectionPool(hostList, defaultPoolConfig, nglog)
	if err != nil {
		log.Errorf("can't constructs new connection pool, err: %v", err)
		return db, err
	}

	db.sessChan = make(chan *nebula.Session)

	db.initSess(cfg.PoolSize)

	return db, nil
}

func (nd *NGDB) Close() {
	nd.cancel()
}
