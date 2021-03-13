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
assert response.status_code == 200, "Wrong http code received on retrieve root folder: " + str(response.status_code)
body = json.loads(response.text)
root_folder = body['currentFolder']
assert root_folder is None, "root folder should not return any current folder"
# Ensures GET on non-existing folder returns 404 
response = session.get(ROOT_URL + "/folders/123456")
assert response.status_code, "Wrong http code received on retrieve non-existing folder: " + str(response.status_code)

### CREATE FOLDER
# Create a folder inside the root folder
to_create_folder = CreateFolderDTO("Folder 1", None)
response = session.post(ROOT_URL + "/folders", to_create_folder.toJson())
assert response.status_code == 201, "Wrong http code received on create new folder in root folder: " + str(response.status_code)
created_folder1_id = response.text
# Check whether the folder was created
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
folders = body['folders']
found_created_folder = False
for folder in folders:
    if str(folder['id']) == str(created_folder1_id):
        found_created_folder = True
        assert folder['name'] == "Folder 1", "The name of the folder just created is wrong"
assert found_created_folder == True, "Could not find the folder just created"
# Create a folder inside the created folder
to_create_folder = CreateFolderDTO("Folder 2", int(created_folder1_id))
response = session.post(ROOT_URL + "/folders", to_create_folder.toJson())
assert response.status_code == 201, "Wrong http code received on create new folder in root folder: " + str(response.status_code)
created_folder2_id = response.text
# Check whether the folder was created
response = session.get(ROOT_URL + "/folders/" + str(created_folder1_id))
body = json.loads(response.text)
folders = body['folders']
found_created_folder = False
for folder in folders:
    if str(folder['id']) == str(created_folder2_id):
        found_created_folder = True
        assert folder['name'] == "Folder 2", "The name of the folder just created is wrong"
assert found_created_folder == True, "Could not find the folder just created"

### UPDATE FOLDER

### MOVE FOLDER

### DELETE A FOLDER

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

# UPLOAD A FILE IN ROOT FOLDER
response = session.post(
    ROOT_URL + "/UploadFile", 
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

# MOVE FILE FROM ROOT FOLDER TO ANOTHER FOLDER

# DELETE THE FILE

print("Integration tests finished successfully.")
