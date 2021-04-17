package fsrepository

// TODO use env vars to configure username/password/host/port

// TODO gracefully shutdown the go application:
// https://medium.com/@BastianRob/gracefully-shutdown-your-go-application-2ef2871025f0
// defer driver.Close()

import (
	"errors"
	"fmt"
	"os"

	"github.com/loisfa/remote-file-system/api/fsmodel"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
)

// Item can be a folder or a file
type createdItem struct {
	Id int
}

const (
	// ment to be exposed? if not => TODO remove upper case,
	NEO4J_HOST     = "NEO4J_HOST"
	NEO4J_PORT     = "NEO4J_PORT"
	NEO4J_USER     = "NEO4J_USER"
	NEO4J_PASSWORD = "NEO4J_PASSWORD"

	defaultHost     = "localhost"
	defaultPort     = "7687"
	defaultUser     = "neo4j"
	defaultPassword = "password"

	dbId     = "id"
	dbName   = "name"
	dbPath   = "path"
	dbFolder = "folder"
	dbFile   = "file"
	dbExists = "exists"
)

type IFileSystemRepository interface {
	UpdateFolder(folderID int, folderName string) error
	MoveFolder(folderID int, destFolderID int) error
	MoveFile(fileID int, destFolderID int) error
	DeleteFolderAndContent(folderID int) error
	DeleteFile(folderID int) error
	GetFile(fileID int) (*fsmodel.File, error)
	ExistsFile(fileID int) (*bool, error)
	GetFilesIn(folderID int) (*[]fsmodel.File, error)
	GetFolder(folderID int) (*fsmodel.Folder, error)
	GetRootFolderID() (*int, error)
	IsRootFolder(folderID int) (*bool, error)
	ExistsFolder(folderID int) (*bool, error)
	GetFoldersIn(folderID int) (*[]fsmodel.Folder, error)
	CreateFile(fileName string, filePath string, folderParentID int) (*int, error)
	CreateFolder(folderName string, folderParentID int) (*int, error)
}

type Neo4JFileSystemRepository struct {
	driver neo4j.Driver
}

func NewNeo4JFileSystemRepository() Neo4JFileSystemRepository {
	return Neo4JFileSystemRepository{
		driver: initDriver(),
	}
}

func (repo Neo4JFileSystemRepository) UpdateFolder(folderID int, folderName string) error {
	query, queryMap := updateFolderQuery(folderID, folderName)
	return executeUpdateQuery(repo.driver)(query, queryMap)
}

func (repo Neo4JFileSystemRepository) MoveFolder(folderID int, destFolderID int) error {
	query, queryMap := moveFolderQuery(folderID, destFolderID)
	return executeUpdateQuery(repo.driver)(query, queryMap)
}

func (repo Neo4JFileSystemRepository) MoveFile(fileID int, destFolderID int) error {
	query, queryMap := moveFileQuery(fileID, destFolderID)
	return executeUpdateQuery(repo.driver)(query, queryMap)
}

func (repo Neo4JFileSystemRepository) DeleteFolderAndContent(folderID int) error {
	query, queryMap := deleteFolderAndContentQuery(folderID)
	return executeUpdateQuery(repo.driver)(query, queryMap)
}

func (repo Neo4JFileSystemRepository) DeleteFile(folderID int) error {
	query, queryMap := deleteFileQuery(folderID)
	return executeUpdateQuery(repo.driver)(query, queryMap)
}

func (repo Neo4JFileSystemRepository) GetFile(fileID int) (*fsmodel.File, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToFileFn := getFileByIDQuery(fileID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			return nil, err
		}
		return mapResultToFileFn(result)
	})

	if err != nil {
		return nil, err
	}

	return result.(*fsmodel.File), nil
}

func (repo Neo4JFileSystemRepository) ExistsFile(fileID int) (*bool, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToExistFn := existsFileByIDQuery(fileID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			return nil, err
		}
		return mapResultToExistFn(result)
	})

	if err != nil {
		return nil, err
	}

	return result.(*bool), nil
}

func (repo Neo4JFileSystemRepository) GetFilesIn(folderID int) (*[]fsmodel.File, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToFilesFn := getFilesInFolderQuery(folderID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			return nil, err
		}
		return mapResultToFilesFn(result)
	})

	if err != nil {
		return nil, err
	}

	return result.(*[]fsmodel.File), nil
}

func (repo Neo4JFileSystemRepository) GetFolder(folderID int) (*fsmodel.Folder, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToFolderFn := getFolderByIDQuery(folderID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			return nil, err
		}
		return mapResultToFolderFn(result)
	})

	if err != nil {
		return nil, err
	}

	return result.(*fsmodel.Folder), nil
}

func (repo Neo4JFileSystemRepository) GetRootFolderID() (*int, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToFolderIDFn := getRootFolderIDQuery()
		result, err := tx.Run(query, queryMap)
		if err != nil {
			return nil, err
		}
		return mapResultToFolderIDFn(result)
	})

	if err != nil {
		return nil, err
	}

	return result.(*int), nil
}

func (repo Neo4JFileSystemRepository) IsRootFolder(folderID int) (*bool, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToIsRootFolderFn := isRootFolderQuery(folderID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			return nil, err
		}
		return mapResultToIsRootFolderFn(result)
	})

	if err != nil {
		return nil, err
	}

	return result.(*bool), nil
}

func (repo Neo4JFileSystemRepository) ExistsFolder(folderID int) (*bool, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToExistFn := existsFolderByIDQuery(folderID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			return nil, err
		}
		return mapResultToExistFn(result)
	})

	if err != nil {
		return nil, err
	}

	return result.(*bool), nil
}

func (repo Neo4JFileSystemRepository) GetFoldersIn(folderID int) (*[]fsmodel.Folder, error) {
	session := repo.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query, queryMap, mapResultToFoldersFn := getFoldersInFolderQuery(folderID)
		result, err := tx.Run(query, queryMap)
		if err != nil {
			return nil, err
		}
		return mapResultToFoldersFn(result)
	})

	if err != nil {
		return nil, err
	}

	return result.(*[]fsmodel.Folder), nil
}

func (repo Neo4JFileSystemRepository) CreateFile(fileName string, filePath string, folderParentID int) (*int, error) {
	query, queryMap := createNewFileWithParentQuery(fileName, filePath, folderParentID)
	return executeCreateQuery(repo.driver)(query, queryMap)
}

func (repo Neo4JFileSystemRepository) CreateFolder(folderName string, folderParentID int) (*int, error) {
	query, queryMap := createNewFolderWithParentQuery(folderName, folderParentID)
	return executeCreateQuery(repo.driver)(query, queryMap)
}

// InitDriver returns a valid driver
// handles driver lifetime based on your application lifetime requirements  driver's lifetime is usually
// bound by the application lifetime, which usually implies one driver instance per application
func initDriver() neo4j.Driver {
	host := os.Getenv(NEO4J_HOST)
	if len(host) == 0 {
		fmt.Printf("Could not find envirnment variable %s. Fallback to default '%s'\n", NEO4J_HOST, defaultHost)
		host = defaultHost
	}

	port := os.Getenv(NEO4J_PORT)
	if len(port) == 0 {
		fmt.Printf("Could not find envirnment variable %s. Fallback to default '%s'\n", NEO4J_PORT, defaultPort)
		port = defaultPort
	}

	username := os.Getenv(NEO4J_USER)
	if len(username) == 0 {
		fmt.Printf("Could not find envirnment variable %s. Fallback to default '%s'\n", NEO4J_USER, defaultUser)
		username = defaultUser
	}

	password := os.Getenv(NEO4J_PASSWORD)
	if len(password) == 0 {
		fmt.Printf("Could not find envirnment variable %s. Fallback to default ******\n", NEO4J_PASSWORD)
		password = defaultPassword
	}

	uri := fmt.Sprintf("neo4j://%s:%s", host, port)

	// Neo4j 4.0, defaults to no TLS therefore use bolt:// or neo4j://
	// Neo4j 3.5, defaults to self-signed certificates, TLS on, therefore use bolt+ssc:// or neo4j+ssc://
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		panic(err)
	}

	err = driver.VerifyConnectivity()
	if err != nil {
		panic(err)
	}

	return driver
}

func executeCreateQuery(driver neo4j.Driver) func(string, map[string]interface{}) (*int, error) {
	return func(query string, queryMap map[string]interface{}) (*int, error) {
		// Sessions are short-lived, cheap to create and NOT thread safe. Typically create one or more sessions
		// per request in your web application. Make sure to call Close on the session when done.
		// For multi-database support, set sessionConfig.DatabaseName to requested database
		// Session config will default to write mode, if only reads are to be used configure session for
		// read mode.
		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		result, err := session.WriteTransaction(createItem(query, queryMap))
		if err != nil {
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
			return nil, err
		}

		record, err := result.Single()
		if err != nil {
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
		_, err := tx.Run(query, queryMap)
		// In face of driver native errors, make sure to return them directly.
		// Depending on the error, the driver may try to execute the function again.
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func executeUpdateQuery(driver neo4j.Driver) func(string, map[string]interface{}) error {
	return func(query string, queryMap map[string]interface{}) error {
		// Sessions are short-lived, cheap to create and NOT thread safe. Typically create one or more sessions
		// per request in your web application. Make sure to call Close on the session when done.
		// For multi-database support, set sessionConfig.DatabaseName to requested database
		// Session config will default to write mode, if only reads are to be used configure session for
		// read mode.
		session := driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		_, err := session.WriteTransaction(updateItem(query, queryMap))
		return err
	}
}

func getFileByIDQuery(fileID int) (string, map[string]interface{}, func(result neo4j.Result) (*fsmodel.File, error)) {
	return `MATCH (file:File{id: $fileID})
		RETURN file`,
		map[string]interface{}{
			"fileID": fileID,
		},
		func(result neo4j.Result) (*fsmodel.File, error) {
			record, err := result.Single()
			if err != nil {
				return nil, err
			}

			return mapRecordToFile(record)
		}
}

func existsFileByIDQuery(fileID int) (string, map[string]interface{}, func(result neo4j.Result) (*bool, error)) {
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

func getFolderByIDQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*fsmodel.Folder, error)) {
	return `MATCH (folder:Folder{id: $folderID})
		RETURN folder`,
		map[string]interface{}{
			"folderID": folderID,
		},
		func(result neo4j.Result) (*fsmodel.Folder, error) {
			record, err := result.Single()
			if err != nil {
				return nil, err
			}

			return mapRecordToFolder(record)
		}
}

// createRootFolderIfNotExistsQuery:
// use only if required since it increments the folder_id_sequence even if no root folder to create
func createRootFolderIfNotExistsQuery() (string, map[string]interface{}) {
	return `MATCH (seq:Sequence {key:'folder_id_sequence'})
		CALL apoc.atomic.add(seq, 'value', 1, 5)
		YIELD newValue as folder_id
		
		MERGE (root:Folder {is_root: true})
			ON CREATE
        		SET root.id = folder_id
        		SET root.is_root = true
        		SET root.name = 'Root folder'
		RETURN root.id as folderID`,
		make(map[string]interface{})
}

func getRootFolderIDQuery() (string, map[string]interface{}, func(result neo4j.Result) (*int, error)) {
	return `MATCH (root:Folder {is_root: true})
		RETURN root as folder`,
		make(map[string]interface{}),
		func(result neo4j.Result) (*int, error) {
			record, err := result.Single()
			if err != nil {
				return nil, err
			}
			return mapRecordToFolderID(record)
		}
}

// isRootFolderQuery:
// It assumes you already have checked whether the folder exists or not
func isRootFolderQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*bool, error)) {
	return `MATCH (folder:Folder {id: $folderID})
		RETURN exists(folder.is_root) and folder.is_root = true`,
		map[string]interface{}{
			"folderID": folderID},
		func(result neo4j.Result) (*bool, error) {
			record, err := result.Single()
			if err != nil {
				return nil, err
			}
			isRoot := record.Values[0].(bool)
			return &isRoot, nil
		}
}

func existsFolderByIDQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*bool, error)) {
	return `OPTIONAL MATCH (folder:Folder{id: $folderID})
		RETURN folder IS NOT NULL AS exists`,
		map[string]interface{}{
			"folderID": folderID,
		},
		func(result neo4j.Result) (*bool, error) {
			record, err := result.Single()
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

func getFilesInFolderQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*[]fsmodel.File, error)) {
	return `MATCH (parentFolder:Folder{id: $folderID})
	MATCH (file:File)-[:IS_INSIDE]->(parentFolder)
	RETURN file`,
		map[string]interface{}{
			"folderID": folderID,
		}, mapResultToFiles
}

func getFoldersInFolderQuery(folderID int) (string, map[string]interface{}, func(result neo4j.Result) (*[]fsmodel.Folder, error)) {
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

func updateFolderQuery(folderID int, folderName string) (string, map[string]interface{}) {
	return `MATCH (folder:Folder {id: $folderID})
	SET folder.name = $folderName`,
		map[string]interface{}{
			"folderID":   folderID,
			"folderName": folderName,
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

// TODO: Return the ids of all the deleted items?
func deleteFolderAndContentQuery(folderID int) (string, map[string]interface{}) {
	return `OPTIONAL MATCH (f)-[:IS_INSIDE *1..]->(Folder {id: $folderID})
	OPTIONAL MATCH (folder:Folder {id: $folderID})
	DETACH DELETE f, folder`,
		map[string]interface{}{
			"folderID": folderID,
		}
}

// TODO Return the ids of all the deleted items?
func deleteFileQuery(fileID int) (string, map[string]interface{}) {
	return `MATCH (file:File {id: $fileID})
	DETACH DELETE file`,
		map[string]interface{}{
			"fileID": fileID,
		}
}

func mapRecordToFile(record *neo4j.Record) (*fsmodel.File, error) {
	file, found := record.Get(dbFile)
	if !found {
		return nil, errors.New("Could not find 'file' inside the File record")
	}
	fileProps := (file.(dbtype.Node)).Props

	id, found := fileProps[dbId]
	if !found {
		return nil, errors.New("Could not retrieve 'id' of the file result")
	}
	name, found := fileProps[dbName]
	if !found {
		return nil, errors.New("Could not retrieve 'name' of the file result")
	}
	path, found := fileProps[dbPath]
	if !found {
		return nil, errors.New("Could not retrieve 'name' of the file result")
	}

	return &fsmodel.File{
		Id:   int(id.(int64)),
		Name: name.(string),
		Path: path.(string),
	}, nil
}

func mapResultToFiles(result neo4j.Result) (*[]fsmodel.File, error) {
	var files []fsmodel.File
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

func mapRecordToFolder(record *neo4j.Record) (*fsmodel.Folder, error) {
	folder, found := record.Get(dbFolder)
	if !found {
		return nil, errors.New("Could not find 'file' inside the Folder record")
	}

	folderProps := (folder.(dbtype.Node)).Props
	id, found := folderProps[dbId]
	if !found {
		return nil, errors.New("Could not retrieve 'id' of the Folder record")
	}
	name, found := folderProps[dbName]
	if !found {
		return nil, errors.New("Could not retrieve 'name' of the Folder record")
	}

	return &fsmodel.Folder{
		Id:   int(id.(int64)),
		Name: name.(string),
	}, nil
}

func mapRecordToFolderID(record *neo4j.Record) (*int, error) {
	folder, err := mapRecordToFolder(record)
	if err != nil {
		return nil, err
	}

	return &folder.Id, nil
}

func mapResultToFolders(result neo4j.Result) (*[]fsmodel.Folder, error) {
	var folders []fsmodel.Folder

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
