package mgopool

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
)

func GetCondition(new ...bson.M) (c Condition) {
	c.Bson = bson.M{}
	for _, m := range new {
		for k, v := range m {
			c.Bson[k] = v
		}
	}
	return
}

type Condition struct {
	Bson bson.M
}

func (c *Condition) Equal(key string, cond interface{}) bson.M {
	c.Bson[key] = bson.M{OpEqual: cond}
	return c.Bson
}

func (c *Condition) NotEqual(key string, cond interface{}) bson.M {
	c.Bson[key] = bson.M{OpNotEqual: cond}
	return c.Bson
}

func (c *Condition) Set(key string, cond interface{}) bson.M {
	c.Bson[key] = cond
	return c.Bson
}

func (c *Condition) Great(key string, cond interface{}) bson.M {
	c.Bson[key] = bson.M{OpGt: cond}
	return c.Bson
}

func (c *Condition) GreatEqual(key string, cond interface{}) bson.M {
	c.Bson[key] = bson.M{OpGte: cond}
	return c.Bson
}

func (c *Condition) Less(key string, cond interface{}) bson.M {
	c.Bson[key] = bson.M{OpLt: cond}
	return c.Bson
}

func (c *Condition) LessEqual(key string, cond interface{}) bson.M {
	c.Bson[key] = bson.M{OpLte: cond}
	return c.Bson
}

func (c *Condition) In(key string, cond interface{}) bson.M {
	c.Bson[key] = bson.M{OpIn: cond}
	return c.Bson
}

func (c *Condition) NotInFloat(key string, cond []float64) bson.M {
	size := len(cond)
	if size == 1 {
		return c.NotEqual(key, cond[0])
	}
	tmp := make([]bson.M, size)
	for i, k := range cond {
		tmp[i] = bson.M{key: bson.M{OpNotEqual: k}}
	}

	c.Bson[OpAnd] = tmp

	return c.Bson
}

func (c *Condition) NotInInteger(key string, cond []int) bson.M {
	size := len(cond)
	if size == 1 {
		return c.NotEqual(key, cond[0])
	}
	tmp := make([]bson.M, size)
	for i, k := range cond {
		tmp[i] = bson.M{key: bson.M{OpNotEqual: k}}
	}

	c.Bson[OpAnd] = tmp

	return c.Bson
}

func (c *Condition) NotInString(key string, cond []string) bson.M {
	size := len(cond)
	if size == 1 {
		return c.NotEqual(key, cond[0])
	}

	tmp := make([]bson.M, size)
	for i, k := range cond {
		tmp[i] = bson.M{key: bson.M{OpNotEqual: k}}
	}

	c.Bson[OpAnd] = tmp

	return c.Bson
}

func (c *Condition) NotIn(key string, cond interface{}) bson.M {
	c.Bson[OpNin] = cond

	return c.Bson
}

func (c *Condition) Regex(key string, cond interface{}) bson.M {

	c.Bson[key] = bson.M{OpRegex: cond}
	return c.Bson
}

func (c *Condition) Or(cond []bson.M) bson.M {
	c.Bson[OpOr] = cond
	return c.Bson
}

func (c *Condition) And(cond []bson.M) bson.M {
	c.Bson[OpAnd] = cond
	return c.Bson
}

func (c *Condition) Exists(key string) bson.M {
	c.Bson[key] = bson.M{OpExists: true}
	return c.Bson
}

func (c *Condition) NotExists(key string) bson.M {
	c.Bson[key] = bson.M{OpExists: false}
	return c.Bson
}

func (c *Condition) String() string {
	return fmt.Sprintf("%+v", c.Bson)
}
