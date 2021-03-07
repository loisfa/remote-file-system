import os
import requests
import json
import time
import math
import cgi
import sys
from model.dto import CreateFolderDTO, UpdateFolderDTO

# TODO think of using env variables
PORT=8080
ROOT_URL="http://localhost:" + str(PORT)

session = requests.Session()

print("Starting integration tests...")

# HEALTH CHECK
response = session.get(ROOT_URL + "/health-check")
if (response.status_code != 200):
    print("The API health-check was unsuccessful. Received Http status: " + str(response.status_code) + ". Please make sure the API.")

# GET FOLDER
# Retrieve the id of the root folder, and check it exists 
response = session.get(ROOT_URL + "/folders")
assert response.status_code == 200, "Wrong http code received on retrieve root folder"
body = json.loads(response.text)
root_folder_id = body['currentFolder']['id']
assert root_folder_id is not None, "root folder id is null"
# Ensures GET on non-existing folder returns 404 
response = session.get(ROOT_URL + "/folders/123456")
assert response.status_code, "Wrong http code received on retrieve non-existing folder: " + str(response.status_code)

### CREATE FOLDER
# Create a folder inside the root folder
to_create_folder = CreateFolderDTO("Folder 1", root_folder_id)
response = session.post(ROOT_URL + "/folders", to_create_folder.toJson())
assert response.status_code == 201, "Wrong http code received on create new folder in root folder"
created_folder_id = response.text
# Check whether the folder was created
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
folders = body['folders']
found_created_folder = False
for folder in folders:
    if str(folder['id']) == str(created_folder_id):
        found_created_folder = True
        assert folder['name'] == "Folder 1", "The name of the folder just created is wrong"
assert found_created_folder == True, "Could not find the folder just created"

### UPDATE FOLDER
# Updated root name
to_update_folder = UpdateFolderDTO("Root folder new name", None)
response = session.put(ROOT_URL + "/folders/" + str(root_folder_id), to_update_folder.toJson())
assert response.status_code == 204, "Wrong http code received on folder update: " + str(response.status_code)
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
assert body['currentFolder']['name'] == "Root folder new name", "Wrong name for root folder on update"
# Updated created folder's name
to_update_folder = UpdateFolderDTO("Folder 1.1", None)
response = session.put(ROOT_URL + "/folders/" + created_folder_id, to_update_folder.toJson())
assert response.status_code == 204, "Wrong http code received on folder update: " + str(response.status_code)
# Check whether the folder name was updated
response = session.get(ROOT_URL + "/folders/" + created_folder_id)
body = json.loads(response.text)
assert body['currentFolder']['name'] == "Folder 1.1", "The name of the folder just updated is wrong"

### MOVE FOLDER
# Ensure cannot move root folder
response = session.put(ROOT_URL + "/MoveFolder/" + str(root_folder_id) + "?dest=" + str(created_folder_id))
assert response.status_code == 500, "Wrong http code received on move folder: " + str(response.status_code)
# Create folder 2 in the root
to_create_folder_2 = CreateFolderDTO("Folder 2", root_folder_id)
response = session.post(ROOT_URL + "/folders", to_create_folder_2.toJson())
assert response.status_code == 201, "Wrong http code received on create new folder in root folder"
created_folder_2_id = response.text
# Check the created folder 2 is in the root
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
folders = body['folders']
found_created_folder = False
for folder in folders:
    if str(folder['id']) == str(created_folder_2_id):
        found_created_folder = True
assert found_created_folder == True, "Could not find the updated folder"
# Move the folder 2 into folder 1
response = session.put(ROOT_URL + "/MoveFolder/" + created_folder_2_id + "?dest=" + created_folder_id)
assert response.status_code == 204, "Wrong http code received on move folder: " + str(response.status_code) 
# Check the folder 2 is now inside folder 1
response = session.get(ROOT_URL + "/folders/" + created_folder_id)
body = json.loads(response.text)
folders = body['folders']
found_created_folder = False
assert folders is not None, "Folder 1 does not contain any folder"
for folder in folders:
    if str(folder['id']) == str(created_folder_2_id):
        found_created_folder = True
assert found_created_folder == True, "Could not find the updated folder"

### DELETE A FOLDER
# Ensure cannot delete root folder
response = session.delete(ROOT_URL + "/folders/" + str(root_folder_id))
assert response.status_code == 400, "Wrong http code received on delete root folder: " + str(response.status_code)
# Delete folder 1
response = session.delete(ROOT_URL + "/folders/" + str(created_folder_id))
assert response.status_code == 204, "Wrong http code received on delete folder 1: " + str(response.status_code)
# Ensures folder 1 does not exist anymore
response = session.get(ROOT_URL + "/folders/" + str(created_folder_id))
assert response.status_code == 404, "Wrong http code received on retrieve deleted folder 1: " + str(response.status_code)
# Check the folder 1 is not inside root folder anymore
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
folders = body['folders']
found_created_folder = False
if folders is None:
    assert True
else:
    for folder in folders:
        if str(folder['id']) == str(created_folder_id):
            found_created_folder = True
    assert found_created_folder == False, "Found folder 1 but should have been deleted"
# Ensures folder 2 (subfolder of folder 1) does not exist anymore
response = session.get(ROOT_URL + "/folders/" + str(created_folder_id))
assert response.status_code == 404, "Wrong http code received on retrieve deleted folder 2: " + str(response.status_code)

### Files tests Config
epoch_seconds = math.floor(time.time())
tmp_path = 'tmp-' + str(epoch_seconds)
tmp_files_path = tmp_path + '/files'

try:
    os.makedirs(tmp_files_path)
except OSError as err:
    print("OS error: {0}".format(err))
except:
    print("Unexpected error:", sys.exc_info()[0])
    raise

file1_name = 'temp_file_1.txt'
file1_path = tmp_files_path + '/' + file1_name
file1 = open(file1_path, "w+")
file1.write("Text file 1: this is the text.\n")
file1.close()

### UPLOAD A FILE IN ROOT FOLDER
response = session.post(
    ROOT_URL + "/UploadFile?dest=" + str(root_folder_id), 
    files = { 'upload': open(file1_path, 'rb') })
assert response.status_code == 201, "Wrong http code received on create file in root folder: " + str(response.status_code)
body = json.loads(response.text)
uploaded_file1_id = body
# Ensures the file is part of the root folder content
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
files = body['files']
found_created_file = False
for file in files:
    if str(file['id']) == str(uploaded_file1_id):
        found_created_file = True
        assert file['name'] == file1_name, "The name of the file just uploaded is wrong"
assert found_created_file == True, "Could not find the file just uploaded"
# Download the file and ensure the content corresponds
response = session.get(ROOT_URL + "/DownloadFile/" + str(uploaded_file1_id))
assert response.status_code == 200, "Wrong http code received on download uploaded file: " + str(response.status_code)
headers = response.headers['Content-Disposition']
value, params = cgi.parse_header(headers)
retrieved_filename = params['filename']
assert retrieved_filename == file1_name, "Wrong file name for downloaded file: " + retrieved_filename
assert response.content == open(file1.name, 'rb').read(), "Wrong content for the downloaded file: " + str(response.content) 

### MOVE FILE FROM ROOT FOLDER TO ANOTHER FOLDER
# Create a folder inside the root folder
to_create_folder_3 = CreateFolderDTO("Folder 3", root_folder_id)
response = session.post(ROOT_URL + "/folders", to_create_folder.toJson())
assert response.status_code == 201, "Wrong http code received on create new folder in root folder"
created_folder_3_id = response.text
# Move file from root to folder 3
response = session.put(ROOT_URL + "/MoveFile/" + str(uploaded_file1_id) + "?dest=" + str(created_folder_3_id))
assert response.status_code == 204, "Wrong http code received on move file: " + str (response.status_code)
# Ensure the file is part of the folder content
response = session.get(ROOT_URL + "/folders/" + created_folder_3_id)
body = json.loads(response.text)
files = body['files']
found_created_file = False
for file in files:
    if str(file['id']) == str(uploaded_file1_id):
        found_created_file = True
        assert file['name'] == file1_name, "The name of the file just uploaded is wrong"
assert found_created_file == True, "Could not find the file just uploaded"
# Ensure the file is NOT anymore in the root folder
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
files = body['files']
if files != None:
    found_created_file = False
    for file in files:
        if str(file['id']) == str(uploaded_file1_id):
            found_created_file = True
    assert found_created_file == False, "The moved file is still in the origin folder (=duplicate)"
# Download the file and ensure the content corresponds
response = session.get(ROOT_URL + "/DownloadFile/" + str(uploaded_file1_id))
assert response.status_code == 200, "Wrong http code received on download uploaded file: " + str(response.status_code)
headers = response.headers['Content-Disposition']
value, params = cgi.parse_header(headers)
retrieved_filename = params['filename']
assert retrieved_filename == file1_name, "Wrong file name for downloaded file: " + retrieved_filename
assert response.content == open(file1.name, 'rb').read(), "Wrong content for the downloaded file: " + str(response.content) 

### DELETE THE FILE
response = session.delete(ROOT_URL + "/files/" + str(uploaded_file1_id))
assert response.status_code == 204, "Wrong http code received on delte file: " + str(response.status_code)
# Ensure the file is NOT part of the folder anymore
response = session.get(ROOT_URL + "/folders/" + created_folder_3_id)
body = json.loads(response.text)
files = body['files']
if files != None:
    found_created_file = False
    for file in files:
        if str(file['id']) == str(uploaded_file1_id):
            found_created_file = True
    assert found_created_file == False, "The deleted file still appears inside the folder"
# Ensure the file cannot be lookup
response = session.get(ROOT_URL + "/DownloadFile/" + str(uploaded_file1_id))
assert response.status_code == 404, "Wrong http code received on trying to access deleted file: " + str(response.status_code)

### ENSURE FILES ARE REMOVED WHEN FOLDERS ARE DELETED
# /folder1/file.txt
# /folder1/folder1.2/file.txt
# Create /folder1
to_create_folder1 = CreateFolderDTO("folder1", root_folder_id)
response = session.post(ROOT_URL + "/folders", to_create_folder1.toJson())
assert response.status_code == 201, "Wrong http code received on create /folder1 in root folder: " + str(response.status_code)
created_folder1_id: int = int(response.text)
# Create /folder1/folder1.1
to_create_folder1_2 = CreateFolderDTO("folder1.2", created_folder1_id)
response = session.post(ROOT_URL + "/folders", to_create_folder1_2.toJson())
created_folder2_id: int = int(response.text)
assert response.status_code == 201, "Wrong http code received on create /folder1.2 in /folder1: " + str(response.status_code) 
# Add fileA.txt in /folder1
response = session.post(
    ROOT_URL + "/UploadFile?dest=" + str(created_folder1_id), 
    files = { 'upload': open(file1_path, 'rb') })
assert response.status_code == 201, "Wrong http code received on create file in /folder1: " + str(response.status_code)
body = json.loads(response.text)
uploaded_fileA_id = body
# Add fileB.txt in /folder1/folder1.2
response = session.post(
    ROOT_URL + "/UploadFile?dest=" + str(created_folder2_id), 
    files = { 'upload': open(file1_path, 'rb') })
assert response.status_code == 201, "Wrong http code received on create file in /folder1: " + str(response.status_code)
body = json.loads(response.text)
uploaded_fileB_id = body
# Ensure both files exist
response = session.get(ROOT_URL + "/DownloadFile/" + str(uploaded_fileA_id))
assert response.status_code == 200, "Wrong http code received on download uploaded fileA: " + str(response.status_code)
response = session.get(ROOT_URL + "/DownloadFile/" + str(uploaded_fileB_id))
assert response.status_code == 200, "Wrong http code received on download uploaded fileB: " + str(response.status_code)
# Delete /folder1
response = session.delete(ROOT_URL + "/folders/" + str(created_folder1_id))
assert response.status_code == 204, "Wrong http code received on delete /folder1: " + str(response.status_code)
# Ensure /folder1 and /folder1/folder1.2 are deleted
response = session.get(ROOT_URL + "/folders/" + str(created_folder1_id))
assert response.status_code == 404, "Wrong http code received on get deleted /folder1: " + str(response.status_code)
response = session.get(ROOT_URL + "/folders/" + str(created_folder2_id))
assert response.status_code == 404, "Wrong http code received on get deleted /folder1/folder1.2: " + str(response.status_code)
# Ensure both files are not found anymore
response = session.get(ROOT_URL + "/DownloadFile/" + str(uploaded_fileA_id))
assert response.status_code == 404, "Wrong http code received on download deleted fileA: " + str(response.status_code)
response = session.get(ROOT_URL + "/DownloadFile/" + str(uploaded_fileB_id))
assert response.status_code == 404, "Wrong http code received on download deleted fileB: " + str(response.status_code)

os.remove(file1.name)
os.rmdir(tmp_files_path)
os.rmdir(tmp_path)

print("Integration tests finished successfully.")
