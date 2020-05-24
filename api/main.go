package main

// TODO try neo4j to model the graph structure:
// https://stackoverflow.com/questions/31079881/simple-recursive-cypher-query

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type DBFolder struct {
	Id       int // readonly
	Name     string
	ParentId *int // root folder <--> ParentId is nil
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

var int0 = 0
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

var foldersAutoIncrementIndex int
var filesAutoIncrementIndex int

func initDB() {
	dbFoldersMap[rootFolder.Id] = rootFolder
	dbFoldersMap[photosFolder.Id] = photosFolder
	dbFoldersMap[summerPhotosFolder.Id] = summerPhotosFolder
	foldersAutoIncrementIndex = 3
	
	dbFilesMap[textFile1.Id] = textFile1
	dbFilesMap[textFile2.Id] = textFile2
	filesAutoIncrementIndex = 2

	return
}

func DBGetContentIn(folderId int) ApiFolderContent {
	currentFolder, _ := dbFoldersMap[folderId]
	// TODO do something with OK (not found error)
	
	var apiCurrentFolder *ApiFolder = &ApiFolder{
		currentFolder.Id,
		currentFolder.Name,
		currentFolder.ParentId}

	var apiFolders []*ApiFolder
	// TODO with map no need to loop
	for _, folder := range dbFoldersMap {
		if folder.ParentId != nil && *(folder.ParentId) == folderId {
			apiFolder := &ApiFolder{
				folder.Id,
				folder.Name,
				&folderId}
			apiFolders = append(apiFolders, apiFolder)
		}
	}

	var apiFiles []*ApiFile
	// TODO with map no need to loop
	for _, file := range dbFilesMap {
		if file.ParentId == folderId {
			apiFile := &ApiFile{
				file.Id,
				file.Name}
			apiFiles = append(apiFiles, apiFile)
		}
	}

	return ApiFolderContent{apiCurrentFolder, apiFolders, apiFiles}
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

func DBMoveFolder(folderId int, parentId *int) error {
	toUpdateFolder, ok := dbFoldersMap[folderId]
	if ok == false {
		return errors.New("Could not find folder for specified id")
	}

	toUpdateFolder.ParentId = parentId
	return nil
}

func DBMoveFile(fileId int, parentId *int) error {
	toUpdateFile, ok := dbFilesMap[fileId]
	if ok == false {
		return errors.New("Could not find file for specified id")
	}

	toUpdateFile.ParentId = *parentId
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
			fmt.Println("Deleting file", file.Path) // TODO: delete the file actually from the file storage? (for now soft delete)
			delete(dbFilesMap, fileId)
		}
	}
}

func DBDeleteFolderAndContent(folderId int) error {
	fmt.Println("deleting db folder (and chidren)", folderId)
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
	fmt.Println("deleting file", fileId)
	_, ok := dbFilesMap[fileId]

	if ok == false {
		return errors.New("Could not find folder for specified id") // find a way to fire 404
	}

	toDeleteFileIds := []int{fileId}
	removeFiles(toDeleteFileIds)

	return nil
}

type ApiFolder struct {
	Id       int    `json:"id"` // readonly
	Name     string `json:"name"`
	ParentId *int   `json:"parentId"`
	// Path string `json:"path"`
	// Folders []*ApiFolder `json:"folders"` // readonly
	// Files   []*ApiFile   `json:"files"`   // readonly
}

type ApiFile struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
}

type ApiFolderContent struct {
	CurrentFolder *ApiFolder `json:"currentFolder"`
	Folders []*ApiFolder `json:"folders"` // readonly
	Files   []*ApiFile   `json:"files"`   // readonly
}

func hasAccess(userId int, fileId string) bool {
	return true // TODO
}

func ServeFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var fileId int
	fileIdStr := vars["fileId"]
	var err error
	if fileId, err = strconv.Atoi(fileIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file := dbFilesMap[fileId]

	if file != nil {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
		http.ServeFile(w, r, file.Path)
	} else {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

func GetFolderContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var folderId int
	idStr := vars["folderId"]
	var err error
	if folderId, err = strconv.Atoi(idStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	apiFolderContent := DBGetContentIn(folderId)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiFolderContent)
}

func GetRootFolderContent(w http.ResponseWriter, r *http.Request) {
	apiFolderContent := DBGetContentIn(rootFolder.Id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiFolderContent)
}

func CreateFolder(w http.ResponseWriter, r *http.Request) {
	var f ApiFolder
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := DBCreateFolder(f.Name, *(f.ParentId))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, strconv.Itoa(id))
}

func UpdateFolder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var id *int
	var idInt int
	var err error
	idStr := vars["folderId"]
	if idInt, err = strconv.Atoi(idStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id = &idInt

	var f ApiFolder
	err = json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = DBUpdateFolder(*id, f.Name)
	if err != nil {
		http.Error(w, err.Error(), 500) // TODO improve
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteFolderAndContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var id *int
	var idInt int
	var err error
	idStr := vars["folderId"]
	if idInt, err = strconv.Atoi(idStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id = &idInt

	err = DBDeleteFolderAndContent(*id)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func MoveFolder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var folderId int
	var err error
	folderIdStr := vars["folderId"]
	if folderId, err = strconv.Atoi(folderIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("folderId", folderId)
	var destFolderId *int
	var destFolderIdInt int
	destFolderIdStr := vars["destFolderId"]
	if destFolderIdInt, err = strconv.Atoi(destFolderIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	destFolderId = &destFolderIdInt
	fmt.Println("destFolderId", destFolderId)

	err = DBMoveFolder(folderId, destFolderId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var id *int
	var idInt int
	var err error
	idStr := vars["fileId"]
	if idInt, err = strconv.Atoi(idStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id = &idInt

	err = DBDeleteFile(*id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func MoveFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var fileId int
	var err error
	fileIdStr := vars["fileId"]
	if fileId, err = strconv.Atoi(fileIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var destFolderId *int
	var destFolderIdInt int
	destFolderIdStr := vars["destFolderId"]
	if destFolderIdInt, err = strconv.Atoi(destFolderIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	destFolderId = &destFolderIdInt

	err = DBMoveFile(fileId, destFolderId)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UploadFile(w http.ResponseWriter, r *http.Request) { // TODO add the id in the response
	fmt.Println("File Upload Endpoint Hit")

	vars := mux.Vars(r)

	var destFolderId int
	destFolderIdStr := vars["destFolderId"]
	var err error
	if destFolderId, err = strconv.Atoi(destFolderIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("upload")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded file: %+v\n", handler.Filename)

	exts := strings.Split(handler.Filename, ".")
	ext := exts[len(exts)-1]
	tempFile, err := ioutil.TempFile("temp-files", fmt.Sprintf("upload-*.%s", ext))
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Could not read file")
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tempFile.Write(fileBytes)

	fileId := DBCreateFile(handler.Filename, tempFile.Name(), destFolderId)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, strconv.Itoa(fileId))
}

func main() {
	initDB()

	r := mux.NewRouter()


	/*
	 * FOLDERS
	 */

	r.HandleFunc("/folders/{folderId:[0-9]+}", GetFolderContent).Methods(http.MethodGet)
	r.HandleFunc("/folders", GetRootFolderContent).Methods(http.MethodGet)
		
	r.HandleFunc("/folders", CreateFolder).Methods(http.MethodPost)
	
	r.HandleFunc("/folders/{folderId:[0-9]+}", UpdateFolder).Methods(http.MethodPut)
	r.HandleFunc("/folders/{folderId:[0-9]+}", DeleteFolderAndContent).Methods(http.MethodDelete, http.MethodOptions) // SEE if can delere methodOptions

	r.HandleFunc("/MoveFolder/{folderId:[0-9]+}", MoveFolder).Queries("dest", "{destFolderId:[0-9]+}").Methods(http.MethodPut)

	// TODO download selected folder as a .zip

	/*
	 * FILES
	 */
	 
	r.HandleFunc("/files/{fileId:[0-9]+}", DeleteFile).Methods(http.MethodDelete, http.MethodOptions)

	r.HandleFunc("/DownloadFile/{fileId:[0-9]+}", ServeFile).Methods(http.MethodGet)
	
	r.HandleFunc("/UploadFile", UploadFile).Queries("dest", "{destFolderId:[0-9]+}").Methods(http.MethodPost)

	r.HandleFunc("/MoveFile/{fileId:[0-9]+}", MoveFile).Queries("dest", "{destFolderId:[0-9]+}").Methods(http.MethodPut)

	http.Handle("/", r)
	
	corsMw := mux.CORSMethodMiddleware(r)
	r.Use(corsMw)

	// TODO: see if can be deleted (in favor of is just above)
	corsObj := handlers.AllowedOrigins([]string{"*"})
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", 
		"Accept", "Accept-Language", "Content-Language", "Origin"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	
	http.ListenAndServe(":8080", handlers.CORS(corsObj, headersOk, methodsOk)(r))

	fmt.Println("Server running on port 8080")
}
