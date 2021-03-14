package fsmodel

type File struct {
	Id       int
	Name     string
	Path     string
	ParentId int // TODO should fill it!
}

type Folder struct {
	Id       int
	Name     string
	ParentId *int // nil in case of root folder
}
