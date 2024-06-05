# Features

- 更好的数据绑定
- 更好的连接处理和重连机制
- 兼容 3.0，3.3 版本（已测试）

# Installation

```bash
go get github.com/loveuer/ngorm/v2
```

# Usage

- new client
```go
import (
    "github.com/loveuer/ngorm/v2"
)

client, err := ngorm.NewClient(gctx, &ngorm.Config{
    Endpoints:    []string{},
    Username:     "root",
    Password:     "****",
    DefaultSpace: "xxxx",
    Logger:       nil,
})

if err != nil {
    logrus.Panic("init ngorm client err:", err)
}
```

- fetch
```go
type Result struct {
    Id      string   `json:"id" nebula:"VertexID"`
    Names   []string `json:"names" nebula:"NAMES"`
    Address []string `json:"address" nebula:"ADDRESS"`
}

var (
    result = new(Result)
)

if err := client.Fetch(uuid).
	Model(&Result{}).
	Tags("NAMES", "ADDRESS").
	Key("v").
	Scan(result); err != nil {
		
	panic(err)
}
```

- go
```go
type Result struct {
    Id      string   `json:"id" nebula:"VertexID"`
    Names   []string `json:"names" nebula:"NAMES"`
    Address []string `json:"address" nebula:"ADDRESS"`
}

var (
    results = make([]*Result, 0)
)

if err := client.GoFrom(uuid).
    Model(&Result{}).
    Over("contact", EdgeTypeReverse).
    Tags("NAMES", "ADDRESS").
    Scan(&results); err != nil {
		
    panic(err)		
}
```

- raw
```go
type Result struct {
    Id      string   `json:"id" nebula:"VertexID"`
    Names   []string `json:"names" nebula:"names"`
    Address []string `json:"address" nebula:"address"`
}

var (
    ngql = "fetch prop on NAMES,ADDRESS 'uuid-1', 'uuid-2' yield id(vertex) as id, NAMES.v as names, ADDRESS.v as address"
    results = make([]*Result, 0)
    result_any any
    result_map = make(map[string]any)
    result_map_slice = make([]map[string]any)
)

if err := client.Raw(ngql).Scan(&result); err != nil {
	panic(err)
}

if err := client.Raw(ngql).Scan(&result_any); err != nil {
    panic(err)
}

if err := client.Raw(ngql).Scan(&result_map); err != nil {
    panic(err)
}

if err := client.Raw(ngql).Scan(&result_slice); err != nil {
    panic(err)
}
```