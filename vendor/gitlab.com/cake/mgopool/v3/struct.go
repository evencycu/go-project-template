package mgopool

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson"

)

const (
	dotChar       = "."
	escapeDotChar = "|"
)

type EscapeString string

func (e EscapeString) GetBSON() (interface{}, error) {
	return e.escape(), nil
}

func (e *EscapeString) SetBSON(raw bson.Raw) error {
	var escapeString string
	err := bson.Unmarshal(raw, &escapeString)
	if err != nil {
		return err
	}

	e.unescape(escapeString)

	return nil
}

func (e EscapeString) escape() string {
	return strings.Replace(string(e), dotChar, escapeDotChar, -1)
}

func (e *EscapeString) unescape(key string) {
	*e = EscapeString(strings.Replace(key, escapeDotChar, dotChar, -1))
}

func (e EscapeString) String() string {
	return strings.Replace(string(e), escapeDotChar, dotChar, -1)
}

type EscapeM map[string]interface{}

func (e EscapeM) GetBSON() (interface{}, error) {
	result := bson.M{}
	for k, v := range e {
		result[strings.Replace(k, dotChar, escapeDotChar, -1)] = v
	}
	return result, nil
}

func (e *EscapeM) SetBSON(raw bson.Raw) error {
	escapeM := make(bson.M)
	err := bson.Unmarshal(raw, &escapeM)
	if err != nil {
		return err
	}

	*e = make(EscapeM)
	for k, v := range escapeM {
		e.Set(strings.Replace(k, escapeDotChar, dotChar, -1), v)
	}
	escapeM = nil
	return nil
}

func (e EscapeM) Set(key string, value interface{}) {
	e[key] = value
}
