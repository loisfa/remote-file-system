package fsservice

import (
	"fmt"

	"github.com/loisfa/remote-file-system/api/fsmodel"
	"github.com/loisfa/remote-file-system/api/fsrepository"
	"github.com/pkg/errors"
)

const (
	NotFound         = "Not found"
	BadRequest       = "Bad Request" // go for bad request when not obvious which resource is not found (ex: params={folderID+destFolderID} => not obvious)
	IllegalOperation = "Illegal operation"
)

type CustomError struct {
	cause string
}

func (error CustomError) Error() string {
	return error.cause
}

type IFileSystemService interface {
	GetRootFolderID() (*int, error)                                  // the function ensures it exists
	GetFolder(folderID int) (*fsmodel.Folder, error)                 // the function ensures it exists
	GetFile(fileID int) (*fsmodel.File, error)                       // the function ensures it exists
	GetFoldersIn(folderID int) (*[]fsmodel.Folder, error)            // the function ensures it exists
	GetFilesIn(folderID int) (*[]fsmodel.File, error)                // the function ensures it exists
	CreateFolder(name string, parentID int) (*int, error)            // the function ensures the parent exists
	CreateFile(name string, path string, parentID int) (*int, error) // the function ensures the parent exists
	UpdateFolder(folderID int, name string) error                    // the function ensures it exists
	MoveFolder(folderID int, destFolderID int) error                 // the function ensures it and parent exist
	MoveFile(fileID int, destFolderID int) error                     // the function ensures it and parent exist
	DeleteFolderAndContent(folderID int) error                       // the function ensures it exists
	DeleteFile(fileID int) error                                     // the function ensures it exists
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
	if err := svc.errorIfFolderNotFound(folderID); err != nil {
		return nil, err
	}

	folder, err := svc.repo.GetFolder(folderID)
	if err != nil {
		return nil, err
	}
	if folder == nil {
		// required since no tx mgmt => no guarantee folder was not deleted since previous check
		return nil, errors.New(NotFound)
	}
	return folder, err
}

func (svc FileSystemService) ExistsFolder(folderID int) (*bool, error) {
	return svc.repo.ExistsFolder(folderID)
}

func (svc FileSystemService) ExistsFile(folderID int) (*bool, error) {
	return svc.repo.ExistsFile(folderID)
}

func (svc FileSystemService) GetFile(fileID int) (*fsmodel.File, error) {
	if exists, err := svc.repo.ExistsFile(fileID); err != nil || exists == nil || *exists == false {
		return nil, errors.New(NotFound)
	}

	file, err := svc.repo.GetFile(fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		// this check is required since no transaction mgmt implemented => no guarantee the file was not deleted since previous check
		return nil, errors.New(NotFound)
	}
	return file, err
}

// TODO could have a single database call to return at the same time: currentFolder, folders, files
func (svc FileSystemService) GetFoldersIn(folderID int) (*[]fsmodel.Folder, error) {
	if err := svc.errorIfFolderNotFound(folderID); err != nil {
		return nil, errors.New(NotFound)
	}
	return svc.repo.GetFoldersIn(folderID)
}

func (svc FileSystemService) GetFilesIn(folderID int) (*[]fsmodel.File, error) {
	if err := svc.errorIfFolderNotFound(folderID); err != nil {
		return nil, errors.New(NotFound)
	}
	return svc.repo.GetFilesIn(folderID)
}

func (svc FileSystemService) CreateFolder(name string, parentID int) (*int, error) {
	if err := svc.errorIfFolderNotFound(parentID); err != nil {
		return nil, errors.WithMessage(
			errors.New(BadRequest),
			fmt.Sprintf("Not found folder specified (id=%d) when trying to create folder named %s inside.", parentID, name))
	}
	return svc.repo.CreateFolder(name, parentID)
}

func (svc FileSystemService) CreateFile(name string, path string, parentID int) (*int, error) {
	if err := svc.errorIfFolderNotFound(parentID); err != nil {
		return nil, errors.WithMessage(
			errors.New(BadRequest),
			fmt.Sprintf("Not found folder specified (id=%d) when trying to create file named %s inside.", parentID, name))
	}
	return svc.repo.CreateFile(name, path, parentID)
}

func (svc FileSystemService) UpdateFolder(folderID int, name string) error {
	if err := svc.errorIfFolderNotFound(folderID); err != nil {
		return errors.New(NotFound)
	}
	return svc.repo.UpdateFolder(folderID, name)
}

func (svc FileSystemService) MoveFolder(folderID int, destFolderID int) error {
	if err := svc.errorIfFolderNotFound(folderID); err != nil {
		return errors.WithMessage(
			errors.New(BadRequest),
			fmt.Sprintf("Could not find folder %d trying to be moved.", folderID))
	}
	if err := svc.errorIfFolderNotFound(destFolderID); err != nil {
		return errors.WithMessage(
			errors.New(BadRequest),
			fmt.Sprintf("Could not find destination folder %d where folder %d is trying to be moved.", destFolderID, folderID))
	}

	if isRoot, err := svc.repo.IsRootFolder(folderID); err != nil || isRoot == nil || *isRoot == true {
		return errors.WithMessage(
			errors.New(IllegalOperation),
			fmt.Sprintf("The root folder %d cannot be moved to any other folder. Attempted target folder %d.", destFolderID, folderID))
	}
	return svc.repo.MoveFolder(folderID, destFolderID)
}

func (svc FileSystemService) MoveFile(fileID int, destFolderID int) error {
	if err := svc.errorIfFileNotFound(fileID); err != nil {
		return errors.WithMessage(
			errors.New(BadRequest),
			fmt.Sprintf("Could not find file %d trying to be moved.", fileID))
	}
	if err := svc.errorIfFolderNotFound(destFolderID); err != nil {
		return errors.WithMessage(
			errors.New(BadRequest),
			fmt.Sprintf("Could not find destination folder %d where file %d is trying to be moved.", destFolderID, fileID))
	}

	return svc.repo.MoveFile(fileID, destFolderID)
}

func (svc FileSystemService) DeleteFolderAndContent(folderID int) error {
	if found, err := svc.repo.ExistsFolder(folderID); err != nil || found == nil || *found == false {
		return errors.New(NotFound)
	}

	if isRoot, err := svc.repo.IsRootFolder(folderID); err != nil || isRoot == nil || *isRoot == true {
		return errors.WithMessage(
			errors.New(IllegalOperation),
			fmt.Sprintf("Cannot delete root folder %d", folderID))
	}

	return svc.repo.DeleteFolderAndContent(folderID)
}

func (svc FileSystemService) DeleteFile(fileID int) error {
	if found, err := svc.repo.ExistsFile(fileID); err != nil || found == nil || *found == false {
		return errors.New("The file does not exist. It cannot be deleted.")
	}
	return svc.repo.DeleteFile(fileID)
}

func (svc FileSystemService) errorIfFileNotFound(fileID int) error {
	exists, err := svc.ExistsFile(fileID)
	if err != nil {
		return err
	}
	if *exists == false {
		return errors.New(NotFound)
	}
	return nil
}

func (svc FileSystemService) errorIfFolderNotFound(folderID int) error {
	exists, err := svc.ExistsFolder(folderID)
	if err != nil {
		return err
	}
	if *exists == false {
		return errors.New(NotFound)
	}
	return nil
}
