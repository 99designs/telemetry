package datadog

import (
	"bytes"
	"encoding/json"
	"sort"
)

type Points map[int64]float64

func (p *Points) UnmarshalJSON(b []byte) error {
	var tmp [][]interface{}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()

	err := dec.Decode(&tmp)
	if err != nil {
		return err
	}

	*p = map[int64]float64{}
	for _, kv := range tmp {
		key, err := kv[0].(json.Number).Int64()
		if err != nil {
			return err
		}
		(*p)[key], err = kv[1].(json.Number).Float64()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Points) MarshalJSON() ([]byte, error) {
	tmp := make([][]interface{}, len(p))
	i := 0
	for k, v := range p {
		tmp[i] = []interface{}{k, v}
		i++
	}

	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i][0].(int64) < tmp[j][0].(int64)
	})

	return json.Marshal(tmp)
}

type Series struct {
	Metric string   `json:"metric"`
	Points Points   `json:"points"`
	Type   string   `json:"type"`
	Tags   []string `json:"tags"`
	Host   string   `json:"host"`
}

func (s Series) Copy(suffix string) Series {
	newS := s
	newS.Points = Points{}
	newS.Metric += suffix
	return newS
}

type BatchSeries struct {
	Series []Series `json:"series"`
}
