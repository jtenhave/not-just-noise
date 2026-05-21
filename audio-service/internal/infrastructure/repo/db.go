package repo

type DB interface {
	ReadQuery(query string, dest interface{}, args ...interface{}) error
	WriteQuery(query string, source interface{}) (int64, error)
}
