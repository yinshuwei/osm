package main

import (
	_ "github.com/lib/pq"
	//_ "github.com/go-sql-driver/mysql"
	"log"
	"time"

	"github.com/yinshuwei/osm"
)

// ResUser 测试用实体
type ResUser struct {
	ID          int64
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
	osm.ShowSQL = true
	log.SetFlags(log.Ldate | log.Lshortfile)

	o, err := osm.New("postgres", "host=db01 user=golang password=123456 dbname=golang sslmode=disable", []string{"test.xml"})
	//o, err := osm.New("mysql", "root:root@/test?charset=utf8", []string{"test.xml"})
	if err != nil {
		log.Println(err.Error())
		return
	}

	start := time.Now().Nanosecond() / 1000000

	user := ResUser{Email: "test@foxmail.com", ID: 17}

	/*************/
	log.Println("structs")
	var users []ResUser
	_, err = o.Select("selectResUsers", user)(&users)
	if err != nil {
		log.Println(err)
		return
	}
	for _, u := range users {
		log.Println(u.ID, u.Email)
	}

	/*************/
	log.Println("struct")
	u := ResUser{}
	o.Select("selectResUser", user)(&u)

	log.Println(u.ID, u.Email)

	/***************/
	log.Println("value")
	var id int64
	var email string
	o.Select("selectResUserValue", user)(&id, &email)

	log.Println(id, email)

	/***************/
	log.Println("kvs")
	var idEmailMap map[int64]string
	o.Select("selectResUserKvs", user)(&idEmailMap)

	for k, v := range idEmailMap {
		log.Println(k, v)
	}

	/*****************/
	log.Println("insert")
	insertResUser := ResUser{
		Email:      "test@foxmail.com",
		Mobile:     "13113113113",
		Nickname:   "haha",
		Birth:      time.Now(),
		CreateTime: time.Now(),
	}
	log.Println(o.Insert("insertResUser", insertResUser))

	/*****************/
	log.Println("update")
	updateResUser := ResUser{
		ID:         4,
		Email:      "test@foxmail.com",
		Birth:      time.Now(),
		CreateTime: time.Now(),
	}
	log.Println(o.Update("updateResUser", updateResUser))

	/*****************/
	log.Println("delete")
	deleteResUser := ResUser{ID: 3}
	log.Println(o.Delete("deleteResUser", deleteResUser))

	// tx, err := o.Begin()

	// /*****************/
	// log.Println("insert")
	// txInsertResUser := ResUser{
	// 	Email:      "test@foxmail.com",
	// 	Mobile:     "13113113113",
	// 	Nickname:   "haha",
	// 	Birth:      time.Now(),
	// 	CreateTime: time.Now(),
	// }
	// log.Println(tx.Insert("insertResUser", txInsertResUser))

	// tx.Commit()

	log.Println(time.Now().Nanosecond()/1000000-start, "ms")

	err = o.Close()
	if err != nil {
		log.Println(err.Error())
	}
}
