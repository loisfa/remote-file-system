import requests
import json

print("Starting integration tests...")

# prerequisite = ensure that the server is started (use test containers?)
# TODO define a health endpoint on the API?

session = requests.Session()
response = session.get('https://api.agify.io/?name=joe')
body = json.loads(response.text) 
print(body['age'])

# scenario
# create a folder at the root (should we first have the id of the root? => needs a get?
# asserts that the folder was created