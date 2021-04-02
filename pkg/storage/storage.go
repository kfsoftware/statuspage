package storage

type Storage interface {
	StoreFailure()
	UpdateCheck()
}
