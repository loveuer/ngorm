package ngorm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"testing"
)

var (
	client *Client
)

func testInit() {
	var (
		err error
	)

	logrus.SetLevel(logrus.DebugLevel)

	client, err = NewClient(context.TODO(), &Config{
		Endpoints:    []string{"10.220.10.19:9669"},
		Username:     "root",
		Password:     "123",
		DefaultSpace: "test_base",
		Logger:       nil,
	})

	if err != nil {
		log.Fatal("new client err:", err)
	}
}

func TestScanVertex2Any(t *testing.T) {
	testInit()

	var (
		a    any
		err  error
		ngql = "fetch prop on NAMES,ADDRESS '000164', '00031N6' yield NAMES.v as names, ADDRESS.v as address"
		//ngql = "fetch prop on * '000164', '00031N6' yield vertex as v"
	)

	if err = client.Raw(ngql).Scan(&a); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("result:\n%s\n", string(bs))
}

func TestScanVertex2Map(t *testing.T) {
	testInit()

	var (
		m    = make(map[string]any)
		err  error
		ngql = "fetch prop on NAMES,ADDRESS '000164', '00031N6' yield NAMES.v as names, ADDRESS.v as address"
		//ngql = "fetch prop on * '000164', '00031N6' yield vertex as v"
	)

	if err = client.Raw(ngql).Scan(&m); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("result:\n%s\n", string(bs))
}

func TestScanVertex2MapSlice(t *testing.T) {
	testInit()

	var (
		ms   = make([]map[string]any, 0)
		err  error
		ngql = "fetch prop on NAMES,ADDRESS '000164', '00031N6' yield NAMES.v as names, ADDRESS.v as address"
		//ngql = "fetch prop on * '000164', '00031N6' yield vertex as v"
	)

	if err = client.Raw(ngql).Scan(&ms); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(ms)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("result:\n%s\n", string(bs))
}

type Base struct {
	Id      string   `nebula:"id"`
	Names   []string `nebula:"names"`
	Address []string `nebula:"address"`
}

func TestScanVertex2Struct(t *testing.T) {
	testInit()

	var (
		v   = new(Base)
		err error
		//ngql = "fetch prop on NAMES,ADDRESS '000164', '00031N6' yield NAMES.v as names, ADDRESS.v as address"
		//ngql = "fetch prop on * '000164', '00031N6' yield vertex as v"
		ngql = "go 1 steps from 'H8ko' over contact yield $$.NAMES.v as names,$$.ADDRESS.v as address,  id($$) as id | limit 1"
	)

	if err = client.Raw(ngql).Scan(v); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(v)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("[test] scan struct result:\n%s\n", string(bs))
}

func TestScanVertex2StructSlice(t *testing.T) {
	testInit()

	var (
		vsptr = make([]*Base, 0)
		vs    = make([]Base, 0)
		err   error
		//ngql = "fetch prop on NAMES,ADDRESS '000164', '00031N6' yield NAMES.v as names, ADDRESS.v as address"
		//ngql = "fetch prop on * '000164', '00031N6' yield vertex as v"
		ngql = "go 1 steps from 'H8ko' over contact yield $$.NAMES.v as names,$$.ADDRESS.v as address,  id($$) as id | limit 3"
	)

	if err = client.Raw(ngql).Scan(&vsptr); err != nil {
		t.Error(err)
	}

	bsptr, err := json.Marshal(vsptr)
	if err != nil {
		t.Error(err)
	}

	if err = client.Raw(ngql).Scan(&vs); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(vs)
	if err != nil {
		t.Error(err)
	}

	if string(bs) != string(bsptr) {
		t.Error("not same")
	}

	fmt.Printf("result:\n%s\n", string(bs))
}

func TestScanString(t *testing.T) {
	testInit()

	var (
		s   = ""
		err error
		//ngql = "fetch prop on NAMES,ADDRESS '000164', '00031N6' yield NAMES.v as names, ADDRESS.v as address"
		//ngql = "fetch prop on * '000164', '00031N6' yield vertex as v"
		ngql = "go 1 steps from 'H8ko' over contact yield id($$) as id | limit 1"
	)

	if err = client.Raw(ngql).Scan(&s); err != nil {
		t.Error(err)
	}

	fmt.Printf("[test] result: %v\n", s)
}
