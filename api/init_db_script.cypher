// Scripts to be executed to initialize the database
// From a bash run:
// `cypher-shell -u <USER> -p <PASSWORD> < <THIS_SCRIPT>`

// Uniqueness constraint on file ids
CREATE CONSTRAINT unique_file_id
ON (file:File)
ASSERT file.id IS UNIQUE

// Uniqueness constraint on folder ids
CREATE CONSTRAINT unique_folder_id
ON (folder:Folder)
ASSERT folder.id IS UNIQUE

// Sequence for the file ids 
CREATE (s:Sequence {key:"file_id_sequence", value: 0})

// Sequence for the folder ids 
CREATE (s:Sequence {key:"folder_id_sequence", value: 0})
