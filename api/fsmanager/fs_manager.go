package fsmanager

import (
	"errors"
	"fmt"
)

// TODO rename all the DBGet... --> Get... (remove "DB" prefix)
func DBGetRootFolderID() (id *int, err error) {
	return GetRootFolderID()
}

// TODO rename all the DBGet... --> Get... (remove "DB" prefix)
func DBGetFolder(folderID int) (*Folder, error) {
	fmt.Println("DBGetFolder: Getting folder in %d", folderID)
	exists, err := ExistsFolder(folderID)
	fmt.Println("DBGetFolder: exists: ", exists, "-- err: ", err)
	if err != nil {
		return nil, err
	}
	if !*exists {
		return nil, nil
	}

	folder, err := GetFolder(folderID)
	if err != nil {
		return nil, err
	}
	return folder, err
}

func DBExistsFolder(folderID int) (*bool, error) {
	return ExistsFolder(folderID)
}

func DBGetFile(fileID int) (*File, error) {
	file, err := GetFile(fileID)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("Could not find file for the specified id")
	}

	return file, err
}

// TODO could have one single database call to return currentFolder, folders, files
func DBGetFoldersIn(folderID int) (*[]Folder, error) {
	return GetFoldersIn(folderID)
}

func DBGetFilesIn(folderID int) (*[]File, error) {
	return GetFilesIn(folderID)
}

func DBCreateFolder(name string, parentID int) (*int, error) {
	return CreateFolder(name, parentID)
}

func DBCreateFile(name string, path string, parentID int) (*int, error) {
	return CreateFile(name, path, parentID)
}

func DBUpdateFolder(folderID int, name string) error {
	return UpdateFolder(folderID, name)
}

func DBMoveFolder(folderID int, destFolderID int) error {
	if found, err := ExistsFolder(destFolderID); err != nil || found != nil && *found == false {
		return errors.New("The destination folder does not exist. Folder cannot be moved there.")
	}

	if found, err := ExistsFolder(folderID); err != nil || found != nil && *found == false {
		return errors.New("The folder was not found.")
	}

	fmt.Println("DBMoveFolder about to check")
	if isRoot, err := IsRootFolder(folderID); err != nil || isRoot != nil && *isRoot == true {
		return errors.New("Cannot perform 'Move' operation on the root folder")
	}

	return MoveFolder(folderID, destFolderID)
}

func DBMoveFile(fileID int, destFolderID int) error {
	if found, err := ExistsFolder(destFolderID); err != nil || found != nil && *found == false {
		return errors.New("The destination folder does not exist. File cannot be moved there.")
	}
	return MoveFile(fileID, destFolderID)
}

func DBDeleteFolderAndContent(folderID int) error {
	if found, err := ExistsFolder(folderID); err != nil || found != nil && *found == false {
		return errors.New("The folder does not exist. It cannot be deleted.")
	}

	if isRoot, err := IsRootFolder(folderID); err != nil || isRoot != nil && *isRoot == true {
		return errors.New("Trying to delete the root folder. Operation not permitted")
	}

	return DeleteFolderContent(folderID)
}

func DBDeleteFile(fileID int) error {
	if found, err := ExistsFile(fileID); err != nil || found != nil && *found == false {
		return errors.New("The file does not exist. It cannot be deleted.")
	}
	return DeleteFile(fileID)
}
