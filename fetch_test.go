package ngorm

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestFetch(t *testing.T) {
	testInit()

	var vids any

	if err := client.Fetch("02oP", "0KhNqj1").Tags("NAMES", "ADDRESS").Key("v").Scan(&vids); err != nil {
		t.Error(err)
	}

	bs, err := json.Marshal(vids)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("result:\n%s\n", string(bs))
}
