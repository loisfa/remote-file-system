package fsmanager

import (
	"errors"
)

func GetRootFolderID() (id *int, err error) {
	return DBGetRootFolderID()
}

func GetFolder(folderID int) (*Folder, error) {
	exists, err := DBExistsFolder(folderID)
	if err != nil {
		return nil, err
	}
	if !*exists {
		return nil, nil
	}

	folder, err := DBGetFolder(folderID)
	if err != nil {
		return nil, err
	}
	return folder, err
}

func ExistsFolder(folderID int) (*bool, error) {
	return DBExistsFolder(folderID)
}

func GetFile(fileID int) (*File, error) {
	if exists, err := DBExistsFile(fileID); err != nil || exists == nil || *exists == false {
		return nil, errors.New("The file does not exist. Cannot be fetched")
	}

	file, err := DBGetFile(fileID)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("Could not find file for the specified id")
	}

	return file, err
}

// TODO could have one single database call to return at the same time: currentFolder, folders, files
func GetFoldersIn(folderID int) (*[]Folder, error) {
	return DBGetFoldersIn(folderID)
}

func GetFilesIn(folderID int) (*[]File, error) {
	return DBGetFilesIn(folderID)
}

func CreateFolder(name string, parentID int) (*int, error) {
	return DBCreateFolder(name, parentID)
}

func CreateFile(name string, path string, parentID int) (*int, error) {
	return DBCreateFile(name, path, parentID)
}

func UpdateFolder(folderID int, name string) error {
	return DBUpdateFolder(folderID, name)
}

func MoveFolder(folderID int, destFolderID int) error {
	if found, err := DBExistsFolder(destFolderID); err != nil || found == nil || *found == false {
		return errors.New("The destination folder does not exist. Folder cannot be moved there.")
	}

	if found, err := DBExistsFolder(folderID); err != nil || found == nil || *found == false {
		return errors.New("The folder was not found.")
	}

	if isRoot, err := DBIsRootFolder(folderID); err != nil || isRoot == nil || *isRoot == true {
		return errors.New("Cannot perform 'Move' operation on the root folder")
	}

	return DBMoveFolder(folderID, destFolderID)
}

func MoveFile(fileID int, destFolderID int) error {
	if found, err := DBExistsFolder(destFolderID); err != nil || found == nil || *found == false {
		return errors.New("The destination folder does not exist. File cannot be moved there.")
	}
	return DBMoveFile(fileID, destFolderID)
}

func DeleteFolderAndContent(folderID int) error {
	if found, err := DBExistsFolder(folderID); err != nil || found == nil || *found == false {
		return errors.New("The folder does not exist. It cannot be deleted.")
	}

	if isRoot, err := DBIsRootFolder(folderID); err != nil || isRoot == nil || *isRoot == true {
		return errors.New("Trying to delete the root folder. Operation not permitted")
	}

	return DBDeleteFolderAndContent(folderID)
}

func DeleteFile(fileID int) error {
	if found, err := DBExistsFile(fileID); err != nil || found == nil || *found == false {
		return errors.New("The file does not exist. It cannot be deleted.")
	}
	return DBDeleteFile(fileID)
}
