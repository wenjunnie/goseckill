package common

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

//创建MySQL连接
func NewMysqlConn() (db *sql.DB, err error) {
	db, err = sql.Open("mysql", "root:nie970309@tcp(127.0.0.1:3306)/goseckill?charset=utf8")
	return
}

//获取返回值，获取一条
func GetResultRow(rows *sql.Rows) map[string]string {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([][]byte, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}
	record := make(map[string]string)
	for rows.Next() {
		//将行数据保存到record字典
		rows.Scan(scanArgs...)
		for i, v := range values {
			if v != nil {
				record[columns[i]] = string(v)
			}
		}
	}
	return record
}

//获取所有
func GetResultRows(rows *sql.Rows) map[int]map[string]string {
	//返回所有列
	columns, _ := rows.Columns()
	//这里表示一行填充数据
	scanArgs := make([]interface{}, len(columns))
	//这里表示一行所有列的值，用[]byte表示
	values := make([][]byte, len(columns))
	//这里scanArgs引用values，把数据填充到[]byte里
	for k, _ := range values {
		scanArgs[k] = &values[k]
	}
	i := 0
	result := make(map[int]map[string]string)
	for rows.Next() {
		//填充数据
		rows.Scan(scanArgs...)
		//每行数据
		row := make(map[string]string)
		//把values中的数据复制到row中
		for k, v := range values {
			key := columns[k]
			//这里把[]byte数据转成string
			row[key] = string(v)
		}
		//放入结果集
		result[i] = row
		i++
	}
	return result
}
