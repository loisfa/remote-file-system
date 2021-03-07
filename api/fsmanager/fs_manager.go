package fsmanager

// TODO try neo4j to model the graph structure:
// https://stackoverflow.com/questions/31079881/simple-recursive-cypher-query

import (
	"errors"
	"fmt"
)

var rootFolder = &Folder{
	Id:       0,
	Name:     "",
	ParentId: nil,
}
var photosFolder = &Folder{
	Id:       1,
	Name:     "Photos",
	ParentId: &(rootFolder.Id),
}
var summerPhotosFolder = &Folder{
	Id:       2,
	Name:     "Summer",
	ParentId: &(photosFolder.Id),
}

var textFile1 = &File{
	Id:       0,
	Name:     "file1.txt",
	Path:     "temp-files/file1.txt",
	ParentId: &photosFolder.Id,
}
var textFile2 = &File{
	Id:       1,
	Name:     "file2.txt",
	Path:     "temp-files/file2.txt",
	ParentId: &rootFolder.Id,
}

var FoldersMap map[int]*Folder = make(map[int]*Folder)
var FilesMap map[int]*File = make(map[int]*File)

// public for testing purpose
func GetDbFolderMap() map[int]*Folder {
	return FoldersMap
}

// public for testing purpose
func GetDbFileMap() map[int]*File {
	return FilesMap
}

var foldersAutoIncrementIndex int
var filesAutoIncrementIndex int

func InitDB() {
	// TODO reuse the existing 'createXxx' methods to perform those operations
	FoldersMap[rootFolder.Id] = rootFolder
	FoldersMap[photosFolder.Id] = photosFolder
	FoldersMap[summerPhotosFolder.Id] = summerPhotosFolder
	foldersAutoIncrementIndex = 3

	FilesMap[textFile1.Id] = textFile1
	FilesMap[textFile2.Id] = textFile2
	filesAutoIncrementIndex = 2

	return
}

// TODO rename all the DBGet... --> Get... (remove "DB" prefix)
func DBGetFolder(folderId int) (*Folder, error) {
	fmt.Println("DBGetFolder: Getting folder in %d", folderId)
	exists, err := ExistsFolder(folderId)
	fmt.Println("DBGetFolder: exists: ", exists, "-- err: ", err)
	if err != nil {
		return nil, err
	}
	if !*exists {
		return nil, nil
	}

	folder, err := GetFolder(folderId)
	if err != nil {
		return nil, err
	}
	return folder, err
}

func DBExistsFolder(folderId int) (*bool, error) {
	return ExistsFolder(folderId)
}

func DBGetFile(fileId int) (*File, error) {
	file, err := GetFile(fileId)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("Could not find file for specified id")
	}

	return file, err
}

func DBGetFoldersIn(folderId *int) (*[]Folder, error) {
	// TODO check that the folder exist
	return GetFoldersIn(folderId)
}

func DBGetFilesIn(folderId *int) (*[]File, error) {
	// TODO check that the folder exist
	return GetFilesIn(folderId)
}

func DBCreateFolder(name string, parentId *int) (*int, error) {
	return CreateFolder(name, parentId)
}

func DBCreateFile(name string, path string, parentId *int) (*int, error) {
	return CreateFile(name, path, parentId)
}

// TODO
func DBUpdateFolder(folderId int, name string) error {
	toUpdateFolder, ok := FoldersMap[folderId]

	if ok == false {
		return errors.New("Could not find folder for specified id")
	}

	toUpdateFolder.Name = name
	return nil
}

// TODO
func DBMoveFolder(folderId int, targetParentId *int) error {
	if folderId == rootFolder.Id {
		return errors.New("Illegal Operation: trying to move root folder")
	}

	toUpdateFolder, ok := FoldersMap[folderId]
	if ok == false {
		return errors.New("Could not find folder for specified id")
	}

	toUpdateFolder.ParentId = targetParentId
	return nil
}

func DBMoveFile(fileId int, targetParentId *int) error {
	toUpdateFile, ok := FilesMap[fileId]

	if ok == false {
		return errors.New("Could not find file for specified id")
	}

	toUpdateFile.ParentId = targetParentId
	return nil
}

// TODO
// could use goroutines with recursive?
func getContentToDelete(folderId int) (toDeleteFolderIds []int, toDeleteFileIds []int) {

	toDeleteFolderIds = []int{}

	toDeleteFileIds = []int{}

	for _, folder := range FoldersMap {
		if folder.ParentId != nil && *(folder.ParentId) == folderId {
			toDeleteFolderIds = append(toDeleteFolderIds, folder.Id)
			break
		}
	}

	for _, file := range FilesMap {
		if file.ParentId != nil && *file.ParentId == folderId {
			toDeleteFileIds = append(toDeleteFileIds, file.Id)
			break
		}
	}

	var innerToDeleteFolderIds, innerToDeleteFileIds []int
	for _, id := range toDeleteFolderIds {
		innerToDeleteFolderIds, innerToDeleteFileIds = getContentToDelete(id)
	}

	toDeleteFolderIds = append(toDeleteFolderIds, innerToDeleteFolderIds...)
	toDeleteFileIds = append(toDeleteFileIds, innerToDeleteFileIds...)

	return toDeleteFolderIds, toDeleteFileIds
}

func removeFolders(folderIds []int) {
	for _, folderId := range folderIds {
		delete(FoldersMap, folderId)
	}
}

func removeFiles(fileIds []int) {
	for _, fileId := range fileIds {
		file, ok := FilesMap[fileId]
		if ok == true {
			fmt.Println("Deleting file", file.Path) // TODO: delete the file actually from the file storage? (for now soft delete)
			delete(FilesMap, fileId)
		}
	}
}

// TODO
func DBDeleteFolderAndContent(folderId int) error {
	if folderId == rootFolder.Id {
		return errors.New("Illegal Operation: trying to delete root folder.")
	}

	fmt.Println("Deleting db folder (and chidren)", folderId)
	_, ok := FoldersMap[folderId]

	if ok == false {
		return errors.New("Could not find folder for specified id")
	}

	toDeleteFolderIds, toDeleteFileIds := getContentToDelete(folderId)
	toDeleteFolderIds = append(toDeleteFolderIds, folderId)

	fmt.Println("Folders to remove", len(toDeleteFolderIds))
	fmt.Println("Files to remove", len(toDeleteFileIds))

	removeFiles(toDeleteFileIds)
	removeFolders(toDeleteFolderIds)

	return nil
}

// TODO
func DBDeleteFile(fileId int) error {
	fmt.Println("Deleting file", fileId)
	_, ok := FilesMap[fileId]

	if ok == false {
		return errors.New("Could not find folder for specified id") // find a way to fire 404
	}

	toDeleteFileIds := []int{fileId}
	removeFiles(toDeleteFileIds)

	return nil
}
