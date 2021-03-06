package fsmanager

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// TODO gracefully shutdown the go application:
// https://medium.com/@BastianRob/gracefully-shutdown-your-go-application-2ef2871025f0
// defer driver.Close()

// TODO beware database injection (string params)

type File struct {
	Id   int
	Name string
	Path string
}

type Folder struct {
	Id   int
	Name string
}

// Item can be a folder or a file
type CreatedItem struct {
	Id int64
}

const (
	uri      = "neo4j://localhost:7687" // TODO
	username = "neo4j"                  // TODO not commit
	password = "password"               // TODO not commit

	dbId   = "id"
	dbName = "name"
	dbPath = "path"
)

func mapRecordToFile(record *neo4j.Record) (*File, error) {
	id, _ := record.Get(dbId)
	name, _ := record.Get(dbName)
	path, _ := record.Get(dbPath)

	return &File{
		Id:   id.(int),
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
	id, _ := record.Get(dbId)
	name, _ := record.Get(dbName)

	return &Folder{
		Id:   id.(int),
		Name: name.(string),
	}, nil
}

func mapResultToFolders(result neo4j.Result) (*[]Folder, error) {
	var folders []Folder
	for result.Next() == true {
		record := result.Record()

		folder, err := mapRecordToFolder(record)
		if err != nil {
			return nil, err
		}

		folders = append(folders, *folder)
	}
	return &folders, nil
}

// TODO
func getFileByID(fileID int) (string, map[string]interface{}, func(result neo4j.Result) (*File, error)) {
	return `MATCH (file:File)
		WHERE file.id = $fileID
		RETURN file`,
		map[string]interface{}{
			"folderID": fileID,
		},
		func(result neo4j.Result) (*File, error) {
			record, err := result.Single()
			if err != nil {
				return nil, err
			}

			return mapRecordToFile(record)
		}
}

// TODO
func getRootFilesQuery() (string, map[string]interface{}, func(result neo4j.Result) (*[]File, error)) {
	return `MATCH (file:File)
		WHERE NOT (file)-[:IS_INSIDE]->(:Folder)
		RETURN file`, nil, mapResultToFiles

}

// TODO
func getRootFoldersQuery() (string, map[string]interface{}, func(result neo4j.Result) (*[]Folder, error)) {
	return `MATCH (folder:Folder)
		WHERE NOT (folder)-[:IS_INSIDE]->(:Folder)
		RETURN folder`, nil, mapResultToFolders
}

// TODO
func getFilesInFolderQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*[]File, error)) {
	return `MATCH (parentFolder:Folder)
	WHERE parentFolder.id = $folderID
	MATCH (folder:Folder)-[:IS_INSIDE]->(parentFolder)
	RETURN folder`,
		map[string]interface{}{
			"folderID": folderID,
		}, mapResultToFiles
}

// TODO
func getFoldersInFolderQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*[]Folder, error)) {
	return `MATCH (parentFolder:Folder)
	WHERE parentFolder.id = $folderID
	MATCH (file:Folder)-[:IS_INSIDE]->(parentFolder)
	RETURN file`,
		map[string]interface{}{
			"folderID": folderID,
		},
		mapResultToFolders
}

func createNewFileWithParentQuery(fileName string, filePath string, parentFolderID int) (string, map[string]interface{}) {
	return `MATCH (parentFolder:Folder)
	WHERE parentFolder.id = $parentFolderID
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
	return `MATCH (parentFolder:Folder)"
	WHERE parentFolder.id = $parentFolderID"
	MATCH (seq:Sequence {key:'folder_id_sequence'})"
	CALL apoc.atomic.add(seq, 'value', 1, 5)"
	YIELD newValue as folder_id"
	CREATE (folder:Folder { id: folder_id, name: $folderName})"
	CREATE (folder)-[:IS_INSIDE]->(parentFolder)"
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

// InitDriver returns a valid driver
// handles driver lifetime based on your application lifetime requirements  driver's lifetime is usually
// bound by the application lifetime, which usually implies one driver instance per application
func InitDriver() neo4j.Driver {
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

func GetFile(driver neo4j.Driver, fileID int) (*File, error) {
	session := driver.NewSession(neo4j.SessionConfig{})
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

func GetFilesIn(driver neo4j.Driver, folderID *int) (*[]File, error) {
	session := driver.NewSession(neo4j.SessionConfig{})
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

func GetFoldersIn(driver neo4j.Driver, folderID *int) (*[]Folder, error) {
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var query string
		var queryMap map[string]interface{}
		var mapResultToFoldersFn func(neo4j.Result) (*[]Folder, error)
		if folderID == nil {
			query, queryMap, mapResultToFoldersFn = getRootFoldersQuery()
		} else {
			query, queryMap, mapResultToFoldersFn = getFoldersInFolderQuery(*folderID)
		}

		result, err := tx.Run(query, queryMap)
		if err != nil {
			fmt.Println("Error on transaction run")
			return nil, err
		}
		return mapResultToFoldersFn(result)
	})

	if err != nil {
		fmt.Println("Transaction failure")
		return nil, err
	}

	return result.(*[]Folder), nil
}

func CreateFile(driver neo4j.Driver, fileName string, filePath string, folderParentID *int) (*CreatedItem, error) {
	var query string
	var queryMap map[string]interface{}
	if folderParentID == nil {
		query, queryMap = createNewFileWithoutParentQuery(fileName, filePath)
	} else {
		query, queryMap = createNewFileWithParentQuery(fileName, filePath, *folderParentID)
	}
	return executeWriteQuery(driver)(query, queryMap)
}

func CreateFolder(driver neo4j.Driver, folderName string, folderParentID *int) (*CreatedItem, error) {
	var query string
	var queryMap map[string]interface{}
	if folderParentID == nil {
		query, queryMap = createNewFolderWithoutParentQuery(folderName)
	} else {
		query, queryMap = createNewFolderWithParentQuery(folderName, *folderParentID)
	}
	return executeWriteQuery(driver)(query, queryMap)
}

func executeWriteQuery(driver neo4j.Driver) func(string, map[string]interface{}) (*CreatedItem, error) {
	return func(query string, queryMap map[string]interface{}) (*CreatedItem, error) {
		// Sessions are short-lived, cheap to create and NOT thread safe. Typically create one or more sessions
		// per request in your web application. Make sure to call Close on the session when done.
		// For multi-database support, set sessionConfig.DatabaseName to requested database
		// Session config will default to write mode, if only reads are to be used configure session for
		// read mode.
		fmt.Println("About to create the session")
		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		result, err := session.WriteTransaction(createItem(query, queryMap))
		if err != nil {
			fmt.Println("Transaction failure")
			return nil, err
		}

		return result.(*CreatedItem), nil
	}
}

func createItem(query string, queryMap map[string]interface{}) func(tx neo4j.Transaction) (interface{}, error) {
	return func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(query, queryMap)
		// In face of driver native errors, make sure to return them directly.
		// Depending on the error, the driver may try to execute the function again.
		if err != nil {
			fmt.Println("error on transaction run")
			return nil, err
		}

		record, err := result.Single()
		if err != nil {
			fmt.Println("error on single")
			return nil, err
		}

		// You can also retrieve values by name, with e.g. `id, found := record.Get("n.id")`
		return &CreatedItem{
			Id: record.Values[0].(int64),
		}, nil
	}
}
