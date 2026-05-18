package repo

type DB interface {
	Select(dest interface{}, query string, args ...interface{}) error
	NamedExec(source interface{}, query string) error
}
