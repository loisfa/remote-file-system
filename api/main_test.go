package main

import (
	"runtime/debug"
	"testing"
)

/*
// TODO: this is not a db check (it uses Api models)
func TestDbFolderContent(t *testing.T) {
	// fsmanager.InitDB()

	folderIdToFolderMap := fsmanager.GetDbFolderMap()
	rootFolder := folderIdToFolderMap[0] // TODO issue 0 is a magic number

	var rootFolderContent ApiFolderContent = DBGetContentIn(rootFolder.Id)
	assertEqual(t, rootFolderContent.CurrentFolder.Id, rootFolder.Id)
	assertEqual(t, rootFolderContent.CurrentFolder.Name, rootFolder.Name)
	assertEqual(t, rootFolderContent.CurrentFolder.ParentId, rootFolder.ParentId)

	assertEqual(t, len(rootFolderContent.Folders), 1)
	assertEqual(t, rootFolderContent.Folders[0].Id, 1)
	assertEqual(t, rootFolderContent.Folders[0].Name, "Photos")
	assertEqual(t, *rootFolderContent.Folders[0].ParentId, rootFolder.Id)

	assertEqual(t, len(rootFolderContent.Files), 1)
	assertEqual(t, rootFolderContent.Files[0].Id, 1)
	assertEqual(t, rootFolderContent.Files[0].Name, "file2.txt")
}
*/

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
