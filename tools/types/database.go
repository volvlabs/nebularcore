package types

type DatabaseType int

const (
	DatabasePostgres DatabaseType = iota
	GoogleCloudsqlPostgres
	DatabaseSqlite
)
