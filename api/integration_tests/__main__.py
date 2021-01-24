import requests
import json
from model import FolderDTO, FileDTO

# TODO think of using env variables
PORT=8080
ROOT_URL="http://localhost:" + str(PORT)

session = requests.Session()

print("Starting integration tests...")

# Check if the API is running
response = session.get(ROOT_URL + "/health-check")
if (response.status_code != 200):
    print("The API health-check was unsuccessful. Received Http status: " + status_code + ". Please make sure the API.")

# Retrieve the id of the root folder, and check it exists 
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
root_folder_id = body['currentFolder']['id']
assert root_folder_id is not None, "root folder id is null"

# Create a folder inside the root folder
to_create_folder = FolderDTO(None, "Folder 1", root_folder_id)
response = session.post(ROOT_URL + "/folders", to_create_folder.toJson())
assert response.status_code == 201, "Wrong http code received on create new folder in root folder"
created_folder_id = response.text

# Check if the folder was created
response = session.get(ROOT_URL + "/folders")
body = json.loads(response.text)
folders = body['folders']
found_created_folder = False
for folder in folders:
    if str(folder['id']) == str(created_folder_id):
        found_created_folder = True
        assert folder['name'] == "Folder 1", "The name of the folder just created is wrong"
assert found_created_folder == True, "Could not find the folder just created"

print("Integration tests finished")
