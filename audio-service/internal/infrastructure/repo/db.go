package repo

type DB interface {
	ReadQuery(query string, dest interface{}, args ...interface{}) error
	WriteQuery(query string, source interface{}) error
}
