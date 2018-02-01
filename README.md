osm
===

osm(Object Sql Mapping And Template)是用go编写的ORM工具，目前已在生产环境中使用，只支持mysql和postgresql(其他数据库没有测试过)。

以前是使用MyBatis开发java服务端，它的sql mapping很灵活，把sql独立出来，程序通过输入与输出来完成所有的数据库操作。

osm就是对MyBatis的简单模仿。当然动态sql的生成是使用go和template包，所以sql mapping的格式与MyBatis的不同。sql xml 格式如下：

	<?xml version="1.0" encoding="utf-8"?>
	<osm>
	 <select id="selectUsers" result="structs">
	   SELECT id,email
	   FROM user
	   {{if ne .Email ""}} where email=#{Email} {{end}}
	   order by id
	 </select>
	</osm>


## osm获取

	go get github.com/yinshuwei/osm

## api doc

http://godoc.org/github.com/yinshuwei/osm

## Quickstart

创建数据库
	
	create database test;
	use test;

创建user表
	
	CREATE TABLE `user` (
	  `id` int(11) NOT NULL AUTO_INCREMENT,
	  `email` varchar(255) DEFAULT NULL,
	  `nickname` varchar(45) DEFAULT NULL,
	  `create_time` varchar(255) DEFAULT NULL,
	  PRIMARY KEY (`id`)
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='user table';

sql template文件test.xml

	<?xml version="1.0" encoding="utf-8"?>
	<osm>
	    <insert id="insertUser">
	    <![CDATA[
	INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});
	    ]]>
	    </insert>
	
	    <select id="selectUser" result="structs">
	    <![CDATA[
	SELECT id,email,nickname,create_time FROM user 
	WHERE 
	{{if ne .Email ""}}email=#{Email} and{{end}}
	{{if ne .Nickname ""}}nickname=#{Nickname} and{{end}}
	1=1;
	    ]]>
	    </select>
	
	    <delete id="deleteUser">
	    <![CDATA[
	DELETE FROM user WHERE email=#{Email}
	    ]]>
	    </delete>
	</osm>

example.go
	
	package main
	
	import (
		"fmt"
		_ "github.com/go-sql-driver/mysql"
		"github.com/yinshuwei/osm"
		"time"
	)
	
	type User struct {
		Id         int64
		Email      string
		Nickname   string
		CreateTime time.Time
	}
	
	func main() {
		o, err := osm.New("mysql", "root:root@/test?charset=utf8", []string{"test.xml"})
		if err != nil {
			fmt.Println(err.Error())
		}
	
		//添加
		user := User{
			Email:      "test@foxmail.com",
			Nickname:   "haha",
			CreateTime: time.Now(),
		}
		fmt.Println(o.Insert("insertUser", user))
	
		//动态查询
		user = User{
			Email: "test@foxmail.com",
		}
		var results []User
		o.Select("selectUser", user)(&results)
		for _, u := range results {
			fmt.Println(u)
		}
	
		//删除
		fmt.Println(o.Delete("deleteUser", user))
	
		err = o.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

