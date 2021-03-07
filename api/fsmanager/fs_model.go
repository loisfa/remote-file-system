package fsmanager

type File struct {
	Id       int
	Name     string
	Path     string
	ParentId *int
}

type Folder struct {
	Id       int
	Name     string
	ParentId *int
}
