package main

import (
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
	"github.com/yinshuwei/osm"
)

// ResUser 测试用实体
type ResUser struct {
	ID           int64
	Email        string
	Mobile       string
	Nickname     string
	Password     string
	HeadImageURL string
	Description  string
	Name         string
	Birth        time.Time
	Province     string
	City         string
	Company      string
	Address      string
	Sex          string
	ContactInfo  string
	CreateTime   time.Time
}

func main() {
	osm.ShowSQL = true
	log.SetFlags(log.Ldate | log.Lshortfile)

	// o, err := osm.New("postgres", "host=db01 user=golang password=123456 dbname=golang sslmode=disable", []string{"test.xml"})
	o, err := osm.New("mysql", "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8", []string{"test.xml"})
	if err != nil {
		log.Println(err.Error())
		return
	}

	start := time.Now()

	user := ResUser{Email: "test@foxmail.com", ID: 1}

	/*************/
	log.Println("structs")
	var users []ResUser
	_, err = o.Select("selectResUsers", user)(&users)
	if err != nil {
		log.Println(err)
		return
	}
	for _, u := range users {
		log.Println(u.ID, u.Email, u.HeadImageURL)
	}

	/*************/
	log.Println("\n\nstruct")
	u := ResUser{}
	o.Select("selectResUser", user)(&u)

	log.Println(u.ID, u.Email, u.HeadImageURL)

	/***************/
	log.Println("\n\nvalue")
	var id int64
	var email string
	var headImageURL string
	o.Select("selectResUserValue", user)(&id, &email, &headImageURL)

	log.Println(id, email, headImageURL)

	/***************/
	log.Println("\n\nvalues")
	var ids []int64
	var emails []string
	var headImageUrls []string
	o.Select("selectResUserValues", user)(&ids, &emails, &headImageUrls)

	log.Println(ids, emails, headImageUrls)

	/***************/
	log.Println("\n\nkvs")
	var idEmailMap map[int64]string
	o.Select("selectResUserKvs", user)(&idEmailMap)

	for k, v := range idEmailMap {
		log.Println(k, v)
	}

	/*****************/
	log.Println("\n\ninsert")
	insertResUser := ResUser{
		Email:        "test@foxmail.com",
		Mobile:       "13113113113",
		Nickname:     "haha",
		HeadImageURL: "www.baidu.com",
		Password:     "password",
		Description:  "地球人",
		Name:         "张三",
		Province:     "上海",
		City:         "上海",
		Company:      "电信",
		Address:      "沪南路1155号",
		Sex:          "女",
		ContactInfo:  `{"QQ":"8888888"}`,
		Birth:        time.Now(),
		CreateTime:   time.Now(),
	}
	log.Println(o.Insert("insertResUser", insertResUser))

	/*****************/
	log.Println("\n\nupdate")
	updateResUser := ResUser{
		ID:           5,
		Email:        "test@foxmail.com",
		Birth:        time.Now(),
		HeadImageURL: "www.qq.com",
		CreateTime:   time.Now(),
	}
	log.Println(o.Update("updateResUser", updateResUser))

	/*****************/
	log.Println("\n\ndelete")
	deleteResUser := ResUser{ID: 6}
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

	log.Println(time.Now().Sub(start))

	err = o.Close()
	if err != nil {
		log.Println(err.Error())
	}
}
