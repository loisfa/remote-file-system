neo4j_home => /var/lib/neo4j/ + /usr/share/neo4j/
neo4j_config => /etc/neo4j

apoc install:
https://community.neo4j.com/t/how-can-i-install-apoc-library-for-neo4j-version-3-4-6-edition-community/1495/4
https://stackoverflow.com/questions/42286508/apoc-is-only-partially-installing-its-extension-in-neo4j-one-procedure/42357481#42357481

CREATE = 1 item
UPDATE (name) = 1 item
GET = list of items
DELETE (hard delete) = a recursive delete items in a branch tree 

1/INIT
-file/folder ID uniqueness:

# uniqueness constraint on file ids
CREATE CONSTRAINT unique_file_id
ON (file:File)
ASSERT file.id IS UNIQUE

# uniqueness constraint on folder ids
CREATE CONSTRAINT unique_folder_id
ON (folder:Folder)
ASSERT folder.id IS UNIQUE

# index on file id 
CREATE INDEX FOR (file:File) ON (file.id)
=> NOT NEEDED: There is a uniqueness constraint on (:File {id}), so an index is already created that matches this.

# index on folder id
CREATE INDEX FOR (folder:Folder) ON (folder.id)
=> There is a uniqueness constraint on (:Folder {id}), so an index is already created that matches this.

# sequence for the file ids 
CREATE (s:Sequence {key:"file_id_sequence", value: 0})

# sequence for the folder ids 
CREATE (s:Sequence {key:"folder_id_sequence", value: 0})


2/RUNTIME
TODO: have to create the root folders: not necessarily since by definition, a root folder is a folder which is not inside any folder

# Get the root items (can be multiple of them + can be files or folders)
MATCH (item)
WHERE NOT (folder)-[:IS_INSIDE]->(:Folder) AND (item:File OR item:Folder)
RETURN item

# Get the content (=files and folders) of a folder
# input={folderId?}
MATCH (parentFolder:Folder)
WHERE parentFolder.id = ${parentFolderId}
MATCH (item)-[:IS_INSIDE]->(parentFolder) AND (item:File or item.Folder)
RETURN item

# create new file
# input={fileName, filePath, parentFolderId}
MATCH (parentFolder:Folder)
WHERE parentFolder.id = ${parentFolderId}
MATCH (seq:Sequence {key:"file_id_sequence"})
CALL apoc.atomic.add(seq, 'value', 1, 5)
YIELD newValue as file_id
CREATE (file:File { id: file_id, name: '${fileName}', path: ${filePath} })
CREATE (file)-[:IS_INSIDE]->(parentFolder)
RETURN file.id AS fileId

# without parentFolderId
MATCH (seq:Sequence {key:"file_id_sequence"})
CALL apoc.atomic.add(seq, 'value', 1, 5)
YIELD newValue as file_id
CREATE (file:File { id: file_id, name: '${fileName}', path: '${filePath}' })
RETURN file.id AS fileId


# create new folder
# input={folderName, parentFolderId}
MATCH (parentFolder:Folder)
WHERE parentFolder.id = ${parentFolderId}
MATCH (seq:Sequence {key:"folder_id_sequence"})
CALL apoc.atomic.add(seq, 'value', 1, 5)
YIELD newValue as folder_id
CREATE (folder:Folder { id: folder_id, name: '${folderName}' })
CREATE (folder)-[:IS_INSIDE]->(parentFolder)
RETURN folder.id AS folderId

# without parentFolderId
MATCH (seq:Sequence {key:"folder_id_sequence"})
CALL apoc.atomic.add(seq, 'value', 1, 5)
YIELD newValue as folder_id
CREATE (folder:Folder { id: folder_id, name: '${folderName}' })
RETURN folder.id AS folderId

