package main

// TODO try neo4j to model the graph structure:
// https://stackoverflow.com/questions/31079881/simple-recursive-cypher-query

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/loisfa/remote-file-system/api/fsmanager"
)

// Rename by FolderDTO?? (same for other Api* models) 
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

// TODO why upper case? Is not exposed outside
func GetContentIn(folderId int) (*ApiFolderContent, error) {
	currentFolder, err := fsmanager.DBGetFolder(folderId)
	if err != nil {
		return nil, err
	}
	
	var apiCurrentFolder *ApiFolder = &ApiFolder{
		currentFolder.Id,
		currentFolder.Name,
		currentFolder.ParentId}

	subFolders, err := fsmanager.DBGetFoldersIn(folderId)
	if err != nil {
			return nil, err
		}

	var apiFolders []*ApiFolder
	for _, folder := range subFolders {
		apiFolder := &ApiFolder{
			folder.Id,
			folder.Name,
			&folderId}
		apiFolders = append(apiFolders, apiFolder)
	}
	
	fileIdToFileMap := fsmanager.GetDbFileMap()
	var apiFiles []*ApiFile
	// TODO with map no need to loop
	for _, file := range fileIdToFileMap {
		if file.ParentId == folderId {
			apiFile := &ApiFile{
				file.Id,
				file.Name}
			apiFiles = append(apiFiles, apiFile)
		}
	}

	return &ApiFolderContent{apiCurrentFolder, apiFolders, apiFiles}, nil
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
	fileIdToFileMap := fsmanager.GetDbFileMap()
	file := fileIdToFileMap[fileId]

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

	apiFolderContent, err := GetContentIn(folderId)
	if (err != nil) {
		http.Error(w, err.Error(), 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiFolderContent)
}

func GetRootFolderContent(w http.ResponseWriter, r *http.Request) {
	apiFolderContent, err := GetContentIn(0) // 0 magic number! TODO use the map instead to retrieve root folder

	if (err != nil) {
		http.Error(w, err.Error(), 404)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
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

	id := fsmanager.DBCreateFolder(f.Name, *(f.ParentId))
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

	err = fsmanager.DBUpdateFolder(*id, f.Name)
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

	err = fsmanager.DBDeleteFolderAndContent(*id)
	if err != nil {
		http.Error(w, err.Error(), 400)
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

	err = fsmanager.DBMoveFolder(folderId, destFolderId)
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

	err = fsmanager.DBDeleteFile(*id)
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

	err = fsmanager.DBMoveFile(fileId, destFolderId)
	if err != nil {
		http.Error(w, err.Error(), 400)
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

	fileId := fsmanager.DBCreateFile(handler.Filename, tempFile.Name(), destFolderId)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, strconv.Itoa(fileId))
}

func healthCheckStatusOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	fsmanager.InitDB()

	r := mux.NewRouter()

	r.HandleFunc("/health-check", healthCheckStatusOK).Methods(http.MethodGet)

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
