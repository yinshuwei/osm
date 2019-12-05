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

* 直接执行SQL(不支持go template解析)示例

example_sql.go
	
	package main

	import (
		"fmt"
		"time"

		_ "github.com/go-sql-driver/mysql"
		"github.com/yinshuwei/osm"
	)

	// User 用户Model
	type User struct {
		ID         int64
		Email      string
		Nickname   string
		CreateTime time.Time
	}

	func main() {
		o, err := osm.New("mysql", "root:123456@/test?charset=utf8", []string{})
		if err != nil {
			fmt.Println(err.Error())
		}

		//添加
		user := User{
			Email:      "test@foxmail.com",
			Nickname:   "haha",
			CreateTime: time.Now(),
		}
		sql := "INSERT INTO user (email,nickname,create_time) VALUES (#{Email},#{Nickname},#{CreateTime});"
		fmt.Println(o.InsertBySQL(sql, user))

		//查询
		user = User{
			Email: "test@foxmail.com",
		}
		var results []User
		sql = "SELECT id,email,nickname,create_time FROM user WHERE email=#{Email};"
		_, err = o.SelectStructs(sql, user)(&results)
		if err != nil {
			fmt.Println(err.Error())
		}
		for _, u := range results {
			fmt.Println(u)
		}

		//删除
		fmt.Println(o.DeleteBySQL("DELETE FROM user WHERE email=#{Email}", user))

		err = o.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}



* 执行template中的SQL(支持go template解析)示例

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
		"time"

		_ "github.com/go-sql-driver/mysql"
		"github.com/yinshuwei/osm"
	)

	// User 用户model
	type User struct {
		ID         int64
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



## 查询结果类型


* value 查出的结果为单行,并存入不定长的变量上(...)
	
	xml

		<select id="selectResUserValue" result="value">
			SELECT id, email, head_image_url FROM res_user WHERE email=#{Email};
		</select>

	go

		user := ResUser{Email: "test@foxmail.com"}
		var id int64
		var email, headImageURL string
		o.Select("selectResUserValue", user)(&id, &email, &headImageURL)

		log.Println(id, email, headImageURL)

* values 查出的结果为多行,并存入不定长的变量上(...，每个都为array，每个array长度都与结果集行数相同)
	
	xml

		<select id="selectResUserValues" result="values">
			SELECT id,email,head_image_url FROM res_user WHERE city=#{City} order by id;
		</select>

	go

		user := ResUser{City: "上海"}
		var ids []int64
		var emails, headImageUrls []string
		o.Select("selectResUserValues", user)(&ids, &emails, &headImageUrls)

		log.Println(ids, emails, headImageUrls)

* struct  查出的结果为单行,并存入struct
	
	xml

		<select id="selectResUser" result="struct">
			SELECT id, email, head_image_url FROM res_user WHERE email=#{Email};
		</select>

	go

		user := ResUser{Email: "test@foxmail.com"}
		var result ResUser
		o.Select("selectResUser", user)(&result)

		log.Printf("%#v", result)

* structs 查出的结果为多行,并存入struct array
	
	xml

		<select id="selectResUsers" result="structs">
			SELECT id,email,head_image_url FROM res_user WHERE city=#{City} order by id;
		</select>

	go

		user := ResUser{City: "上海"}
		var results []*ResUser // 或var results []ResUser
		o.Select("selectResUsers", user)(&results)
		log.Printf("%#v", results)

* kvs 查出的结果为多行,每行有两个字段,前者为key,后者为value,存入map (双列)
	
	xml

		<select id="selectResUserKvs" result="kvs">
			SELECT id,email FROM res_user WHERE city=#{City} order by id;
		</select>

	go

		user := ResUser{City: "上海"}
	    var idEmailMap map[int64]string
		o.Select("selectResUserKvs", user)(&idEmailMap)
		log.Println(idEmailMap)


## struct与SQL列对应关系

* 正常的转换过程

    用"_"分隔 （例：XXX_YYY -> XXX,YYY）

	每个部分全部转为首字大写其余字符小写 （例：XXX,YYY -> Xxx,Yyy）
	
	拼接（例：Xxx,Yyy -> XxxYyy）

* 常见缩写单词，下面这些单词两种形式都可以，struct上可以任选其一。
	
	比如"UserId"和"UserID"可以正常对应到"user_id"列上。但是同一个struct中不可以既有"UserId"成员又有"UserID"成员，如果同时存在只会有一个成员会被赋值。
    
		Acl  或   ACL 
		Api  或   API 
		Ascii  或 ASCII 
		Cpu  或   CPU 
		Css  或   CSS 
		Dns  或   DNS 
		Eof  或   EOF 
		Guid  或  GUID 
		Html  或  HTML 
		Http  或  HTTP 
		Https  或 HTTPS 
		Id  或    ID 
		Ip  或    IP 
		Json  或  JSON 
		Lhs  或   LHS 
		Qps  或   QPS 
		Ram  或   RAM 
		Rhs  或   RHS 
		Rpc  或   RPC 
		Sla  或   SLA 
		Smtp  或  SMTP 
		Sql  或   SQL 
		Ssh  或   SSH 
		Tcp  或   TCP 
		Tls  或   TLS 
		Ttl  或   TTL 
		Udp  或   UDP 
		Ui  或    UI 
		Uid  或   UID 
		Uuid  或  UUID 
		Uri  或   URI 
		Url  或   URL 
		Utf8  或  UTF8 
		Vm  或    VM 
		Xml  或   XML 
		Xmpp  或  XMPP 
		Xsrf  或  XSRF 
		Xss  或   XSS 
