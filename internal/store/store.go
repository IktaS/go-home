package store

//Repo is an interface that defines what a repository should have
type Repo interface {
	Init(interface{}) error
	Save(interface{}) error
	Get(interface{}) error
	GetAll() error
	Delete(interface{}) error
}
