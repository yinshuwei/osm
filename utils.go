package osm

import (
	"database/sql"
	"log"
	"reflect"
	"strings"
	"time"
)

const (
	formatDate     = "2006-01-02"
	formatDateTime = "2006-01-02 15:04:05"
)

// format time string
func timeFormat(t time.Time, format string) string {
	return t.Format(format)
}

// // camel string, xx_yy to XxYy
// func toGoName(s string) string {
// 	data := make([]byte, 0, len(s))
// 	j := false
// 	k := false
// 	num := len(s) - 1
// 	for i := 0; i <= num; i++ {
// 		d := s[i]
// 		if k == false && d >= 'A' && d <= 'Z' {
// 			k = true
// 		}
// 		if d >= 'a' && d <= 'z' && (j || k == false) {
// 			d = d - 32
// 			j = false
// 			k = true
// 		}
// 		if k && d == '_' && num > i /**&& s[i+1] >= 'a' && s[i+1] <= 'z'**/ {
// 			j = true
// 			continue
// 		}
// 		data = append(data, d)
// 	}
// 	return string(data[:len(data)])
// }

// func toGoName(name string) string {
// 	names := strings.Split(name, "_")

// 	newName := ""
// 	for _, ntemp := range names {
// 		newName += strings.ToUpper(ntemp[0:1]) + strings.ToLower(ntemp[1:len(ntemp)])
// 	}
// 	return newName
// }

// camel string, xx_yy to XxYy, 特列字符 ID
func toGoName(name string) string {
	num := len(name)
	data := make([]byte, len(name))
	j := 0
	k := true
	for i := 0; i < num; i++ {
		d := name[i]
		if d == '_' {
			k = true
			if j >= 2 && data[j-2] == 'I' && data[j-1] == 'd' {
				data[j-1] = 'D'
			}
		} else {
			if k {
				if d >= 'a' && d <= 'z' {
					d = d - 32
				}
			} else {
				if d >= 'A' && d <= 'Z' {
					d = d + 32
				}
			}
			data[j] = d
			j++
			k = false
		}
	}
	if j > 1 && data[j-1] == 'd' && data[j-2] == 'I' {
		data[j-1] = 'D'
	}
	return string(data[:j])
}

// scanRow 从sql.Rows中读一行数据
func scanRow(
	rows *sql.Rows,
	isPtrs []bool,
	elementTypes []reflect.Type,
	values []reflect.Value,
) error {
	lenContainers := len(isPtrs)
	var refs []interface{}
	isTimes := make([]bool, lenContainers)
	for i, isPtr := range isPtrs {
		if isPtr {
			if elementTypes[i].Elem().String() == "time.Time" {
				isTimes[i] = true
				refs = append(refs, new(string))
			} else {
				values[i].Set(reflect.New(elementTypes[i].Elem()))
				refs = append(refs, values[i].Interface())
			}
		} else {
			if elementTypes[i].String() == "time.Time" {
				isTimes[i] = true
				refs = append(refs, new(string))
			} else {
				refs = append(refs, values[i].Addr().Interface())
			}
		}
	}

	err := rows.Scan(refs...)
	if err != nil {
		return err
	}

	for i, isPtr := range isPtrs {
		if isTimes[i] {
			strPtr, ok := refs[i].(*string)
			if ok && strPtr != nil {
				str := *strPtr
				if str != "" {
					var t time.Time
					if len(str) >= 19 {
						if strings.Contains(str, "T") {
							t, err = time.Parse(time.RFC3339Nano, str)
						} else {
							str = str[:19]
							t, err = time.Parse(formatDateTime, str)
						}
					} else if len(str) >= 10 {
						str = str[:10]
						t, err = time.Parse(formatDate, str)
					}
					if err == nil {
						t = t.Local()
						if isPtr {
							values[i].Set(reflect.ValueOf(&t))
						} else {
							values[i].Set(reflect.ValueOf(t))
						}
					} else {
						log.Println(err)
					}
				}
			}
		}
	}
	return nil
}
