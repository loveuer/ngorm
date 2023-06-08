package ngorm

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGo(t *testing.T) {
	testInit()

	var vids any

	if err := client.GoFrom("4m6ziH3").
		Over("contact", EdgeTypeReverse).
		Tags("NAMES", "ADDRESS").
		Scan(&vids); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(vids)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("result:\n%s\n", string(bs))
}
