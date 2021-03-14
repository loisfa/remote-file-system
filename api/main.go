package main

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

// TODO: do not expose the database errors, to be rewritten with message

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/health-check", healthCheckStatusOK).Methods(http.MethodGet)

	/*
	 * FOLDERS
	 */

	r.HandleFunc("/folders/{folderId:[0-9]+}", getFolderContent).Methods(http.MethodGet)
	r.HandleFunc("/folders", getRootFolderContent).Methods(http.MethodGet)

	r.HandleFunc("/folders", createFolder).Methods(http.MethodPost)

	r.HandleFunc("/folders/{folderId:[0-9]+}", updateFolder).Methods(http.MethodPut)

	r.HandleFunc("/folders/{folderId:[0-9]+}", deleteFolderAndContent).Methods(http.MethodDelete, http.MethodOptions) // SEE if can delete methodOptions

	r.HandleFunc("/MoveFolder/{folderId:[0-9]+}", moveFolder).Queries("dest", "{destFolderId:[0-9]+}").Methods(http.MethodPut)

	// TODO: download selected folder as a .zip

	/*
	 * FILES
	 */
	r.HandleFunc("/files/{fileId:[0-9]+}", deleteFile).Methods(http.MethodDelete, http.MethodOptions)

	r.HandleFunc("/DownloadFile/{fileId:[0-9]+}", serveFile).Methods(http.MethodGet)

	r.HandleFunc("/UploadFile", uploadFile).Queries("dest", "{destFolderId:[0-9]+}").Methods(http.MethodPost)

	r.HandleFunc("/MoveFile/{fileId:[0-9]+}", moveFile).Queries("dest", "{destFolderId:[0-9]+}").Methods(http.MethodPut)

	http.Handle("/", r)

	corsMw := mux.CORSMethodMiddleware(r)
	r.Use(corsMw)

	// TODO: see if can be deleted (in favor of what is just above)
	corsObj := handlers.AllowedOrigins([]string{"*"})
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization",
		"Accept", "Accept-Language", "Content-Language", "Origin"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	http.ListenAndServe(":8080", handlers.CORS(corsObj, headersOk, methodsOk)(r))

	fmt.Println("Server running on port 8080")
}

type ApiFolder struct {
	Id       int    `json:"id"` // readonly
	Name     string `json:"name"`
	ParentId *int   `json:"parentId"` // nil in case folder is at the root
}

type ApiFile struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ApiFolderContent struct {
	CurrentFolder ApiFolder   `json:"currentFolder"` // nil in case of root folder
	Folders       []ApiFolder `json:"folders"`       // readonly
	Files         []ApiFile   `json:"files"`         // readonly
}

func getContentIn(folderId int) (*ApiFolderContent, error) {
	currentFolder, err := fsmanager.GetFolder(folderId)
	if err != nil {
		return nil, err
	}

	subFolders, err := fsmanager.GetFoldersIn(folderId)
	if err != nil {
		return nil, err
	}

	files, err := fsmanager.GetFilesIn(folderId)
	if err != nil {
		return nil, err
	}

	apiCurrentFolder := ApiFolder{
		(*currentFolder).Id,
		(*currentFolder).Name,
		(*currentFolder).ParentId}

	apiFolders := make([]ApiFolder, 0)
	for idx := range *subFolders {
		folder := (*subFolders)[idx]
		apiFolders = append(apiFolders, mapFolderToApiFolder(folder))
	}

	apiFiles := make([]ApiFile, 0)
	for idx := range *files {
		file := (*files)[idx]
		apiFiles = append(apiFiles, mapFileToApiFile(file))
	}

	return &ApiFolderContent{apiCurrentFolder, apiFolders, apiFiles}, nil
}

func mapFolderToApiFolder(folder fsmanager.Folder) ApiFolder {
	return ApiFolder{
		folder.Id,
		folder.Name,
		&folder.Id}
}

func mapFileToApiFile(file fsmanager.File) ApiFile {
	return ApiFile{
		file.Id,
		file.Name}
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var fileId int
	fileIdStr := vars["fileId"]
	var err error
	if fileId, err = strconv.Atoi(fileIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, err := fsmanager.GetFile(fileId)
	if file == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	http.ServeFile(w, r, file.Path)
}

func getFolderContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var folderId int
	idStr := vars["folderId"]
	var err error
	if folderId, err = strconv.Atoi(idStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	found, err := fsmanager.ExistsFolder(folderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if found == nil || *found == false {
		http.Error(w, "Folder not found. Cannot retrieve its content", http.StatusNotFound)
		return
	}

	apiFolderContent, err := getContentIn(folderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if apiFolderContent == nil {
		http.Error(w, "Folder not found. Cannot retrieve its content", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiFolderContent)
}

func getRootFolderContent(w http.ResponseWriter, r *http.Request) {
	id, err := fsmanager.GetRootFolderID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiFolderContent, err := getContentIn(*id)

	if err != nil {
		http.Error(w, "Folder not found. Cannot retrieve its content", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiFolderContent)
}

func createFolder(w http.ResponseWriter, r *http.Request) {
	var folder ApiFolder
	err := json.NewDecoder(r.Body).Decode(&folder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	destFolderId := folder.ParentId
	if destFolderId == nil {
		http.Error(w, "Create folder: missing destination folder id", http.StatusBadRequest)
		return
	}
	if exists, err := fsmanager.ExistsFolder(*destFolderId); err != nil || exists == nil || !*exists {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := fsmanager.CreateFolder(folder.Name, *destFolderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, strconv.Itoa(*id))
}

func updateFolder(w http.ResponseWriter, r *http.Request) {
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

	err = fsmanager.UpdateFolder(*id, f.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteFolderAndContent(w http.ResponseWriter, r *http.Request) {
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

	err = fsmanager.DeleteFolderAndContent(*id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func moveFolder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var folderId int
	var err error
	folderIdStr := vars["folderId"]
	if folderId, err = strconv.Atoi(folderIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var destFolderIdInt int
	destFolderIdStr := vars["destFolderId"]
	if destFolderIdInt, err = strconv.Atoi(destFolderIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = fsmanager.MoveFolder(folderId, destFolderIdInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
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

	err = fsmanager.DeleteFile(*id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func moveFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var fileId int
	var err error
	fileIdStr := vars["fileId"]
	if fileId, err = strconv.Atoi(fileIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var destFolderIdInt int
	destFolderIdStr := vars["destFolderId"]
	if destFolderIdInt, err = strconv.Atoi(destFolderIdStr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = fsmanager.MoveFile(fileId, destFolderIdInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var destFolderId int
	if destFolderIdStr, found := vars["destFolderId"]; found {
		id, err := strconv.Atoi(destFolderIdStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		destFolderId = id
	}

	// 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	exts := strings.Split(handler.Filename, ".")
	ext := exts[len(exts)-1]
	tempFile, err := ioutil.TempFile("tmp-files", fmt.Sprintf("upload-*.%s", ext))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tempFile.Write(fileBytes)

	fileId, err := fsmanager.CreateFile(handler.Filename, tempFile.Name(), destFolderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, strconv.Itoa(*fileId))
}

func healthCheckStatusOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
