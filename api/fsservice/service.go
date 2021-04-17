package fsservice

import (
	"errors"

	"github.com/loisfa/remote-file-system/api/fsmodel"
	"github.com/loisfa/remote-file-system/api/fsrepository"
)

type IFileSystemService interface {
	GetRootFolderID() (*int, error)
	GetFolder(folderID int) (*fsmodel.Folder, error)
	ExistsFolder(folderID int) (*bool, error)
	GetFile(fileID int) (*fsmodel.File, error)
	GetFoldersIn(folderID int) (*[]fsmodel.Folder, error)
	GetFilesIn(folderID int) (*[]fsmodel.File, error)
	CreateFolder(name string, parentID int) (*int, error)
	CreateFile(name string, path string, parentID int) (*int, error)
	UpdateFolder(folderID int, name string) error
	MoveFolder(folderID int, destFolderID int) error
	MoveFile(fileID int, destFolderID int) error
	DeleteFolderAndContent(folderID int) error
	DeleteFile(fileID int) error
}

type FileSystemService struct {
	repo fsrepository.IFileSystemRepository
}

// could use a builder pattern?
func NewFileSystemService() FileSystemService {
	return FileSystemService{
		repo: fsrepository.NewNeo4JFileSystemRepository(),
	}
}

func (svc FileSystemService) GetRootFolderID() (id *int, err error) {
	return svc.repo.GetRootFolderID()
}

func (svc FileSystemService) GetFolder(folderID int) (*fsmodel.Folder, error) {
	exists, err := svc.repo.ExistsFolder(folderID)
	if err != nil {
		return nil, err
	}
	if !*exists {
		return nil, nil
	}

	folder, err := svc.repo.GetFolder(folderID)
	if err != nil {
		return nil, err
	}
	return folder, err
}

func (svc FileSystemService) ExistsFolder(folderID int) (*bool, error) {
	return svc.repo.ExistsFolder(folderID)
}

func (svc FileSystemService) GetFile(fileID int) (*fsmodel.File, error) {
	if exists, err := svc.repo.ExistsFile(fileID); err != nil || exists == nil || *exists == false {
		return nil, errors.New("The file does not exist. Cannot be fetched")
	}

	file, err := svc.repo.GetFile(fileID)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("Could not find file for the specified id")
	}

	return file, err
}

// TODO could have one single database call to return at the same time: currentFolder, folders, files
func (svc FileSystemService) GetFoldersIn(folderID int) (*[]fsmodel.Folder, error) {
	return svc.repo.GetFoldersIn(folderID)
}

func (svc FileSystemService) GetFilesIn(folderID int) (*[]fsmodel.File, error) {
	return svc.repo.GetFilesIn(folderID)
}

func (svc FileSystemService) CreateFolder(name string, parentID int) (*int, error) {
	return svc.repo.CreateFolder(name, parentID)
}

func (svc FileSystemService) CreateFile(name string, path string, parentID int) (*int, error) {
	return svc.repo.CreateFile(name, path, parentID)
}

func (svc FileSystemService) UpdateFolder(folderID int, name string) error {
	return svc.repo.UpdateFolder(folderID, name)
}

func (svc FileSystemService) MoveFolder(folderID int, destFolderID int) error {
	if found, err := svc.repo.ExistsFolder(destFolderID); err != nil || found == nil || *found == false {
		return errors.New("The destination folder does not exist. Folder cannot be moved there.")
	}

	if found, err := svc.repo.ExistsFolder(folderID); err != nil || found == nil || *found == false {
		return errors.New("The folder was not found.")
	}

	if isRoot, err := svc.repo.IsRootFolder(folderID); err != nil || isRoot == nil || *isRoot == true {
		return errors.New("Cannot perform 'Move' operation on the root folder")
	}

	return svc.repo.MoveFolder(folderID, destFolderID)
}

func (svc FileSystemService) MoveFile(fileID int, destFolderID int) error {
	if found, err := svc.repo.ExistsFolder(destFolderID); err != nil || found == nil || *found == false {
		return errors.New("The destination folder does not exist. File cannot be moved there.")
	}
	return svc.repo.MoveFile(fileID, destFolderID)
}

func (svc FileSystemService) DeleteFolderAndContent(folderID int) error {
	if found, err := svc.repo.ExistsFolder(folderID); err != nil || found == nil || *found == false {
		return errors.New("The folder does not exist. It cannot be deleted.")
	}

	if isRoot, err := svc.repo.IsRootFolder(folderID); err != nil || isRoot == nil || *isRoot == true {
		return errors.New("Trying to delete the root folder. Operation not permitted")
	}

	return svc.repo.DeleteFolderAndContent(folderID)
}

func (svc FileSystemService) DeleteFile(fileID int) error {
	if found, err := svc.repo.ExistsFile(fileID); err != nil || found == nil || *found == false {
		return errors.New("The file does not exist. It cannot be deleted.")
	}
	return svc.repo.DeleteFile(fileID)
}
