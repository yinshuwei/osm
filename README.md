osm
===

osm(Object Sql Mapping)是用go编写的ORM工具，目前很简单，只能算是半成品，只支持mysql(因为我目前的项目是mysql,所以其他数据库没有测试过)。

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


##osm获取
	go get github.com/yinshuwei/osm

##实例
创建数据库
	
	create database test;
	use test;
创建user表
	
	CREATE TABLE `user` (
	  `id` int(11) NOT NULL AUTO_INCREMENT,
	  `email` varchar(255) DEFAULT NULL,
	  `mobile` varchar(45) DEFAULT NULL,
	  `nickname` varchar(45) DEFAULT NULL,
	  `password` varchar(255) DEFAULT NULL,
	  `description` varchar(255) DEFAULT NULL COMMENT '自我描述',
	  `name` varchar(45) DEFAULT NULL,
	  `birth` date DEFAULT NULL,
	  `province` varchar(45) DEFAULT NULL COMMENT '省',
	  `city` varchar(45) DEFAULT NULL COMMENT '市',
	  `company` varchar(45) DEFAULT NULL COMMENT '公司',
	  `address` varchar(45) DEFAULT NULL COMMENT '地址',
	  `sex` varchar(45) DEFAULT NULL COMMENT '性别',
	  `contact_info` varchar(1000) DEFAULT NULL COMMENT '联系方式',
	  `create_time` varchar(255) DEFAULT NULL,
	  PRIMARY KEY (`id`)
	) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 COMMENT='用户表';

sql xml文件text.xml

	<?xml version="1.0" encoding="utf-8"?>
	<osm>
		<insert id="insertUser">
	INSERT INTO user
	(email,mobile,nickname,password,description,name,birth,province,city,company,address,sex,contact_info,create_time)
	VALUES
	(#{Email},#{Mobile},#{Nickname},#{Password},#{Description},#{Name},#{Birth},#{Province},#{City},#{Company},#{Address},#{Sex},#{ContactInfo},#{CreateTime});
		</insert>

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
		Id          int64
		Email       string
		Mobile      string
		Nickname    string
		Password    string
		Description string
		Name        string
		Birth       time.Time
		Province    string
		City        string
		Company     string
		Address     string
		Sex         string
		ContactInfo string
		CreateTime  time.Time
	}

	func main() {

		o, err := osm.New("mysql", "root:root@/test?charset=utf8", []string{"test.xml"})
		if err != nil {
			fmt.Println(err.Error())
		}

		user := User{
			// Id:         2,
			Email:      "test@foxmail.com",
			Mobile:     "13113113113",
			Nickname:   "haha",
			Birth:      time.Now(),
			CreateTime: time.Now(),
		}
		fmt.Println(o.Insert("insertUser", user))
		
		err = o.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
