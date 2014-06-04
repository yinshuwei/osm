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

	start := time.Now().Nanosecond() / 1000000

	user := User{Email: "test@foxmail.com", Id: 17}

	/*************/
	fmt.Println("structs")
	var users []User
	o.Query("selectUsers", user)(&users)

	for _, u := range users {
		fmt.Println(u.Id, u.Email)
	}

	/*************/
	fmt.Println("\nstruct")
	u := User{}
	o.Query("selectUser", user)(&u)

	fmt.Println(u.Id, u.Email)

	/***************/
	fmt.Println("\nmaps")
	var userMaps []map[string]osm.Data
	o.Query("selectUserMaps", user)(&userMaps)

	for _, uMap := range userMaps {
		fmt.Println(uMap["Id"].Int64(), uMap["Email"].String())
	}

	/***************/
	fmt.Println("\nmap")
	var userMap map[string]osm.Data
	o.Query("selectUserMap", user)(&userMap)

	fmt.Println(userMap["Id"].Int64(), userMap["Email"].String())

	/***************/
	fmt.Println("\narrays")
	var userArrays [][]osm.Data
	o.Query("selectUserArrays", "test@foxmail.com")(&userArrays)

	for _, uArray := range userArrays {
		if uArray != nil && len(uArray) >= 2 {
			fmt.Println(uArray[0].Int64(), uArray[1].String())
		}
	}

	/***************/
	fmt.Println("\narray")
	var userArray []osm.Data
	o.Query("selectUserArray", user)(&userArray)

	if userArray != nil && len(userArray) >= 2 {
		fmt.Println(userArray[0].Int64(), userArray[1].String())
	}

	/***************/
	fmt.Println("\nvalue")
	var id int64
	var email string
	o.Query("selectUserValue", user)(&id, &email)

	fmt.Println(id, email)

	/***************/
	fmt.Println("\nkvs")
	var idEmailMap map[int64]string
	o.Query("selectUserKvs", user)(&idEmailMap)

	for k, v := range idEmailMap {
		fmt.Println(k, v)
	}

	/*****************/
	fmt.Println("\ninsert")
	insertUser := User{
		// Id:         2,
		Email:      "test@foxmail.com",
		Mobile:     "13113113113",
		Nickname:   "haha",
		Birth:      time.Now(),
		CreateTime: time.Now(),
	}
	fmt.Println(o.Insert("insertUser", insertUser))

	/*****************/
	fmt.Println("\nupdate")
	updateUser := User{
		Id:         4,
		Email:      "test@foxmail.com",
		Birth:      time.Now(),
		CreateTime: time.Now(),
	}
	fmt.Println(o.Update("updateUser", updateUser))

	/*****************/
	fmt.Println("\ndelete")
	deleteUser := User{Id: 3}
	fmt.Println(o.Delete("deleteUser", deleteUser))

	// tx, err := o.Begin()

	// /*****************/
	// fmt.Println("\ninsert")
	// txInsertUser := User{
	// 	// Id:         2,
	// 	Email:      "test@foxmail.com",
	// 	Mobile:     "13113113113",
	// 	Nickname:   "haha",
	// 	Birth:      time.Now(),
	// 	CreateTime: time.Now(),
	// }
	// fmt.Println(tx.Insert("insertUser", txInsertUser))

	// tx.Commit()

	fmt.Println(time.Now().Nanosecond()/1000000-start, "ms")

	err = o.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
}
