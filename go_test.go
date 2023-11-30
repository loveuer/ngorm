package ngorm

import (
	"encoding/json"
	"testing"
)

func TestGo(t *testing.T) {
	testInit()

	type SonarNebula struct {
		Uuid               string   `nebula:"VertexID"`
		Address            []string `nebula:"ADDRESS"`
		RelationCountDST   []string `nebula:"RELATION_COUNT_DST"`
		Platform           []string `nebula:"PROFILE_TAG"`
		Email              []string `nebula:"EMAIL"`
		Phone              []string `nebula:"PHONE"`
		Names              []string `nebula:"NAMES"`
		RelationCount      int64    `nebula:"RELATION_COUNT"`
		RelationCountDst   []string `nebula:"RELATION_COUNT_DST"`
		RelationCountryDst []string `nebula:"RELATION_COUNTRY_DST"`
		Source             []string `nebula:"SOURCE_TAG"`
		MessageCount       int64    `nebula:"MSG_COUNT"`
	}

	var (
		vids = make([]*SonarNebula, 0)
	)

	if err := client.GoFrom("0007").
		Over("contact", EdgeTypeForward).
		Limit(10000).
		//Tags("NAMES", "ADDRESS").
		Scan(&vids); err != nil {
		t.Error(err)
	}

	for idx, vid := range vids {
		bs, err := json.Marshal(vid)
		if err != nil {
			t.Error(err)
			return
		}

		t.Logf("idx=%d result:%s\n", idx, string(bs))
	}
}
