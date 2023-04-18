package dbconn

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type Table struct {
	TableName string
}

type Entity struct {
	ID   uint   `json:"id" gorm:"primary_key"`
	Name string `json:"name"`
}

type Column struct {
	ColumnName string `gorm:"column:column_name"`
}

var dbConn *gorm.DB

func InitDB(dialect, hostname, name string, dbPort int, user, password string) (*gorm.DB, error) {
	dataSourceName := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s port=%d", hostname, user, name, password, dbPort)
	if dialect == "postgres" {
		_dbConn, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{})
		dbConn = _dbConn
		return dbConn, err
	}
	return nil, errors.New("dialect must be 'postgres'")
}

func GetTables() ([]Table, error) {
	var tables []Table
	if dbConn == nil {
		return nil, errors.New("dbConn is nil")
	}
	err := dbConn.Raw("SELECT table_name  FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables).Error
	return tables, err
}

func GetColumns(tableName string) ([]Column, error) {
	var columns []Column
	err := dbConn.Raw("SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = ?", tableName).Scan(&columns).Error
	return columns, err
}
