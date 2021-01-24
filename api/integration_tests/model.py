import json

class FolderDTO():
    def __init__(self, id, name, parent_id):
        self.id = id
        self.name = name
        self.parent_id = parent_id
    
    def toJson(self):
        obj = dict()
        obj['id'] = self.id
        obj['name'] = self.name
        obj['parentId'] = self.parent_id
        return json.dumps(obj)

class FileDTO():
    def __init__(self, id, name):
        self.id = id
        self.name = name

    def toJson(self):
        obj = dict()
        obj['id'] = self.id
        obj['name'] = self.name
        return json.dumps(obj)

class FolderContentDTO():
    def __init__(self, current_folder, folders, files):
        self.current_folder = current_folder
        self.folders = folders
        self.files = files
    
    def toJson(self):
        obj = dict()
        obj['current_folder'] = self.current_folder
        obj['folders'] = self.folders
        obj['files'] = self.files
        return json.dumps(obj)
