package ngorm

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	testInit()

	var vids any

	//if err := client.Fetch("02oP", "0KhNqj1").Tags("NAMES", "ADDRESS").Key("v").Scan(&vids); err != nil {
	if err := client.Fetch("4m6ziH3", "FmF", "zyW").Tags("NAMES", "ADDRESS").Key("v").Scan(&vids); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(vids)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("result:\n%s\n", string(bs))
}

func TestFetchToStruct(t *testing.T) {
	testInit()

	type FocusListRes struct {
		Uuid             string    `nebula:"VertexID" json:"uuid"`
		Tid              string    `json:"tid"`
		Avatar           string    `json:"avatar"`
		Platform         string    `json:"platform"`
		OperationType    string    `json:"operation_type"`
		Rid              string    `json:"rid"`
		Email            []string  `nebula:"EMAIL" json:"email"`
		RelationCountDST []string  `nebula:"RELATION_COUNT_DST" json:"relation_count_dst"`
		Region           []string  `nebula:"ADDRESS" json:"region"`
		Photo            []string  `nebula:"PHOTO" json:"photo"`
		Phone            []string  `nebula:"PHONE" json:"phone"`
		Names            []string  `nebula:"NAMES" json:"names"`
		ContactCount     int       `json:"contact_count"`
		BeContactCount   int       `json:"becontact_count"`
		Time             time.Time `json:"time"`
		ChildrenCount    int       `nebula:"CHILDREN_COUNT" json:"children_count"`
	}

	var (
		datas = make([]*FocusListRes, 0)
	)

	if err := client.Fetch("4m6ziH3", "FmF", "zyW").Tags().Key("v").Scan(&datas); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(datas)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("result:\n%s\n", string(bs))
}

func TestFetch2(t *testing.T) {
	testInit()

	type Result struct {
		Uuid             string    `nebula:"VertexID" json:"uuid"`
		Tid              string    `json:"tid"`
		Avatar           string    `json:"avatar"`
		Platform         string    `json:"platform"`
		OperationType    string    `json:"operation_type"`
		Rid              string    `json:"rid"`
		Email            []string  `nebula:"EMAIL" json:"email"`
		RelationCountDST []string  `nebula:"RELATION_COUNT_DST" json:"relation_count_dst"`
		Region           []string  `nebula:"ADDRESS" json:"region"`
		Photo            []string  `nebula:"PHOTO" json:"photo"`
		Phone            []string  `nebula:"PHONE" json:"phone"`
		Names            []string  `nebula:"NAMES" json:"names"`
		ContactCount     int       `json:"contact_count"`
		BeContactCount   int       `json:"becontact_count"`
		Time             time.Time `json:"time"`
	}

	var (
		ids   = []string{"4m6ziH3", "FmF", "zyW", "1Ghobo"}
		datas = make([]*Result, 0)
		err   error
	)

	if err = client.Fetch(ids...).Tags().Key("v").Scan(&datas); err != nil {
		t.Error(err)
	}

	for _, item := range datas {
		t.Logf("%s: %+v", item.Uuid, *item)
	}
}
