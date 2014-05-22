package main

import (
	osmPack "cms/osm"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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

	osm, err := osmPack.NewOsm("mysql", "root:root@/51jczj?charset=utf8", []string{"test.xml"})
	if err != nil {
		fmt.Println(err.Error())
	}

	start := time.Now().Nanosecond() / 1000000

	user := User{Email: "yinshuwei@foxmail.com", Id: 17}

	/*************/
	fmt.Println("structs")
	var users []User
	osm.Query("selectUsers", user)(&users)

	for _, u := range users {
		fmt.Println(u.Id, u.Email)
	}

	/*************/
	fmt.Println("\nstruct")
	u := User{}
	osm.Query("selectUser", user)(&u)

	fmt.Println(u.Id, u.Email)

	/***************/
	fmt.Println("\nmaps")
	var userMaps []map[string]osmPack.Data
	osm.Query("selectUserMaps", user)(&userMaps)

	for _, uMap := range userMaps {
		fmt.Println(uMap["Id"].Int64(), uMap["Email"].String())
	}

	/***************/
	fmt.Println("\nmap")
	var userMap map[string]osmPack.Data
	osm.Query("selectUserMap", user)(&userMap)

	fmt.Println(userMap["Id"].Int64(), userMap["Email"].String())

	/***************/
	fmt.Println("\narrays")
	var userArrays [][]osmPack.Data
	osm.Query("selectUserArrays", "yinshuwei@foxmail.com")(&userArrays)

	for _, uArray := range userArrays {
		if uArray != nil && len(uArray) >= 2 {
			fmt.Println(uArray[0].Int64(), uArray[1].String())
		}
	}

	/***************/
	fmt.Println("\narray")
	var userArray []osmPack.Data
	osm.Query("selectUserArray", user)(&userArray)

	if userArray != nil && len(userArray) >= 2 {
		fmt.Println(userArray[0].Int64(), userArray[1].String())
	}

	/***************/
	fmt.Println("\nvalue")
	var id int64
	var email string
	osm.Query("selectUserValue", user)(&id, &email)

	fmt.Println(id, email)

	/***************/
	fmt.Println("\nkvs")
	var idEmailMap map[int64]string
	osm.Query("selectUserKvs", user)(&idEmailMap)

	for k, v := range idEmailMap {
		fmt.Println(k, v)
	}

	/*****************/
	fmt.Println("\ninsert")
	insertUser := User{
		// Id:         2,
		Email:      "yinshuwei@foxmail.com",
		Mobile:     "13113113113",
		Nickname:   "haha",
		Birth:      time.Now(),
		CreateTime: time.Now(),
	}
	fmt.Println(osm.Insert("insertUser", insertUser))

	/*****************/
	fmt.Println("\nupdate")
	updateUser := User{
		Id:         4,
		Email:      "yinshuwei@foxmail.com",
		Birth:      time.Now(),
		CreateTime: time.Now(),
	}
	fmt.Println(osm.Update("updateUser", updateUser))

	/*****************/
	fmt.Println("\ndelete")
	deleteUser := User{Id: 3}
	fmt.Println(osm.Delete("deleteUser", deleteUser))

	// tx, err := osm.Begin()

	// /*****************/
	// fmt.Println("\ninsert")
	// txInsertUser := User{
	// 	// Id:         2,
	// 	Email:      "yinshuwei@foxmail.com",
	// 	Mobile:     "13113113113",
	// 	Nickname:   "haha",
	// 	Birth:      time.Now(),
	// 	CreateTime: time.Now(),
	// }
	// fmt.Println(tx.Insert("insertUser", txInsertUser))

	// tx.Commit()

	fmt.Println(time.Now().Nanosecond()/1000000-start, "ms")
}
