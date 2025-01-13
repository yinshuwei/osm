package osm

import (
	"database/sql"
	"reflect"
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
// func toGoName(name string) string {
// 	num := len(name)
// 	data := make([]byte, len(name))
// 	j := 0
// 	k := true
// 	for i := 0; i < num; i++ {
// 		d := name[i]
// 		if d == '_' {
// 			k = true
// 			if j >= 2 && data[j-2] == 'I' && data[j-1] == 'd' {
// 				data[j-1] = 'D'
// 			}
// 		} else {
// 			if k {
// 				if d >= 'a' && d <= 'z' {
// 					d = d - 32
// 				}
// 			} else {
// 				if d >= 'A' && d <= 'Z' {
// 					d = d + 32
// 				}
// 			}
// 			data[j] = d
// 			j++
// 			k = false
// 		}
// 	}
// 	if j > 1 && data[j-1] == 'd' && data[j-2] == 'I' {
// 		data[j-1] = 'D'
// 	}
// 	return string(data[:j])
// }

var commonInitialisms = map[string][]byte{
	"Acl":   []byte("ACL"),
	"Api":   []byte("API"),
	"Ascii": []byte("ASCII"),
	"Cpu":   []byte("CPU"),
	"Css":   []byte("CSS"),
	"Dns":   []byte("DNS"),
	"Eof":   []byte("EOF"),
	"Guid":  []byte("GUID"),
	"Html":  []byte("HTML"),
	"Http":  []byte("HTTP"),
	"Https": []byte("HTTPS"),
	"Id":    []byte("ID"),
	"Ip":    []byte("IP"),
	"Json":  []byte("JSON"),
	"Lhs":   []byte("LHS"),
	"Qps":   []byte("QPS"),
	"Ram":   []byte("RAM"),
	"Rhs":   []byte("RHS"),
	"Rpc":   []byte("RPC"),
	"Sla":   []byte("SLA"),
	"Smtp":  []byte("SMTP"),
	"Sql":   []byte("SQL"),
	"Ssh":   []byte("SSH"),
	"Tcp":   []byte("TCP"),
	"Tls":   []byte("TLS"),
	"Ttl":   []byte("TTL"),
	"Udp":   []byte("UDP"),
	"Ui":    []byte("UI"),
	"Uid":   []byte("UID"),
	"Uuid":  []byte("UUID"),
	"Uri":   []byte("URI"),
	"Url":   []byte("URL"),
	"Utf8":  []byte("UTF8"),
	"Vm":    []byte("VM"),
	"Xml":   []byte("XML"),
	"Xmpp":  []byte("XMPP"),
	"Xsrf":  []byte("XSRF"),
	"Xss":   []byte("XSS"),
}

// camel string, xx_yy to XxYy, 两种,一种为特殊片段
func toGoNames(name string) (string, string) {
	num := len(name)
	data := make([]byte, num)
	dataSpecial := make([]byte, num)
	point := 0
	isFirst := true
	firstPoint := 0

	for i := 0; i < num; i++ {
		d := name[i]
		if d == '_' {
			word, ok := commonInitialisms[string(data[firstPoint:point])]
			if ok {
				for j, b := range word {
					dataSpecial[firstPoint+j] = b
				}
			}
			isFirst = true
			firstPoint = point
		} else {
			if isFirst {
				if d >= 'a' && d <= 'z' {
					d = d - 32
				}
			} else {
				if d >= 'A' && d <= 'Z' {
					d = d + 32
				}
			}
			data[point] = d
			dataSpecial[point] = d
			point++
			isFirst = false
		}
	}
	word, ok := commonInitialisms[string(data[firstPoint:point])]
	if ok {
		for j, b := range word {
			dataSpecial[firstPoint+j] = b
		}
	}
	return string(data[:point]), string(dataSpecial[:point])
}

func findFiled(tagMap, nameMap map[string]*structFieldInfo, name string) *structFieldInfo {
	v, ok := tagMap[name]
	if ok {
		return v
	}

	a, b := toGoNames(name)
	t, ok := nameMap[a]
	if ok {
		return t
	}
	t, ok = nameMap[b]
	if ok {
		return t
	}
	return nil
}

// scanRow 从sql.Rows中读一行数据
func (o *osmBase) scanRow(
	logPrefix string,
	rows *sql.Rows,
	fields []*structFieldInfo,
	values []reflect.Value,
) error {
	lenContainers := len(fields)
	srcs := make([]*interface{}, lenContainers)
	refs := make([]interface{}, lenContainers)
	types := make([]reflect.Type, lenContainers)
	for i, field := range fields {
		ref := new(interface{})
		refs[i] = ref
		srcs[i] = ref

		if field == nil {
			continue
		}
		if field.isPtr {
			types[i] = (*(field.t)).Elem()
		} else {
			types[i] = *(field.t)
		}
	}

	err := rows.Scan(refs...)
	if err != nil {
		return err
	}

	for i, src := range srcs {
		if src == nil {
			continue
		}
		field := fields[i]
		if field == nil {
			continue
		}
		o.convertAssign(logPrefix, values[i], *src, field.isPtr, types[i])
	}
	return nil
}

func isValueKind(kind reflect.Kind) bool {
	return kind == reflect.Bool ||
		kind == reflect.Int ||
		kind == reflect.Int8 ||
		kind == reflect.Int16 ||
		kind == reflect.Int32 ||
		kind == reflect.Int64 ||
		kind == reflect.Uint ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64 ||
		kind == reflect.Uintptr ||
		kind == reflect.Float32 ||
		kind == reflect.Float64 ||
		kind == reflect.Complex64 ||
		kind == reflect.Complex128 ||
		kind == reflect.String ||
		kind == reflect.Struct
}

type structFieldInfo struct {
	i int           // index
	n string        // name
	t *reflect.Type // type
	a bool          // anonymous

	isPtr bool
}

func getStructFieldMap(t reflect.Type, tagMap map[string]*structFieldInfo, nameMap map[string]*structFieldInfo, isAnonymous bool) {
	for i := 0; i < t.NumField(); i++ {
		t := t.Field(i)
		if t.Anonymous && t.Type.Kind() == reflect.Struct {
			getStructFieldMap(t.Type, tagMap, nameMap, true)
			continue
		}

		info := &structFieldInfo{i: i, n: t.Name, t: &(t.Type), a: isAnonymous, isPtr: t.Type.Kind() == reflect.Ptr}
		tag := t.Tag.Get("db")
		if tag != "" {
			tagMap[tag] = info
		}
		nameMap[t.Name] = info
	}
}
