package gotools

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func DecodeJsonFile(filename string, schema interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	return decoder.Decode(schema)
}

func ParseJsonFileToMap(filename string) (map[string]interface{}, error) {
	var m map[string]interface{}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if b, err := ioutil.ReadAll(f); err != nil {
		return nil, err
	} else if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	} else {
		return m, nil
	}
}

func ParseJsonFileToSlice(filename string) ([]interface{}, error) {
	var s []interface{}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if b, err := ioutil.ReadAll(f); err != nil {
		return nil, err
	} else if err = json.Unmarshal(b, &s); err != nil {
		return nil, err
	} else {
		return s, nil
	}
}
