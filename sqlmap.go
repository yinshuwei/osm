package osm

import (
	"encoding/xml"
	"os"
	"strings"
	"text/template"
)

const (
	typeSelect = 1
	typeUpdate = iota
	typeInsert = iota
	typeDelete = iota

	resultTypeValue   = "value"   //查出的结果为单行,并存入不定长的变量上(...)
	resultTypeValues  = "values"  //查出的结果为多行,并存入不定长的变量上(...，每个都为array)
	resultTypeStruct  = "struct"  //查出的结果为单行,并存入struct
	resultTypeStructs = "structs" //查出的结果为多行,并存入struct array
	resultTypeKvs     = "kvs"     //查出的结果为多行,每行有两个字段,前者为key,后者为value,存入map (双列)
)

type sqlMapper struct {
	id          string
	sqlTemplate *template.Template
	sql         string
	sqlType     int
	result      string
}

type stmtXML struct {
	ID     string `xml:"id,attr"`
	Result string `xml:"result,attr"`
	SQL    string `xml:",chardata"`
}

type osmXML struct {
	Selects []stmtXML `xml:"select"`
	Deletes []stmtXML `xml:"delete"`
	Updates []stmtXML `xml:"update"`
	Inserts []stmtXML `xml:"insert"`
}

func readMappers(path string) (sqlMappers []*sqlMapper, err error) {
	sqlMappers = make([]*sqlMapper, 0)
	err = nil

	xmlFile, err := os.Open(path)
	if err != nil {
		logger.Println("Error opening file: ", err)
		return
	}
	defer xmlFile.Close()

	osmXMLObj := osmXML{}

	decoder := xml.NewDecoder(xmlFile)

	if err = decoder.Decode(&osmXMLObj); err != nil {
		logger.Println("Error decode file: ", path, err)
		return
	}

	for _, deleteStmt := range osmXMLObj.Deletes {
		sqlMappers = append(sqlMappers, newMapper(deleteStmt, typeDelete))
	}
	for _, insertStmt := range osmXMLObj.Inserts {
		sqlMappers = append(sqlMappers, newMapper(insertStmt, typeInsert))
	}
	for _, selectStmt := range osmXMLObj.Selects {
		sqlMappers = append(sqlMappers, newMapper(selectStmt, typeSelect))
	}
	for _, updateStmt := range osmXMLObj.Updates {
		sqlMappers = append(sqlMappers, newMapper(updateStmt, typeUpdate))
	}
	return
}

func newMapper(stmt stmtXML, sqlType int) (sqlMapperObj *sqlMapper) {
	sqlMapperObj = new(sqlMapper)
	sqlMapperObj.id = stmt.ID
	sqlMapperObj.sqlType = sqlType
	sqlMapperObj.result = stmt.Result

	sqlTemp := strings.Replace(stmt.SQL, "\n", " ", -1)
	sqlTemp = strings.Replace(sqlTemp, "\t", " ", -1)
	for strings.Contains(sqlTemp, "  ") {
		sqlTemp = strings.Replace(sqlTemp, "  ", " ", -1)
	}
	sqlTemp = strings.Trim(sqlTemp, "\t\n ")

	sqlMapperObj.sql = sqlTemp

	var err error
	sqlMapperObj.sqlTemplate, err = template.New(stmt.ID).Parse(sqlTemp)

	if err != nil {
		logger.Println("sql template create error", err.Error())
	}

	return
}

func markSQLError(sql string, index int) string {
	// result := fmt.Sprintf("%s[****ERROR****]->%s", sql[0:index], sql[index:])
	result := strings.Join([]string{sql[0:index], "[****ERROR****]->", sql[index:]}, "")
	return result
}
