package fsservice

import (
	"errors"

	"github.com/loisfa/remote-file-system/api/fsmodel"
	"github.com/loisfa/remote-file-system/api/fsrepository"
)

func GetRootFolderID() (id *int, err error) {
	return fsrepository.GetRootFolderID()
}

func GetFolder(folderID int) (*fsmodel.Folder, error) {
	exists, err := fsrepository.ExistsFolder(folderID)
	if err != nil {
		return nil, err
	}
	if !*exists {
		return nil, nil
	}

	folder, err := fsrepository.GetFolder(folderID)
	if err != nil {
		return nil, err
	}
	return folder, err
}

func ExistsFolder(folderID int) (*bool, error) {
	return fsrepository.ExistsFolder(folderID)
}

func GetFile(fileID int) (*fsmodel.File, error) {
	if exists, err := fsrepository.ExistsFile(fileID); err != nil || exists == nil || *exists == false {
		return nil, errors.New("The file does not exist. Cannot be fetched")
	}

	file, err := fsrepository.GetFile(fileID)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("Could not find file for the specified id")
	}

	return file, err
}

// TODO could have one single database call to return at the same time: currentFolder, folders, files
func GetFoldersIn(folderID int) (*[]fsmodel.Folder, error) {
	return fsrepository.GetFoldersIn(folderID)
}

func GetFilesIn(folderID int) (*[]fsmodel.File, error) {
	return fsrepository.GetFilesIn(folderID)
}

func CreateFolder(name string, parentID int) (*int, error) {
	return fsrepository.CreateFolder(name, parentID)
}

func CreateFile(name string, path string, parentID int) (*int, error) {
	return fsrepository.CreateFile(name, path, parentID)
}

func UpdateFolder(folderID int, name string) error {
	return fsrepository.UpdateFolder(folderID, name)
}

func MoveFolder(folderID int, destFolderID int) error {
	if found, err := fsrepository.ExistsFolder(destFolderID); err != nil || found == nil || *found == false {
		return errors.New("The destination folder does not exist. Folder cannot be moved there.")
	}

	if found, err := fsrepository.ExistsFolder(folderID); err != nil || found == nil || *found == false {
		return errors.New("The folder was not found.")
	}

	if isRoot, err := fsrepository.IsRootFolder(folderID); err != nil || isRoot == nil || *isRoot == true {
		return errors.New("Cannot perform 'Move' operation on the root folder")
	}

	return fsrepository.MoveFolder(folderID, destFolderID)
}

func MoveFile(fileID int, destFolderID int) error {
	if found, err := fsrepository.ExistsFolder(destFolderID); err != nil || found == nil || *found == false {
		return errors.New("The destination folder does not exist. File cannot be moved there.")
	}
	return fsrepository.MoveFile(fileID, destFolderID)
}

func DeleteFolderAndContent(folderID int) error {
	if found, err := fsrepository.ExistsFolder(folderID); err != nil || found == nil || *found == false {
		return errors.New("The folder does not exist. It cannot be deleted.")
	}

	if isRoot, err := fsrepository.IsRootFolder(folderID); err != nil || isRoot == nil || *isRoot == true {
		return errors.New("Trying to delete the root folder. Operation not permitted")
	}

	return fsrepository.DeleteFolderAndContent(folderID)
}

func DeleteFile(fileID int) error {
	if found, err := fsrepository.ExistsFile(fileID); err != nil || found == nil || *found == false {
		return errors.New("The file does not exist. It cannot be deleted.")
	}
	return fsrepository.DeleteFile(fileID)
}
