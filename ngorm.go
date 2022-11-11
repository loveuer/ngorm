package ngorm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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
	pool1    *nebula.ConnectionPool
	pool2    *nebula.ConnectionPool
	username string
	password string
	timeout  int
}

func (db *NGDB) getSession(ctx context.Context) (*nebula.Session, error) {
	var (
		err     error
		ch      = make(chan *nebula.Session)
		fctx, _ = context.WithTimeout(ctx, 30*time.Second)
	)

	go func() {
		var (
			sess  *nebula.Session
			count = 0
		)

		for {
			if sess, err = db.newSession(count % 2); err != nil {
				count++
				logrus.Debugf("nubela new session err: %v", err)
				time.Sleep(200 * time.Millisecond)
				continue
			}

			break
		}

		ch <- sess
	}()

	select {
	case newSess := <-ch:
		return newSess, nil
	case <-fctx.Done():
		return nil, errors.New("acquire new nebula session timeout")
	}
}

func (db *NGDB) newSession(which int) (*nebula.Session, error) {
	switch which + 1 {
	case 1:
		if db.pool1 == nil {
			return nil, errors.New("nebula connection pool 1 is nil")
		}

		return db.pool1.GetSession(db.username, db.password)
	case 2:
		if db.pool2 == nil {
			return nil, errors.New("nebula connection pool 2 is nil")
		}

		return db.pool2.GetSession(db.username, db.password)
	default:
		return nil, fmt.Errorf("err pool idx[%d]", which+1)
	}
}

func (db *NGDB) prepare(ctx context.Context) (*nebula.Session, error) {
	var (
		err  error
		sess *nebula.Session
	)

	if sess, err = db.getSession(ctx); err != nil {
		return nil, err
	}

	if _, err = sess.Execute("USE " + db.space); err != nil {
		return nil, err
	}

	return sess, nil
}

func NewNGDB(space string, config ...Config) (*NGDB, error) {
	var (
		err      error
		err1     error
		err2     error
		cfg      Config
		db       = new(NGDB)
		hostList = make([]nebula.HostAddress, 0)
		//nglog             = logrus.New()
		nglog             = nebula.DefaultLogger{}
		defaultPoolConfig = nebula.GetDefaultConf()
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

	defaultPoolConfig.MaxConnPoolSize = 2 * cfg.PoolSize
	defaultPoolConfig.MinConnPoolSize = cfg.PoolSize

	db.space = space
	db.username = cfg.Username
	db.password = cfg.Password

	if db.pool1, err1 = nebula.NewConnectionPool(hostList, defaultPoolConfig, nglog); err1 != nil {
		log.Warnf("init nebula connection pool 1 err: %v", err1)
	}

	if db.pool2, err2 = nebula.NewConnectionPool(hostList, defaultPoolConfig, nglog); err2 != nil {
		log.Warnf("init nebula connection pool 2 err: %v", err1)
	}

	if err1 != nil && err2 != nil {
		log.Errorf("can't constructs new connection pools, err: %v", err)
		return db, err
	}

	log.Infof("inited nebula connection pools(size: %d)", defaultPoolConfig.MinConnPoolSize)

	return db, nil
}

func (nd *NGDB) Close() {
	nd.cancel()
}
