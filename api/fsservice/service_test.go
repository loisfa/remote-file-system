package fsservice

import (
	"runtime/debug"
	"testing"
)

func TestDbInit(t *testing.T) {
	folderIdToFolderMap := GetDbFolderMap()

	if folderIdToFolderMap == nil {
		t.Errorf("folderIdToFolderMap is nil")
	}

	var rootFolder = folderIdToFolderMap[0]
	var photosFolder = folderIdToFolderMap[1]
	var summerPhotosFolder = folderIdToFolderMap[2]

	if rootFolder != nil {
		t.Errorf("folderIdToFolderMap is not nil before db init")
	}
	if photosFolder != nil {
		t.Errorf("photosFolder is not nil before db init")
	}
	if summerPhotosFolder != nil {
		t.Errorf("summerPhotosFolder is not nil before db init")
	}

	InitDB()

	rootFolder = folderIdToFolderMap[0]
	photosFolder = folderIdToFolderMap[1]
	summerPhotosFolder = folderIdToFolderMap[2]
	if rootFolder == nil {
		t.Errorf("folderIdToFolderMap is nil after db init")
	}
	if photosFolder == nil {
		t.Errorf("photosFolder is nil after db init")
	}
	if summerPhotosFolder == nil {
		t.Errorf("summerPhotosFolder is nil after db init")
	}
}

func TestDbCreateFolder(t *testing.T) {
	InitDB()

	var newFolderId int = DBCreateFolder("New folder", 0)

	folderIdToFolderMap := GetDbFolderMap()

	var createdFolder *DBFolder = folderIdToFolderMap[newFolderId]

	assertEqual(t, createdFolder.Id, newFolderId)
	assertEqual(t, createdFolder.Name, "New folder")
}

func TestDbUpdateFolder(t *testing.T) {
	InitDB()

	var newFolderId int = DBCreateFolder("New folder", 0)

	folderIdToFolderMap := GetDbFolderMap()

	var createdFolder *DBFolder = folderIdToFolderMap[newFolderId]
	assertEqual(t, createdFolder.Id, newFolderId)
	assertEqual(t, createdFolder.Name, "New folder")

	DBUpdateFolder(newFolderId, "New name for folder")
	var updatedFolder *DBFolder = folderIdToFolderMap[newFolderId]
	assertEqual(t, updatedFolder.Id, newFolderId)
	assertEqual(t, updatedFolder.Name, "New name for folder")
}

func TestDbMoveFolder(t *testing.T) {
	InitDB()

	folderIdToFolderMap := GetDbFolderMap()
	var rootFolder *DBFolder = folderIdToFolderMap[0]

	var newFolderId int = DBCreateFolder("New folder", rootFolder.Id)

	var createdFolder *DBFolder = folderIdToFolderMap[newFolderId]
	assertEqual(t, createdFolder.Id, newFolderId)
	assertEqual(t, createdFolder.Name, "New folder")
	assertEqual(t, *createdFolder.ParentId, rootFolder.Id)

	var targetFolder *DBFolder = folderIdToFolderMap[1]
	assertNotNil(t, targetFolder)

	DBMoveFolder(newFolderId, &targetFolder.Id)
	assertEqual(t, createdFolder.Id, newFolderId)
	assertEqual(t, createdFolder.Name, "New folder")
	assertEqual(t, *createdFolder.ParentId, targetFolder.Id)
}

func TestDbDeleteFolderAndContent(t *testing.T) {
	InitDB()

	var newFolderId int = DBCreateFolder("New folder", 0)
	var newInnerFolderId int = DBCreateFolder("New inner folder", newFolderId)
	var newFileId int = DBCreateFile("New file", "./my-path/file1.txt", newFolderId)

	folderIdToFolderMap := GetDbFolderMap()
	fileIdToFileMap := GetDbFileMap()

	assertNotNil(t, folderIdToFolderMap[newFolderId])
	assertNotNil(t, folderIdToFolderMap[newInnerFolderId])
	assertNotNil(t, fileIdToFileMap[newFileId])

	DBDeleteFolderAndContent(newFolderId)

	assertNil(t, folderIdToFolderMap[newFolderId])
	assertNil(t, folderIdToFolderMap[newInnerFolderId])
	assertNil(t, fileIdToFileMap[newFileId])
}

func TestDbCreateFile(t *testing.T) {
	InitDB()

	var newFileId int = DBCreateFile("New file", "/my-path/file5.txt", 0)

	fileIdToFileMap := GetDbFileMap()

	var createdFile *DBFile = fileIdToFileMap[newFileId]

	assertEqual(t, createdFile.Id, newFileId)
	assertEqual(t, createdFile.Name, "New file")
	assertEqual(t, createdFile.Path, "/my-path/file5.txt")
}

func TestDbDeleteFile(t *testing.T) {
	InitDB()

	var newFileId int = DBCreateFile("New file", "/path/word.txt", 0)

	fileIdToFileMap := GetDbFileMap()

	var createdFile *DBFile = fileIdToFileMap[newFileId]
	assertEqual(t, createdFile.Id, newFileId)

	DBDeleteFile(newFileId)

	assertNil(t, fileIdToFileMap[newFileId])
}

func TestDbMoveFile(t *testing.T) {
	InitDB()

	var newFileId int = DBCreateFile("New file", "/path/word.txt", 0)

	fileIdToFileMap := GetDbFileMap()
	folderIdToFolderMap := GetDbFolderMap()

	var createdFile *DBFile = fileIdToFileMap[newFileId]
	assertEqual(t, createdFile.Id, newFileId)
	assertEqual(t, createdFile.ParentId, 0)

	var existingFolder *DBFolder = folderIdToFolderMap[1]
	assertNotNil(t, existingFolder)

	DBMoveFile(newFileId, &existingFolder.Id)
	assertEqual(t, createdFile.Id, newFileId)
	assertEqual(t, createdFile.ParentId, existingFolder.Id)
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Log(string(debug.Stack()))
		t.Fatalf("%s != %s", a, b)
	}
}

func assertNotNil(t *testing.T, a interface{}) {
	if a == nil {
		t.Log(string(debug.Stack()))
		t.Fatalf("%s == nil", a)
	}
}

func assertNil(t *testing.T, a interface{}) {
	if a == nil {
		t.Log(string(debug.Stack()))
		t.Fatalf("%s != nil", a)
	}
}
