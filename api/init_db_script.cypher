// Scripts to be executed to initialize the database
// From a bash run: `cypher-shell -a <ADDRESS>:<PORT> -u <USER> -p <PASSWORD> -f <SCRIPT_FILE_PATH>`

// Uniqueness constraint on file ids
CREATE CONSTRAINT unique_file_id
ON (file:File)
ASSERT file.id IS UNIQUE;

// Uniqueness constraint on folder ids
CREATE CONSTRAINT unique_folder_id
ON (folder:Folder)
ASSERT folder.id IS UNIQUE;

// Uniqueness of the root folder
CREATE CONSTRAINT constraint_unique_is_root ON (folder:Folder) ASSERT folder.is_root IS UNIQUE;

// Create the root folder, this should maybe be done at runtime
CREATE (f:Folder  {id:0, is_root: true, name: 'Root folder'});

// Sequence for the file ids 
CREATE (s:Sequence {key:"file_id_sequence", value: 0});

// Sequence for the folder ids 
CREATE (s:Sequence {key:"folder_id_sequence", value: 0});

// TODO: add constraisnt so that only one 'IS_INSIDE' relationship between two nodes
// Issue => not doable with the non-enterprise edition:
// https://neo4j.com/docs/cypher-manual/current/administration/constraints/#administration-constraints-introduction 
