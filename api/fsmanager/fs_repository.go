package fsmanager

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// TODO gracefully shutdown the go application:
// https://medium.com/@BastianRob/gracefully-shutdown-your-go-application-2ef2871025f0
// defer driver.Close()

const (
	uri      = "neo4j://localhost:7687" // TODO
	username = "neo4j"                  // TODO not commit
	password = "password"               // TODO not commit
)

func getRootItemQuery() (string, map[string]interface{}) {
	return `MATCH (item)
		WHERE NOT (folder)-[:IS_INSIDE]->(:Folder)
		AND (item:File OR item:Folder) RETURN item`, nil
}

func getFolderContentQuery(folderID int) (string, map[string]interface{}) {
	return `MATCH (parentFolder:Folder)
	WHERE parentFolder.id = $folderID
	MATCH (item)-[:IS_INSIDE]->(parentFolder) AND (item:File or item.Folder)
	RETURN item`,
		map[string]interface{}{
			"folderID": folderID,
		}
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

// Create new folder
// input={folderName, parentFolderID}
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

func InsertFile(driver neo4j.Driver) (*Item, error) {
	// Sessions are short-lived, cheap to create and NOT thread safe. Typically create one or more sessions
	// per request in your web application. Make sure to call Close on the session when done.
	// For multi-database support, set sessionConfig.DatabaseName to requested database
	// Session config will default to write mode, if only reads are to be used configure session for
	// read mode.
	fmt.Println("About to create the session")
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	query, queryMap := createNewFileWithoutParentQuery("This is my file name", "/my-file-path/here")
	result, err := session.WriteTransaction(createItemFn(query, queryMap))
	if err != nil {
		fmt.Println("Transaction failure")
		return nil, err
	}

	return result.(*Item), nil
}

func createItemFn(query string, queryMap map[string]interface{}) func(tx neo4j.Transaction) (interface{}, error) {
	return func(tx neo4j.Transaction) (interface{}, error) {
		/* records, err := tx.Run("CREATE (n:Item { id: $id, name: $name }) RETURN n.id, n.name", map[string]interface{}{
			"id":   1,
			"name": "Item 1",
		}) */

		records, err := tx.Run(query, queryMap)
		// In face of driver native errors, make sure to return them directly.
		// Depending on the error, the driver may try to execute the function again.
		if err != nil {
			fmt.Println("error on transaction run")
			return nil, err
		}

		record, err := records.Single()
		if err != nil {
			fmt.Println("error on single")
			return nil, err
		}

		// You can also retrieve values by name, with e.g. `id, found := record.Get("n.id")`
		return &Item{
			Id: record.Values[0].(int64),
		}, nil
	}
}

// Item can be a folder or a file
type Item struct {
	Id int64
}
