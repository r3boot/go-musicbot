package mpdclient

import (
	"encoding/json"
	"fmt"
)

func (a Artists) ToJSON() ([]byte, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("Artists.ToJSON json.Marshal: %v", err)
	}

	return data, nil
}

func (a Artists) Has(name string) bool {
	for _, entry := range a {
		if entry == name {
			return true
		}
	}

	return false
}

func (a Artists) Len() int           { return len(a) }
func (a Artists) Less(i, j int) bool { return a[i] < a[j] }
func (a Artists) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
