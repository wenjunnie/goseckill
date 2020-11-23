package repositories

import (
	"database/sql"
	"goseckill/common"
	"goseckill/datamodels"
)

//第一步，先开发对应的接口
//第二步，实现定义的接口
type IProduct interface {
	//连接数据
	Conn() error
	Insert(*datamodels.Product)(int64, error)
	Delete(int64) bool
	Update(*datamodels.Product) error
	SelectByKey(int64)(*datamodels.Product, error)
	SelectAll()([]*datamodels.Product, error)
}

type ProductManager struct {
	table string
	mysqlConn *sql.DB
}

//构造函数，实现接口自检
func NewProductManager(table string, db *sql.DB) IProduct {
	return &ProductManager{table:table,mysqlConn:db}
}

func (p ProductManager) Conn() error {
	if p.mysqlConn == nil {
		mysql,err := common.NewMysqlConn()
		if err != nil {
			return err
		}
		p.mysqlConn = mysql
	}
	if p.table == "" {
		p.table = "product"
	}
	return nil
}

func (p ProductManager) Insert(product *datamodels.Product) (int64, error) {
	panic("implement me")
}

func (p ProductManager) Delete(i int64) bool {
	panic("implement me")
}

func (p ProductManager) Update(product *datamodels.Product) error {
	panic("implement me")
}

func (p ProductManager) SelectByKey(i int64) (*datamodels.Product, error) {
	panic("implement me")
}

func (p ProductManager) SelectAll() ([]*datamodels.Product, error) {
	panic("implement me")
}
