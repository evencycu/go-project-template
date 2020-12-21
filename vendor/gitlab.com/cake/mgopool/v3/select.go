package mgopool

import (
	"go.mongodb.org/mongo-driver/bson"
)

func SelectHiddenObjectID() *Select {
	s := &Select{}
	s.HideObjectID()
	return s
}

func SelectObjectID() *Select {
	s := &Select{}
	return s
}

func GetSelect() *Select {
	s := &Select{}
	s.HideObjectID()
	return s
}

type Select struct {
	Bson bson.M
	hide []string
	show []string
}

func (s *Select) HideObjectID() *Select {
	s.Hide("_id")
	return s
}

func (s *Select) Hide(h ...string) *Select {
	if s.hide == nil {
		s.hide = []string{}
	}
	s.hide = append(s.hide, h...)
	s.reload()
	return s
}

func (s *Select) Show(h ...string) *Select {
	if s.show == nil {
		s.show = []string{}
	}
	s.show = append(s.show, h...)
	s.reload()
	return s
}

func (s *Select) reload() {
	result := bson.M{}

	for _, k := range s.hide {
		result[k] = 0
	}
	for _, k := range s.show {
		result[k] = 1
	}
	s.Bson = result
}

func (s *Select) Get() bson.M {
	return s.Bson
}
