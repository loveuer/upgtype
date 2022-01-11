package upgtype

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

type StrSlice []string

func (s *StrSlice) Scan(val interface{}) error {

	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("db value not string, but %T", val)
	}

	if len(str) == 0 {
		return nil
	}

	if str[0] == '{' && str[len(str)-1] == '}' {
		str = str[1 : len(str)-1]
	}

	if len(str) == 0 {
		return nil
	}

	bs := make([]byte, 0)
	bss := make([][]byte, 0)

	for idx := 0; idx < len(str); idx++ {

		if str[idx] == ',' && str[idx-1] != 92 {

			if bs[0] != '"' || bs[len(bs)-1] != '"' {
				bs = append(bs, '"')
				bss = append(bss, append([]byte{'"'}, bs...))
			} else {
				bss = append(bss, bs)
			}

			bs = make([]byte, 0)
		} else {
			bs = append(bs, str[idx])
		}

	}

	if bs[0] != '"' || bs[len(bs)-1] != '"' {
		bs = append(bs, '"')
		bss = append(bss, append([]byte{'"'}, bs...))
	} else {
		bss = append(bss, bs)
	}

	strs := make([]string, 0, len(bss))
	for idx := range bss {

		var newStr string
		if err := json.Unmarshal(bss[idx], &newStr); err != nil {
			continue
		}

		strs = append(strs, newStr)
	}

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
