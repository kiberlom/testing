package main

import (
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"regexp"
	"strconv"
	"strings"
)

// столбец - параметр
type One struct {
	Name string
	Zn   interface{}
	De   string
}

// операторы
var raz = []string{
	`!=`,
	`=`,
	`<=`,
	`>=`,
	`>`,
	`<`,
	`!~*`,
	`~*`,
	`!~`,
	`~`,
}

// зарезервированные имена
var reserved = []string{
	"LIMIT",
	"ORDER BY",
	"BY",
	"WHERE",
	"SELECT",
	"FROM",
}

//check reserved
func checkRes(q string) error {

	for _, r := range reserved {

		nr := fmt.Sprintf(`\b(%s)\b`, r)
		matched, err := regexp.MatchString(nr, q)

		if err != nil {
			return errors.New("error: name reserved, reg error")
		}

		if matched {
			return errors.New("error: name reserved")
		}

	}

	return nil
}

// проверка и приведение типов
func typeCheck(a interface{}) interface{} {

	if f, err := strconv.ParseInt(a.(string), 0, 64); err == nil {
		return f
	}
	if f, err := strconv.ParseFloat(a.(string), 64); err == nil {
		return f
	}
	if "true" == a {
		return true
	}
	return a.(string)

}

// проверка имени
func checkName(s string) error {

	// check symbol
	matched, err := regexp.MatchString(`^[A-z0-9\.]+$`, s)
	if err != nil || !matched {
		return errors.New("error: name not allowed symbol")
	}

	//check reserved
	for _, r := range reserved {
		if strings.ToLower(s) == r {
			return errors.New("error: name reserved")
		}
	}

	return nil

}

// разбиваем пару на значение и параметр
func NamePar(s string, f func(string) error) (o One, err error) {

	o = One{}
	err = nil

	for _, v := range raz {
		r := strings.Split(s, v)

		if len(r) > 1 {

			name := strings.TrimSpace(r[0])
			if err := f(name); err != nil {
				return o, err
			}
			o.Name = name
			o.De = v
			o.Zn = typeCheck(strings.TrimSpace(r[1]))
			return
		}
	}

	if err := f(s); err != nil {
		return o, err
	}

	o.Name = s
	o.De = "="
	o.Zn = typeCheck("true")
	return

}

// разбиваем на параметры
func parsOR(s string) (v []One, err error) {

	err = nil

	a := regexp.MustCompile(`\b(OR)\b`)
	m := a.Split(s, -1)

	for _, u := range m {
		o, er := NamePar(u, checkName)
		if er != nil {
			err = er
			return
		}
		v = append(v, o)
	}

	return

}

func parsAND(query string) (arAnd []string) {

	a := regexp.MustCompile(`\b(AND)\b`)

	for _, v := range a.Split(query, -1) {
		s := strings.TrimSpace(v)
		arAnd = append(arAnd, s)
	}

	return
}

func createOR(m []One) squirrel.Or {
	w := squirrel.Or{}
	for _, v := range m {
		switch v.De {
		case "=":
			w = append(w, squirrel.Eq{v.Name: v.Zn})
		case ">":
			w = append(w, squirrel.Gt{v.Name: v.Zn})
		case "!=":
			w = append(w, squirrel.NotEq{v.Name: v.Zn})
		}
	}
	return w
}

func createAND(v One, qb *squirrel.SelectBuilder) squirrel.SelectBuilder {

	r := *qb
	switch v.De {
	case "=":
		r = r.Where(squirrel.Eq{v.Name: v.Zn})
	case ">":
		r = r.Where(squirrel.Gt{v.Name: v.Zn})
	case "!=":
		r = r.Where(squirrel.NotEq{v.Name: v.Zn})
	}

	return r

}

// разбиваем на массив пар
func Parser(query string, qb squirrel.SelectBuilder) (*squirrel.SelectBuilder, error) {

	// проверка на зарезервированные
	if err := checkRes(query); err != nil {
		return nil, err
	}

	// получаем массив с разделителем AND
	ar := parsAND(query)

	// передераем
	for _, h := range ar {
		// парсим или
		mO, err := parsOR(h)
		if err != nil {
			return &qb, err
		}

		// если нет условия
		if len(mO) == 1 {
			qb = createAND(mO[0], &qb)
			continue
		}

		qb = qb.Where(createOR(mO))

	}

	return &qb, nil

}

func main() {

	q := squirrel.Select("*").From("table")
	nq, err := Parser("Field1 = 'fo' AND Field2 != 7 OR Field3 > 11.7", q)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(nq.ToSql())

}
