package osm

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const (
	type_select = 1
	type_update = iota
	type_insert = iota
	type_delete = iota

	result_value   = "value"   //查出的结果为单行,并存入不定长的变量上(...)
	result_struct  = "struct"  //查出的结果为单行,并存入struct
	result_structs = "structs" //查出的结果为多行,并存入struct array
	result_map     = "map"     //查出的结果为单行,并存入map
	result_maps    = "maps"    //查出的结果为多行,并存入map array
	result_array   = "array"   //查出的结果为单行,并存入array
	result_arrays  = "arrays"  //查出的结果为多行,并存入array array
	result_kvs     = "kvs"     //查出的结果为多行,每行有两个字段,前者为key,后者为value,存入map
)

type sqlMapper struct {
	id          string
	sqlTemplate *template.Template
	sql         string
	sqlType     int
	result      string
}

type stmtXml struct {
	Id     string `xml:"id,attr"`
	Result string `xml:"result,attr"`
	Sql    string `xml:",chardata"`
}

type osmXml struct {
	Selects []stmtXml `xml:"select"`
	Deletes []stmtXml `xml:"delete"`
	Updates []stmtXml `xml:"update"`
	Inserts []stmtXml `xml:"insert"`
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

	osmXmlObj := osmXml{}

	decoder := xml.NewDecoder(xmlFile)

	if err = decoder.Decode(&osmXmlObj); err != nil {
		logger.Println("Error decode file: ", err)
		return
	}

	for _, deleteStmt := range osmXmlObj.Deletes {
		sqlMappers = append(sqlMappers, newMapper(deleteStmt, type_delete))
	}
	for _, insertStmt := range osmXmlObj.Inserts {
		sqlMappers = append(sqlMappers, newMapper(insertStmt, type_insert))
	}
	for _, selectStmt := range osmXmlObj.Selects {
		sqlMappers = append(sqlMappers, newMapper(selectStmt, type_select))
	}
	for _, updateStmt := range osmXmlObj.Updates {
		sqlMappers = append(sqlMappers, newMapper(updateStmt, type_update))
	}
	return
}

func newMapper(stmt stmtXml, sqlType int) (sqlMapperObj *sqlMapper) {
	sqlMapperObj = new(sqlMapper)
	sqlMapperObj.id = stmt.Id
	sqlMapperObj.sqlType = sqlType
	sqlMapperObj.result = stmt.Result

	sqlTemp := strings.Replace(stmt.Sql, "\n", " ", -1)
	sqlTemp = strings.Replace(sqlTemp, "\t", " ", -1)
	for strings.Contains(sqlTemp, "  ") {
		sqlTemp = strings.Replace(sqlTemp, "  ", " ", -1)
	}
	sqlTemp = strings.Trim(sqlTemp, "\t\n ")

	sqlMapperObj.sql = sqlTemp

	var err error
	sqlMapperObj.sqlTemplate, err = template.New(stmt.Id).Parse(sqlTemp)

	if err != nil {
		logger.Println("sql template create error", err.Error())
	}

	return
}

func markSqlError(sql string, index int) string {
	result := fmt.Sprintf("%s[****ERROR****]->%s", sql[0:index], sql[index:])
	return result
}
