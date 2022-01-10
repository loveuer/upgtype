package upgtype

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
)

type StrSlice []string

func (s *StrSlice) Scan(val interface{}) error {

	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("db value not string, but %T", val)
	}

	str = strings.TrimRight(strings.TrimLeft(str, "{"), "}")

	if len(str) == 0 {
		return nil
	}

	strs := make([]string, 0)
	one := make([]byte, 0)
	quotation := 0

	for idx := range str {
		if str[idx] == '"' {
			quotation++
			one = append(one, str[idx])
		} else if str[idx] == ',' {
			if quotation%2 != 0 {
				one = append(one, str[idx])
				continue
			}

			strs = append(strs, string(dropQuotedAndBS(one)))
			one = make([]byte, 0)

		} else {
			one = append(one, str[idx])
		}
	}

	strs = append(strs, string(dropQuotedAndBS(one)))

	reflect.ValueOf(s).Elem().Set(reflect.ValueOf(strs))

	return nil
}

func dropQuotedAndBS(bs []byte) []byte {
	if bs[0] == '"' && bs[len(bs)-1] == '"' {
		bs = bs[1 : len(bs)-1]
	}

	cleanBS := make([]byte, 0)
	ii := 0
	for ii < len(bs)-1 {
		if bs[ii] == 92 && bs[ii+1] == '"' {
			ii++
			continue
		}

		cleanBS = append(cleanBS, bs[ii])
		ii++
	}

	return append(cleanBS, bs[len(bs)-1])
}

func (s StrSlice) Value() (driver.Value, error) {
	if s == nil {
		return "{}", nil
	}

	if n := len(s); n > 0 {
		b := make([]byte, 1, 1+3*n)
		b[0] = '{'

		b = appendArrayQuotedBytes(b, []byte(s[0]))
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = appendArrayQuotedBytes(b, []byte(s[i]))
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

func appendArrayQuotedBytes(b, v []byte) []byte {
	b = append(b, '"')
	for {
		i := bytes.IndexAny(v, `"\`)
		if i < 0 {
			b = append(b, v...)
			break
		}
		if i > 0 {
			b = append(b, v[:i]...)
		}
		b = append(b, '\\', v[i])
		v = v[i+1:]
	}
	return append(b, '"')
}
