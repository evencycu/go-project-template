package mgopool

import (
	"fmt"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// replace preserved character in mongo
// '.' & '$' is preserved in mongo
var replacer = strings.NewReplacer(".", "", dollar, "")

type Pipeline struct {
	stages     []bson.M
	g          *group
	groupField map[string]string
}

func NewPipeline() *Pipeline {
	p := &Pipeline{}
	p.stages = []bson.M{}
	p.g = &group{
		aggrField: make(map[string]string),
	}
	p.groupField = make(map[string]string)
	return p
}

func (p *Pipeline) Match(cond bson.M) {
	condCpy := bson.M{}
	for f, c := range cond {
		condCpy[p.fieldRegulate(f)] = c
	}
	p.stages = append(p.stages, bson.M{OpMatch: condCpy})
}

func (p *Pipeline) Group(field ...string) *group {
	group_ := bson.M{}
	for _, f := range field {
		fName, _f := parseField(f)
		group_[fName] = _f
		p.groupField[f] = fName
	}
	p.g.groupField = bson.M{ObjId: group_}
	p.stages = append(p.stages, bson.M{OpGroup: p.g.groupField})
	return p.g
}

func (p *Pipeline) Sort(fields ...string) {
	sort := bson.D{}
	for _, f := range fields {
		order := 1
		if f == "" {
			continue
		}
		switch f[0] {
		case '-':
			order = -1
			f = f[1:]
		case '+':
			f = f[1:]
		}
		f = p.fieldRegulate(f)
		sort = append(sort, bson.E{Key: f, Value: order})
	}
	p.stages = append(p.stages, bson.M{OpSort: sort})
}

func (p *Pipeline) Skip(s int) {
	if s > 0 {
		p.stages = append(p.stages, bson.M{OpSkip: s})
	}
}

func (p *Pipeline) Limit(l int) {
	if l > 0 {
		p.stages = append(p.stages, bson.M{OpLimit: l})
	}
}

func (p *Pipeline) Count(fieldName ...string) {
	field_ := "count"
	if len(fieldName) > 0 {
		field_ = fieldName[0]
	}
	p.stages = append(p.stages, bson.M{OpCount: field_})
}

func (p *Pipeline) Unwind(field string) {
	p.stages = append(p.stages, bson.M{OpUnwind: field})
}

func (p *Pipeline) Out(collection string) {
	p.stages = append(p.stages, bson.M{OpOut: collection})
}

func (p *Pipeline) fieldRegulate(field string) string {
	if f, ok := p.groupField[field]; ok {
		return "_id." + f
	}
	if f, ok := p.g.aggrField[field]; ok {
		return f
	}
	return field
}

func (p *Pipeline) Append(stage bson.M) {
	p.stages = append(p.stages, stage)
}

func (p *Pipeline) Done() []bson.M {
	return p.stages
}

// String: the output string is for mongo shell use
func (p *Pipeline) String(pretty ...bool) string {
	if len(pretty) > 0 && pretty[0] {
		return printSliceBsonM(p.stages, true)
	}
	return printSliceBsonM(p.stages, false)
}

type group struct {
	groupField bson.M
	aggrField  map[string]string
}

func (g *group) Sum(fields ...string) *group {
	for _, f := range fields {
		fName, _f := parseField(f)
		g.groupField[fName] = bson.M{OpSum: _f}
		g.aggrField[f] = fName
	}
	return g
}

func (g *group) Max(fields ...string) *group {
	for _, f := range fields {
		fName, _f := parseField(f)
		g.groupField[fName] = bson.M{OpMax: _f}
		g.aggrField[f] = fName
	}
	return g
}

func (g *group) Min(fields ...string) *group {
	for _, f := range fields {
		fName, _f := parseField(f)
		g.groupField[fName] = bson.M{OpMin: _f}
		g.aggrField[f] = fName
	}
	return g
}

func (g *group) First(fields ...string) *group {
	for _, f := range fields {
		fName, _f := parseField(f)
		g.groupField[fName] = bson.M{OpFirst: _f}
		g.aggrField[f] = fName
	}
	return g
}

func (g *group) Last(fields ...string) *group {
	for _, f := range fields {
		fName, _f := parseField(f)
		g.groupField[fName] = bson.M{OpLast: _f}
		g.aggrField[f] = fName
	}
	return g
}

func (g *group) Aggregate(field string, aggregator bson.M) *group {
	fName, _ := parseField(field)
	g.groupField[fName] = aggregator
	g.aggrField[field] = fName
	return g
}

func print(v interface{}) string {
	var s string
	switch v.(type) {
	case string:
		s = fmt.Sprintf(`"%s"`, v)
	case bson.M:
		s = printBsonM(v.(bson.M))
	case bson.D:
		s = printBsonD(v.(bson.D))
	case []bson.M:
		s = printSliceBsonM(v.([]bson.M), false)
	case []string:
		s = printSliceString(v.([]string))
	default:
		s = fmt.Sprintf("%v", v)
	}
	return s
}

func parseField(field string) (string, interface{}) {
	var fName string
	if n, err := strconv.Atoi(field); err == nil {
		fName = "sum" + field
		return fName, n
	}
	fName = replacer.Replace(field)

	return fName, dollar + field
}

func printBsonM(m bson.M) string {
	bsonArr := []string{}
	for k, v := range m {
		bsonArr = append(bsonArr, fmt.Sprintf(`"%s":%s`, k, print(v)))
	}
	return "{" + strings.Join(bsonArr, ",") + "}"
}

func printBsonD(d bson.D) string {
	bsonArr := []string{}
	for _, e := range d {
		bsonArr = append(bsonArr, fmt.Sprintf(`"%s":%s`, e.Key, print(e.Value)))
	}
	return "{" + strings.Join(bsonArr, ",") + "}"
}

func printSliceBsonM(m []bson.M, pretty bool) string {
	br := ","
	if pretty {
		br = ",\n"
	}
	pStr := "["
	for _, bsonM := range m {
		pStr += printBsonM(bsonM) + br
	}

	pStr += "]"
	return pStr
}

func printSliceString(s []string) string {
	return `["` + strings.Join(s, `","`) + `"]`
}
