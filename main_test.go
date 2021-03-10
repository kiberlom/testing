package main

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"testing"
)

var T = []string{
	"Field1 = 'fo' AND Field2 != 7 OR Field3 > 11.7",
	"Foo.Bar.X = 'hello'",
	"Bar.Alpha = 7",
	"Foo.Bar.Beta > 21 AND Alpha.Bar != 'hello'",
	"Alice.IsActive AND Bob.LastHash = 'ab5534b'",
	"Alice.Name ~ 'A.*` OR Bob.LastName !~ 'Bill.*`",
}

func TestParse(t *testing.T) {

	for _, v := range T {
		q := squirrel.Select("*").From("table")
		nq, err := Parser(v, q)
		if err != nil {
			fmt.Println(err)
			continue
		}
		s, arg, err := nq.ToSql()
		if err!=nil{
			fmt.Println("error")
		}

		fmt.Println(s, arg)
	}
}

