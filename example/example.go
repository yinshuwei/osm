package main

import (
	_ "github.com/lib/pq"
	//_ "github.com/go-sql-driver/mysql"
	"github.com/yinshuwei/osm"
	"github.com/yinshuwei/utils"
	"log"
	"time"
)

type ResUser struct {
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
	osm.ShowSql = true
	utils.SetLogFlags(log.Ldate | log.Lshortfile)

	o, err := osm.New("postgres", "host=db01 user=golang password=123456 dbname=golang sslmode=disable", []string{"test.xml"})
	//o, err := osm.New("mysql", "root:root@/test?charset=utf8", []string{"test.xml"})
	if err != nil {
		log.Println(err.Error())
		return
	}

	start := time.Now().Nanosecond() / 1000000

	user := ResUser{Email: "test@foxmail.com", Id: 17}

	/*************/
	log.Println("structs")
	var users []ResUser
	_, err = o.Select("selectResUsers", user)(&users)
	if err != nil {
		log.Println(err)
		return
	}
	for _, u := range users {
		log.Println(u.Id, u.Email)
	}

	/*************/
	log.Println("struct")
	u := ResUser{}
	o.Select("selectResUser", user)(&u)

	log.Println(u.Id, u.Email)

	/***************/
	log.Println("maps")
	var userMaps []map[string]osm.Data
	o.Select("selectResUserMaps", user)(&userMaps)

	for _, uMap := range userMaps {
		log.Println(uMap["Id"].Int64(), uMap["Email"].String())
	}

	/***************/
	log.Println("map")
	var userMap map[string]osm.Data
	o.Select("selectResUserMap", user)(&userMap)

	log.Println(userMap["Id"].Int64(), userMap["Email"].String())

	/***************/
	log.Println("arrays")
	var userArrays [][]osm.Data
	o.Select("selectResUserArrays", "test@foxmail.com")(&userArrays)

	for _, uArray := range userArrays {
		if uArray != nil && len(uArray) >= 2 {
			log.Println(uArray[0].Int64(), uArray[1].String())
		}
	}

	/***************/
	log.Println("array")
	var userArray []osm.Data
	o.Select("selectResUserArray", user)(&userArray)

	if userArray != nil && len(userArray) >= 2 {
		log.Println(userArray[0].Int64(), userArray[1].String())
	}

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
		Id:         4,
		Email:      "test@foxmail.com",
		Birth:      time.Now(),
		CreateTime: time.Now(),
	}
	log.Println(o.Update("updateResUser", updateResUser))

	/*****************/
	log.Println("delete")
	deleteResUser := ResUser{Id: 3}
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
