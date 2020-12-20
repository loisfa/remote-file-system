package fsmanager

// TODO try neo4j to model the graph structure:
// https://stackoverflow.com/questions/31079881/simple-recursive-cypher-query

import (
	"errors"
	"fmt"
)

type DBFolder struct {
	Id       int // readonly
	Name     string
	ParentId *int // root folder <--> ParentId is nil
	// TODO isDelete boolean => soft delete
}

type DBFile struct {
	Id          int // readonly
	Name        string
	Path        string // readonly
	ParentId    int
}

var rootFolder = &DBFolder{0, "", nil}
var photosFolder = &DBFolder{1, "Photos", &(rootFolder.Id)}
var summerPhotosFolder = &DBFolder{2, "Summer", &(photosFolder.Id)}

var textFile1 = &DBFile{
	0,
	"file1.txt",
	"temp-files/file1.txt",
	photosFolder.Id}
var textFile2 = &DBFile{
	1,
	"file2.txt",
	"temp-files/file2.txt",
	rootFolder.Id}

var dbFoldersMap map[int]*DBFolder = make(map[int]*DBFolder)
var dbFilesMap map[int]*DBFile = make(map[int]*DBFile)

// For testing purpose
func GetDbFolderMap() map[int]*DBFolder {
	return dbFoldersMap
}
// For testing purpose
func GetDbFileMap() map[int]*DBFile {
	return dbFilesMap
}

var foldersAutoIncrementIndex int
var filesAutoIncrementIndex int

func InitDB() {
	// TODO reuse the existing 'createXxx' methods to perform those operations
	dbFoldersMap[rootFolder.Id] = rootFolder
	dbFoldersMap[photosFolder.Id] = photosFolder
	dbFoldersMap[summerPhotosFolder.Id] = summerPhotosFolder
	foldersAutoIncrementIndex = 3
	
	dbFilesMap[textFile1.Id] = textFile1
	dbFilesMap[textFile2.Id] = textFile2
	filesAutoIncrementIndex = 2

	return
}

func DBCreateFolder(name string, parentId int) int {
	folderId := foldersAutoIncrementIndex
	dbFoldersMap[folderId] = &DBFolder{folderId, name, &parentId}
	foldersAutoIncrementIndex = foldersAutoIncrementIndex + 1
	return folderId
}

func DBCreateFile(name string, path string, parentId int) int {
	fileId := filesAutoIncrementIndex
	dbFilesMap[fileId] = &DBFile{fileId, name, path, parentId}
	filesAutoIncrementIndex = filesAutoIncrementIndex + 1
	return fileId
}

func DBUpdateFolder(folderId int, name string) error {
	toUpdateFolder, ok := dbFoldersMap[folderId]

	if ok == false {
		return errors.New("Could not find folder for specified id")
	}

	toUpdateFolder.Name = name
	return nil
}

func DBMoveFolder(folderId int, targetParentId *int) error {
	toUpdateFolder, ok := dbFoldersMap[folderId]
	if ok == false {
		return errors.New("Could not find folder for specified id")
	}

	toUpdateFolder.ParentId = targetParentId
	return nil
}

func DBMoveFile(fileId int, targetParentId *int) error {
	toUpdateFile, ok := dbFilesMap[fileId]
	if ok == false {
		return errors.New("Could not find file for specified id")
	}

	toUpdateFile.ParentId = *targetParentId
	return nil
}

// could use goroutines with recursive?
func getContentToDelete(folderId int) (toDeleteFolderIds []int, toDeleteFileIds []int) {

	toDeleteFolderIds = []int{}

	toDeleteFileIds = []int{}

	for _, folder := range dbFoldersMap {
		if folder.ParentId != nil && *(folder.ParentId) == folderId {
			toDeleteFolderIds = append(toDeleteFolderIds, folder.Id)
			break
		}
	}

	for _, file := range dbFilesMap {
		if file.ParentId == folderId {
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
		delete(dbFoldersMap, folderId)
	}
}

func removeFiles(fileIds []int) {
	for _, fileId := range fileIds {
		file, ok := dbFilesMap[fileId]
		if ok == true {
			fmt.Println("Deleting file ", file.Path) // TODO: delete the file actually from the file storage? (for now soft delete)
			delete(dbFilesMap, fileId)
		}
	}
}

func DBDeleteFolderAndContent(folderId int) error {
	fmt.Println("Deleting db folder (and chidren)", folderId)
	_, ok := dbFoldersMap[folderId]

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


func DBDeleteFile(fileId int) error {
	fmt.Println("Deleting file", fileId)
	_, ok := dbFilesMap[fileId]

	if ok == false {
		return errors.New("Could not find folder for specified id") // find a way to fire 404
	}

	toDeleteFileIds := []int{fileId}
	removeFiles(toDeleteFileIds)

	return nil
}
