package fsmanager

import (
	"errors"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
)

// TODO gracefully shutdown the go application:
// https://medium.com/@BastianRob/gracefully-shutdown-your-go-application-2ef2871025f0
// defer driver.Close()

// TODO beware database injection (string params)

// Item can be a folder or a file
type createdItem struct {
	Id int
}

// Do not use the driver directly, please use the getDriver method
var driver *neo4j.Driver

const (
	uri      = "neo4j://localhost:7687" // TODO
	username = "neo4j"                  // TODO not commit
	password = "password"               // TODO not commit

	dbId     = "id"
	dbName   = "name"
	dbPath   = "path"
	dbFolder = "folder"
	dbFile   = "file"
	dbExists = "exists"
)

func mapRecordToFile(record *neo4j.Record) (*File, error) {
	fmt.Println("record: %+v", record)

	file, _ := record.Get(dbFile)
	fileProps := (file.(dbtype.Node)).Props
	fmt.Printf("%+v\n", fileProps)

	id, found := fileProps[dbId]
	if !found {
		return nil, errors.New("Could not retrieve file id from DB")
	}
	name, found := fileProps[dbName]
	if !found {
		return nil, errors.New("Could not retrieve file name from DB")
	}
	path, found := fileProps[dbPath]
	if !found {
		return nil, errors.New("Could not retrieve file path from DB")
	}

	fmt.Println("file name:", name)
	fmt.Println("file path:", path)
	fmt.Println("file id:", id)
	return &File{
		Id:   int(id.(int64)),
		Name: name.(string),
		Path: path.(string),
	}, nil
}

func mapResultToFiles(result neo4j.Result) (*[]File, error) {
	var files []File
	for result.Next() == true {
		record := result.Record()

		file, err := mapRecordToFile(record)
		if err != nil {
			return nil, err
		}

		files = append(files, *file)
	}
	return &files, nil
}

func mapRecordToFolder(record *neo4j.Record) (*Folder, error) {
	fmt.Println("record: %+v", record)

	folder, _ := record.Get(dbFolder)
	folderProps := (folder.(dbtype.Node)).Props
	fmt.Printf("%+v\n", folderProps)

	id, found := folderProps[dbId]
	if !found {
		return nil, errors.New("Could not retrieve folder id from DB")
	}
	name, found := folderProps[dbName]
	if !found {
		return nil, errors.New("Could not retrieve folder name from DB")
	}

	fmt.Println("folder name:", name)
	fmt.Println("folder id:", id)
	return &Folder{
		Id:   int(id.(int64)),
		Name: name.(string),
	}, nil
}

func mapResultToFolders(result neo4j.Result) (*[]Folder, error) {
	var folders []Folder
	fmt.Println("mapResultToFolders input result:")
	fmt.Println(result)
	for result.Next() == true {
		fmt.Println("result.Next() == true")
		record := result.Record()
		fmt.Println("record %v", record)

		folder, err := mapRecordToFolder(record)
		if err != nil {
			return nil, err
		}

		folders = append(folders, *folder)
		fmt.Println("folders %v", folders)
	}
	fmt.Println("about to return folders: %v", folders)
	return &folders, nil
}

func getFileByID(fileID int) (string, map[string]interface{}, func(result neo4j.Result) (*File, error)) {
	return `MATCH (file:File{id: $fileID})
		RETURN file`,
		map[string]interface{}{
			"fileID": fileID,
		},
		func(result neo4j.Result) (*File, error) {
			record, err := result.Single()
			if err != nil {
				return nil, err
			}

			return mapRecordToFile(record)
		}
}

func existsFileByID(fileID int) (string, map[string]interface{}, func(result neo4j.Result) (*bool, error)) {
	return `OPTIONAL MATCH (file:File{id: $fileID})
		RETURN file IS NOT NULL AS exists`,
		map[string]interface{}{
			"fileID": fileID,
		},
		func(result neo4j.Result) (*bool, error) {
			record, err := result.Single()
			if err != nil {
				return nil, err
			}

			exists, found := record.Get(dbExists)
			if !found {
				return nil, errors.New("Could not find 'exists' in file exists response")
			}

			e := exists.(bool)
			return &e, nil
		}
}

func getFolderByID(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*Folder, error)) {
	return `MATCH (folder:Folder{id: $folderID})
		RETURN folder`,
		map[string]interface{}{
			"folderID": folderID,
		},
		func(result neo4j.Result) (*Folder, error) {
			record, err := result.Single()
			if err != nil {
				return nil, err
			}

			return mapRecordToFolder(record)
		}
}

func existsFolderByID(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*bool, error)) {
	return `OPTIONAL MATCH (folder:Folder{id: $folderID})
		RETURN folder IS NOT NULL AS exists`,
		map[string]interface{}{
			"folderID": folderID,
		},
		func(result neo4j.Result) (*bool, error) {
			fmt.Println("existsFolderByID fn: result:", result)
			record, err := result.Single()
			fmt.Println("existsFolderByID fn: record:", record)
			if err != nil {
				return nil, err
			}

			exists, found := record.Get(dbExists)
			if !found {
				return nil, errors.New("Could not find 'exists' in folder exists response")
			}

			e := exists.(bool)
			return &e, nil
		}
}

func getRootFilesQuery() (string, map[string]interface{}, func(result neo4j.Result) (*[]File, error)) {
	return `MATCH (file:File)
		WHERE NOT (file)-[:IS_INSIDE]->(:Folder)
		RETURN file`, make(map[string]interface{}), mapResultToFiles
}

func getRootFoldersQuery() (string, map[string]interface{}, func(result neo4j.Result) (*[]Folder, error)) {
	return `MATCH (folder:Folder)
		WHERE NOT (folder)-[:IS_INSIDE]->(:Folder)
		RETURN folder`, make(map[string]interface{}), mapResultToFolders
}

func getFilesInFolderQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*[]File, error)) {
	return `MATCH (parentFolder:Folder{id: $folderID})
	MATCH (file:File)-[:IS_INSIDE]->(parentFolder)
	RETURN file`,
		map[string]interface{}{
			"folderID": folderID,
		}, mapResultToFiles
}

func getFoldersInFolderQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*[]Folder, error)) {
	return `MATCH (parentFolder:Folder{id: $folderID})
	MATCH (folder:Folder)-[:IS_INSIDE]->(parentFolder)
	RETURN folder`,
		map[string]interface{}{
			"folderID": folderID,
		},
		mapResultToFolders
}

func createNewFileWithParentQuery(fileName string, filePath string, parentFolderID int) (string, map[string]interface{}) {
	return `MATCH (parentFolder:Folder{id: $parentFolderID})
	MATCH (seq:Sequence {key:'file_id_sequence'})
	CALL apoc.atomic.add(seq, 'value', 1, 5)
	YIELD newValue as file_id
	CREATE (file:File { id: file_id, name: $fileName, path: $filePath})
	CREATE (file)-[:IS_INSIDE]->(parentFolder)
	RETURN file.id AS fileID`,
		map[string]interface{}{
			"fileName":       fileName,
			"filePath":       filePath,
			"parentFolderID": parentFolderID,
		}
}

func createNewFileWithoutParentQuery(fileName string, filePath string) (string, map[string]interface{}) {
	return `MATCH (seq:Sequence {key:'file_id_sequence'})
	CALL apoc.atomic.add(seq, 'value', 1, 5)
	YIELD newValue as file_id
	CREATE (file:File { id: file_id, name: $fileName, path: $filePath})
	RETURN file.id AS fileID`,
		map[string]interface{}{
			"fileName": fileName,
			"filePath": filePath,
		}
}

func createNewFolderWithParentQuery(folderName string, parentFolderID int) (string, map[string]interface{}) {
	return `MATCH (parentFolder:Folder{id: $parentFolderID})
	MATCH (seq:Sequence {key:'folder_id_sequence'})
	CALL apoc.atomic.add(seq, 'value', 1, 5)
	YIELD newValue as folder_id
	CREATE (folder:Folder { id: folder_id, name: $folderName})
	CREATE (folder)-[:IS_INSIDE]->(parentFolder)
	RETURN folder.id AS folderID`,
		map[string]interface{}{
			"folderName":     folderName,
			"parentFolderID": parentFolderID,
		}
}

func createNewFolderWithoutParentQuery(folderName string) (string, map[string]interface{}) {
	return `MATCH (seq:Sequence {key:'folder_id_sequence'})
	CALL apoc.atomic.add(seq, 'value', 1, 5)
	YIELD newValue as folder_id
	CREATE (folder:Folder { id: folder_id, name: $folderName})
	RETURN folder.id AS folderID`,
		map[string]interface{}{
			"folderName": folderName,
		}
}

func updateFolderQuery(folderID int, folderName string) (string, map[string]interface{}) {
	return `MATCH (folder:Folder {id: $folderID})
	SET folder.name = $folderName`,
		map[string]interface{}{
			"folderID":   folderID,
			"folderName": folderName,
		}
}

func updateFileQuery(fileId int, fileName string) (string, map[string]interface{}) {
	return `MATCH (file:File {id: $fileId})
	SET file.name = $fileName`,
		map[string]interface{}{
			"fileId":   fileId,
			"fileName": fileName,
		}
}

func moveFolderToRootFolderQuery(folderId int) (string, map[string]interface{}) {
	return `MATCH (folder:Folder {id: $folderId})
	OPTIONAL MATCH (folder)-[rel:IS_INSIDE]->()
	DELETE rel`,
		map[string]interface{}{
			"folderId": folderId,
		}
}

func moveFolderQuery(folderID int, destFolderID int) (string, map[string]interface{}) {
	return `MATCH (folder:Folder {id: $folderID})
	MATCH (dest:Folder {id: $destFolderID})
	OPTIONAL MATCH (folder)-[rel:IS_INSIDE]->()
	DELETE rel
	CREATE (folder)-[:IS_INSIDE]->(dest)`,
		map[string]interface{}{
			"folderID":     folderID,
			"destFolderID": destFolderID,
		}
}

func moveFileQuery(fileID int, destFolderID int) (string, map[string]interface{}) {
	return `MATCH (file:File {id: $fileID})
	MATCH (dest:Folder{id: $destFolderID})
	OPTIONAL MATCH (file)-[rel:IS_INSIDE]->()
	DELETE rel
	CREATE (file)-[:IS_INSIDE]->(dest)`,
		map[string]interface{}{
			"fileID":       fileID,
			"destFolderID": destFolderID,
		}
}

func moveFileToRootFolderQuery(fileID int) (string, map[string]interface{}) {
	return `MATCH (file:File {id: $fileID})
	OPTIONAL MATCH (file)-[rel:IS_INSIDE]->()
	DELETE rel`,
		map[string]interface{}{
			"fileID": fileID,
		}
}

// Return the ids of all the deleted items?
func deleteFolderContentQuery(folderID int) (string, map[string]interface{}) {
	return `OPTIONAL MATCH (files:File)-[:IS_INSIDE *1..]->(Folder {id: $folderID})
	OPTIONAL MATCH (folders:Folder)-[:IS_INSIDE *1..]->(Folder {id: $folderID})
	OPTIONAL MATCH (folder:Folder {id: $folderID})
	DETACH DELETE files, folders, folder`,
		map[string]interface{}{
			"folderID": folderID,
		}
}

// Return the ids of all the deleted items?
func deleteFileQuery(fileID int) (string, map[string]interface{}) {
	return `MATCH (file:File {id: $fileID})
	DETACH DELETE file`,
		map[string]interface{}{
			"fileID": fileID,
		}
}

// InitDriver returns a valid driver
// handles driver lifetime based on your application lifetime requirements  driver's lifetime is usually
// bound by the application lifetime, which usually implies one driver instance per application
func initDriver() neo4j.Driver {
	// Neo4j 4.0, defaults to no TLS therefore use bolt:// or neo4j://
	// Neo4j 3.5, defaults to self-signed certificates, TLS on, therefore use bolt+ssc:// or neo4j+ssc://
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		fmt.Println("Error on driver creation")
		panic(err)
	}

	err = driver.VerifyConnectivity()
	if err != nil {
		fmt.Println("Error on driver connectivity check")
		panic(err)
	}

	fmt.Println("Driver configuration OK")
	return driver
}

func getDriver() neo4j.Driver {
	if driver == nil {
		d := initDriver()
		driver = &d
	}

	return *driver
}

func GetFile(fileID int) (*File, error) {
	session := getDriver().NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToFileFn := getFileByID(fileID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			fmt.Println("Error on transaction run")
			return nil, err
		}
		return mapResultToFileFn(result)
	})

	if err != nil {
		fmt.Println("Transaction failure")
		return nil, err
	}

	return result.(*File), nil
}

func ExistsFile(fileID int) (*bool, error) {
	session := getDriver().NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToExistFn := existsFileByID(fileID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			fmt.Println("Error on transaction run")
			return nil, err
		}
		return mapResultToExistFn(result)
	})

	if err != nil {
		fmt.Println("Transaction failure")
		return nil, err
	}

	return result.(*bool), nil
}

func GetFilesIn(folderID *int) (*[]File, error) {
	session := getDriver().NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var query string
		var queryMap map[string]interface{}
		var mapResultToFilesFn func(neo4j.Result) (*[]File, error)
		if folderID == nil {
			query, queryMap, mapResultToFilesFn = getRootFilesQuery()
		} else {
			query, queryMap, mapResultToFilesFn = getFilesInFolderQuery(*folderID)
		}

		result, err := tx.Run(query, queryMap)
		if err != nil {
			fmt.Println("Error on transaction run")
			return nil, err
		}
		return mapResultToFilesFn(result)
	})

	if err != nil {
		fmt.Println("Transaction failure")
		return nil, err
	}

	return result.(*[]File), nil
}

func GetFolder(folderID int) (*Folder, error) {
	session := getDriver().NewSession(neo4j.SessionConfig{})
	defer session.Close()

	fmt.Println("GetFolder: Gettign folder in %d", folderID)

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToFolderFn := getFolderByID(folderID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			fmt.Println("Error on transaction run get folder")
			return nil, err
		}
		return mapResultToFolderFn(result)
	})

	if err != nil {
		fmt.Println("Transaction failure")
		return nil, err
	}

	fmt.Println("Result from get folder")
	fmt.Println("%v", result)
	return result.(*Folder), nil
}

func ExistsFolder(folderID int) (*bool, error) {
	session := getDriver().NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToExistFn := existsFolderByID(folderID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			fmt.Println("Error on transaction run exists folder")
			return nil, err
		}
		return mapResultToExistFn(result)
	})

	if err != nil {
		fmt.Println("Transaction failure")
		return nil, err
	}

	return result.(*bool), nil
}

func GetFoldersIn(folderID *int) (*[]Folder, error) {
	session := getDriver().NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var query string
		var queryMap map[string]interface{}
		var mapResultToFoldersFn func(neo4j.Result) (*[]Folder, error)
		if folderID == nil {
			query, queryMap, mapResultToFoldersFn = getRootFoldersQuery()
		} else {
			fmt.Println("Running methid: getFoldersInFolderQuery")
			query, queryMap, mapResultToFoldersFn = getFoldersInFolderQuery(*folderID)
		}

		fmt.Println("About to run transaction")
		result, err := tx.Run(query, queryMap)
		if err != nil {
			fmt.Println("Error on transaction run")
			return nil, err
		}
		fmt.Println("Folder Result from neo4j:")
		fmt.Println(result)
		return mapResultToFoldersFn(result)
	})

	fmt.Println("Transaction went well")

	if err != nil {
		fmt.Println("Transaction failure")
		return nil, err
	}

	fmt.Println("Here is the result of GetFoldersIn %+v:")
	fmt.Println(result)
	return result.(*[]Folder), nil
}

func CreateFile(fileName string, filePath string, folderParentID *int) (*int, error) {
	var query string
	var queryMap map[string]interface{}
	if folderParentID == nil {
		query, queryMap = createNewFileWithoutParentQuery(fileName, filePath)
	} else {
		query, queryMap = createNewFileWithParentQuery(fileName, filePath, *folderParentID)
	}
	return executeCreateQuery(getDriver())(query, queryMap)
}

func CreateFolder(folderName string, folderParentID *int) (*int, error) {
	var query string
	var queryMap map[string]interface{}
	if folderParentID == nil {
		fmt.Println("createNewFolderWithoutParentQuery")
		query, queryMap = createNewFolderWithoutParentQuery(folderName)
	} else {
		fmt.Println("createNewFolderWithParentQuery", *folderParentID)
		query, queryMap = createNewFolderWithParentQuery(folderName, *folderParentID)
	}
	return executeCreateQuery(getDriver())(query, queryMap)
}

func executeCreateQuery(driver neo4j.Driver) func(string, map[string]interface{}) (*int, error) {
	return func(query string, queryMap map[string]interface{}) (*int, error) {
		// Sessions are short-lived, cheap to create and NOT thread safe. Typically create one or more sessions
		// per request in your web application. Make sure to call Close on the session when done.
		// For multi-database support, set sessionConfig.DatabaseName to requested database
		// Session config will default to write mode, if only reads are to be used configure session for
		// read mode.
		fmt.Println("About to create the session")
		session := getDriver().NewSession(neo4j.SessionConfig{})
		defer session.Close()

		result, err := session.WriteTransaction(createItem(query, queryMap))
		if err != nil {
			fmt.Println("Transaction failure")
			return nil, err
		}

		item := result.(*createdItem)
		return &item.Id, nil
	}
}

func createItem(query string, queryMap map[string]interface{}) func(tx neo4j.Transaction) (interface{}, error) {
	return func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(query, queryMap)
		// In face of driver native errors, make sure to return them directly.
		// Depending on the error, the driver may try to execute the function again.
		if err != nil {
			fmt.Println("Create: error on transaction run")
			return nil, err
		}

		record, err := result.Single()
		if err != nil {
			fmt.Println("error on single")
			return nil, err
		}

		// You can also retrieve values by name, with e.g. `id, found := record.Get("n.id")`
		return &createdItem{
			Id: int(record.Values[0].(int64)),
		}, nil
	}
}

func updateItem(query string, queryMap map[string]interface{}) func(tx neo4j.Transaction) (interface{}, error) {
	return func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(query, queryMap)
		fmt.Println("result:")
		fmt.Println(result)
		// In face of driver native errors, make sure to return them directly.
		// Depending on the error, the driver may try to execute the function again.
		if err != nil {
			fmt.Println("Update: error on transaction run")
			return nil, err
		}

		return nil, nil
	}
}

func UpdateFolder(folderID int, folderName string) error {
	query, queryMap := updateFolderQuery(folderID, folderName)
	return executeUpdateQuery(getDriver())(query, queryMap)
}

func UpdateFile(fileID int, fileName string) error {
	query, queryMap := updateFileQuery(fileID, fileName)
	return executeUpdateQuery(getDriver())(query, queryMap)
}

func MoveFolder(folderID int, destFolderID *int) error {
	var query string
	var queryMap map[string]interface{}
	if destFolderID == nil {
		query, queryMap = moveFolderToRootFolderQuery(folderID)
	} else {
		fmt.Println("Building quieries to move folder")
		query, queryMap = moveFolderQuery(folderID, *destFolderID)
	}
	return executeUpdateQuery(getDriver())(query, queryMap)
}

func MoveFile(fileID int, destFolderID *int) error {
	var query string
	var queryMap map[string]interface{}
	if destFolderID == nil {
		query, queryMap = moveFileToRootFolderQuery(fileID)
	} else {
		query, queryMap = moveFileQuery(fileID, *destFolderID)
	}
	return executeUpdateQuery(getDriver())(query, queryMap)
}

func DeleteFolderContent(folderID int) error {
	query, queryMap := deleteFolderContentQuery(folderID)
	fmt.Println("just build deletd folder content query")
	return executeUpdateQuery(getDriver())(query, queryMap)
}

func DeleteFile(folderID int) error {
	query, queryMap := deleteFileQuery(folderID)
	return executeUpdateQuery(getDriver())(query, queryMap)
}

func executeUpdateQuery(driver neo4j.Driver) func(string, map[string]interface{}) error {
	return func(query string, queryMap map[string]interface{}) error {
		// Sessions are short-lived, cheap to create and NOT thread safe. Typically create one or more sessions
		// per request in your web application. Make sure to call Close on the session when done.
		// For multi-database support, set sessionConfig.DatabaseName to requested database
		// Session config will default to write mode, if only reads are to be used configure session for
		// read mode.
		session := getDriver().NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(updateItem(query, queryMap))
		return err
	}
}
