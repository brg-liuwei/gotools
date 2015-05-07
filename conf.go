package gotools

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func DecodeJsonFile(filename string, schema interface{}) error {
	if f, err := os.Open(filename); err != nil {
		return err
	} else {
		decoder := json.NewDecoder(f)
		return decoder.Decode(schema)
	}
}

func ParseJsonFileToMap(filename string) (map[string]interface{}, error) {
	var m map[string]interface{}
	if f, err := os.Open(filename); err != nil {
		return nil, err
	} else if b, err := ioutil.ReadAll(f); err != nil {
		return nil, err
	} else if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	} else {
		return m, nil
	}
}

func ParseJsonFileToSlice(filename string) ([]interface{}, error) {
	var s []interface{}
	if f, err := os.Open(filename); err != nil {
		return nil, err
	} else if b, err := ioutil.ReadAll(f); err != nil {
		return nil, err
	} else if err = json.Unmarshal(b, &s); err != nil {
		return nil, err
	} else {
		return s, nil
	}
}
