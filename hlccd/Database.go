package Database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"time"
)
//状态描述
type Status struct {
	TimeStamp time.Time
	StatusCode int
	StatusDesc string
	StatusErr error
}

//数据库
type DB struct {
	DriverName string
	User string
	Password string
	Tcp string
	Name string
}
//数据表
type Table struct {
	Name string
	Value []string
	Annotation string
}

//打开数据库并返回
func OpenDatabase(D DB)(*sql.DB,bool,Status){
	db, err := sql.Open(D.DriverName,D.User+":"+D.Password+"@tcp("+D.Tcp+")/"+D.Name+"?charset=utf8")
	if db!=nil {
		db.SetConnMaxLifetime(1000)
		_=db.Ping()
		status:=Status{
			TimeStamp:  time.Now(),
			StatusCode: 10101,
			StatusDesc: D.Name+"数据库正常打开并返回",
			StatusErr:  err,
		}
		j,_:=json.Marshal(status)
		fmt.Println(string(j))
		return db,true,status
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10102,
		StatusDesc: D.Name+"数据库打开失败",
		StatusErr:  err,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return nil,false,status
}
//添加数据表
func CreateTable(db *sql.DB,tab Table) (bool,Status) {
	query:="create table "+tab.Name+"("
	for x:=range tab.Value{
		if x > 0 {
			query+=","
		}
		query+=tab.Value[x]
	}
	query=query+")"+tab.Annotation+";"
	_,err:=db.Exec(query)
	if err == nil {
		status:=Status{
			TimeStamp:  time.Now(),
			StatusCode: 10201,
			StatusDesc: tab.Name+"数据表创建成功",
			StatusErr:  err,
		}
		j,_:=json.Marshal(status)
		fmt.Println(string(j))
		return true,status
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10202,
		StatusDesc: tab.Name+"数据表创建失败",
		StatusErr:  err,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return false,status
}
//删除数据表
func DeleteTable(db *sql.DB,name string) (bool,Status) {
	query:="drop table "+name+";"
	_,err:=db.Exec(query)
	if err == nil {
		status:=Status{
			TimeStamp:  time.Now(),
			StatusCode: 10203,
			StatusDesc: name+"数据表删除成功",
			StatusErr:  err,
		}
		j,_:=json.Marshal(status)
		fmt.Println(string(j))
		return true,status
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10204,
		StatusDesc: name+"数据表删除失败",
		StatusErr:  err,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return false,status
}
//插入数据
func InsertData(db *sql.DB,name string,values string) (bool,Status) {
	query:="insert into "+name+" values("+values+");"
	_,err:=db.Exec(query)
	if err == nil {
		status:=Status{
			TimeStamp:  time.Now(),
			StatusCode: 10301,
			StatusDesc: "数据插入已成功插入数据表："+name,
			StatusErr:  err,
		}
		j,_:=json.Marshal(status)
		fmt.Println(string(j))
		return true,status
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10302,
		StatusDesc: "数据插入未成功插入数据表："+name,
		StatusErr:  err,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return false,status
}
//删除数据
func DeleteData(db *sql.DB,name string,where string) (bool,Status) {
	query:="delete from "+name+" where "+where+";"
	_,err:=db.Exec(query)
	if err == nil {
		status:=Status{
			TimeStamp:  time.Now(),
			StatusCode: 10303,
			StatusDesc: "数据已从数据表： "+name+" 中删除",
			StatusErr:  err,
		}
		j,_:=json.Marshal(status)
		fmt.Println(string(j))
		return true,status
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10304,
		StatusDesc: "数据未从数据表： "+name+" 中删除",
		StatusErr:  err,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return false,status
}
//更新数据
func UpdateData(db *sql.DB,name string,values string,where string) (bool,Status) {
	query:="update "+name+" set "+values+" where "+where+";"
	_,err:=db.Exec(query)
	if err == nil {
		status:=Status{
			TimeStamp:  time.Now(),
			StatusCode: 10305,
			StatusDesc: "从数据表: "+name+" 中更新数据成功",
			StatusErr:  err,
		}
		j,_:=json.Marshal(status)
		fmt.Println(string(j))
		return true,status
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10306,
		StatusDesc: "从数据表: "+name+" 中更新数据失败",
		StatusErr:  err,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return false,status
}
//查询数据
func SelectAllData(db *sql.DB,name string,field string,where string) ( *sql.Rows,bool,Status) {
	query:="select "+field+" from "+name
	if where!="" {
		query+=" where "+where+";"
	}else {
		query+=";"
	}
	rows,err:=db.Query(query)
	if err == nil {
		status:=Status{
			TimeStamp:  time.Now(),
			StatusCode: 10401,
			StatusDesc: "从数据表: "+name+" 中查询数据成功",
			StatusErr:  err,
		}
		j,_:=json.Marshal(status)
		fmt.Println(string(j))
		return rows,true,status
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10402,
		StatusDesc: "从数据表: "+name+" 中查询数据失败",
		StatusErr:  err,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return rows,false,status
}
//返回ID主键的最后一条记录
func SelectLastData(db *sql.DB,field string,name string) ( *sql.Rows,bool,Status) {
	rows,err:=db.Query("select "+field+" from "+name+" order by id desc limit 1")
	if err == nil {
		status:=Status{
			TimeStamp:  time.Now(),
			StatusCode: 10403,
			StatusDesc: "从数据表: "+name+" 中最后一条数据成功",
			StatusErr:  err,
		}
		j,_:=json.Marshal(status)
		fmt.Println(string(j))
		return rows,true,status
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10404,
		StatusDesc: "从数据表: "+name+" 中最后一条数据失败",
		StatusErr:  err,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return rows,false,status
}
//在表内查找多组key是否存在
func SelectKeysIsExist(db *sql.DB,name string,KeyType string,key []string) bool {
	query:="select "+KeyType+" from "+name+" where "+KeyType+"='"
	for x:=range key{
		tmp:=query+key[x]+"';"
		rows,err:=db.Query(tmp)
		defer rows.Close()
		if err != nil {
			status:=Status{
				TimeStamp:  time.Now(),
				StatusCode: 10405,
				StatusDesc: "第"+strconv.Itoa(x)+"个在数据表： "+name+" 中不存在",
				StatusErr:  err,
			}
			j,_:=json.Marshal(status)
			fmt.Println(string(j))
			return false
		}
		if rows.Next(){
			var s string
			err:=rows.Scan(&s)
			if err != nil{
				status:=Status{
					TimeStamp:  time.Now(),
					StatusCode: 10406,
					StatusDesc: "第"+strconv.Itoa(x)+"个读入失败",
					StatusErr:  err,
				}
				j,_:=json.Marshal(status)
				fmt.Println(string(j))
				return false
			}
			if s != key[x] {
				status:=Status{
					TimeStamp:  time.Now(),
					StatusCode: 10407,
					StatusDesc: "第"+strconv.Itoa(x)+"个读取结果与输入结果不一致",
					StatusErr:  err,
				}
				j,_:=json.Marshal(status)
				fmt.Println(string(j))
				return false
			}
		}else {
			status:=Status{
				TimeStamp:  time.Now(),
				StatusCode: 10408,
				StatusDesc: "第"+strconv.Itoa(x)+"个在读取时为空",
				StatusErr:  err,
			}
			j,_:=json.Marshal(status)
			fmt.Println(string(j))
			return false
		}
	}
	status:=Status{
		TimeStamp:  time.Now(),
		StatusCode: 10409,
		StatusDesc: "该组数据皆存在于数据表: "+name+" 中",
		StatusErr:  nil,
	}
	j,_:=json.Marshal(status)
	fmt.Println(string(j))
	return true
}
//按条件key查找并返回单个string类型的field
func SelectKeyGetFieldS(db *sql.DB,name string,field string,key string) string {
	query:="select "+field+" from "+name+" where "+key+";"
	rows,err:=db.Query(query)
	defer rows.Close()
	if err != nil {
		return ""
	}
	if rows.Next(){
		var s string
		err:=rows.Scan(&s)
		if err != nil{
			return ""
		}else {
			return s
		}
	}
	return ""
}
//按条件key查找并返回单个int类型的field
func SelectKeyGetFieldI(db *sql.DB,name string,field string,key string) int64 {
	query:="select "+field+" from "+name+" where "+key+";"
	fmt.Println(query)
	rows,err:=db.Query(query)
	defer rows.Close()
	if err != nil {
		return -1
	}
	if rows.Next(){
		var i int64
		err:=rows.Scan(&i)
		if err != nil{
			return -1
		}else {
			return i
		}
	}
	return -1
}
//按条件key查找并返回多个int64类型的field
func SelectKeyGetFieldsI(db *sql.DB,name string,field string,key string) []int64 {
	query:="select "+field+" from "+name+" where "+key+";"
	fmt.Println(query)
	rows,err:=db.Query(query)
	arr :=make([]int64,0)
	defer rows.Close()
	if err != nil {
		return arr
	}
	for rows.Next(){
		var i int64
		err:=rows.Scan(&i)
		if err != nil{
			return arr
		}else {
			arr=append(arr,i)
		}
	}
	return arr
}