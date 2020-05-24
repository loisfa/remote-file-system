<script>
		import Folder from "./Folder.svelte";
		import File from "./File.svelte";
		import {
		  apiCreateFolder,
		  apiMoveFolder,
		  apiDeleteFolder,
		  apiUpdateFolder,
		  apiGetFolderContent
		} from "./api/folderApi.js";
		import {
		  apiMoveFile,
		  apiDeleteFile,
		  apiUploadFile,
		  getDownloadUrl
		} from "./api/fileApi.js";

		// the one currently displayed
		const ROOT_FOLDER = {
		  name: "",
		  id: 0,
		  parentId: undefined
		};
		const NEW_FOLDER_DEFAULT_NAME = "New Folder";

		const idOrdering = (a, b) => {
		  if (a.id > b.id) return 1;
		  else if (a.id == b.id) return 0;
		  else return -1;
		};

		let currentFolder = ROOT_FOLDER;
		let folders = [];
		let files = [];
		let isAddingFolder = false;
		let addingFolderName = NEW_FOLDER_DEFAULT_NAME;

		let movingFolder = null;
		let movingFile = null;
		let file;

		const createFolder = () => {
		  const newFolder = {
		    name: addingFolderName,
		    parentId: currentFolder.id || 0
		  };
		  console.log(newFolder);
		  apiCreateFolder(newFolder).then(data => {
		    const newFolderWithId = { ...newFolder, id: data };
		    folders = folders
		      ? [...folders, newFolderWithId].sort(idOrdering)
		      : [newFolderWithId];
		    isAddingFolder = false;
		    addingFolderName = NEW_FOLDER_DEFAULT_NAME;
		  });
		};

		const deleteFolder = folderId => {
		  apiDeleteFolder(folderId).then(() => {
		    apiGetFolderContent(currentFolder && currentFolder.id).then(data => {
		      folders = data.folders && data.folders.sort(idOrdering);
		      files = data.files && data.files.sort(idOrdering);
		    });
		  });
		};

		const updateFolderName = (prevFolder, name) => {
		  apiUpdateFolder({ ...prevFolder, name }).catch(err => {
		    folders = [...folders].sort(idOrdering); // reinitialize the folders with its name before update trial
		  });
		};

		const openFolder = folderId => {
		  apiGetFolderContent(folderId).then(data => {
		    folders = data.folders && data.folders.sort(idOrdering);
		    files = data.files && data.files.sort(idOrdering);
		    currentFolder = data.currentFolder;
		  });
		};

		const redirectToDownload = file => {
		  window.open(getDownloadUrl(file), "_blank");
		};

		const uploadFile = event => {
		  const toUploadFiles = event.dataTransfer.files;
		  console.log(toUploadFiles);
		  for (let i = 0; i < toUploadFiles.length; i++) {
		    const toUploadFile = toUploadFiles[i];
		    apiUploadFile(toUploadFile, currentFolder.id).then(id => {
		      files = [...files, { id, name: toUploadFile.name }];
		    });
		  }
		};

		const startFolderMoveMode = folder => {
		  movingFolder = folder;
		};

		const stopMoveMode = () => {
		  movingFolder = null;
		  movingFile = null;
		};

		const moveFolder = () => {
		  apiMoveFolder(movingFolder.id, currentFolder.id).then(response => {
		    folders =
		      folders && folders.length
		        ? [...folders, movingFolder].sort(idOrdering)
		        : [movingFolder];
		    movingFolder = null;
		  });
		};

		const startFileMoveMode = file => {
		  movingFile = file;
		};

		const moveFile = () => {
		  apiMoveFile(movingFile.id, currentFolder.id).then(response => {
		    files =
		      files && files.length
		        ? [...files, movingFile].sort(idOrdering)
		        : [movingFile];
		    movingFile = null;
		  });
		};

		const deleteFile = fileId => {
		  apiDeleteFile(fileId).then(response => {
		    apiGetFolderContent().then(data => {
		      folders = data.folders && data.folders.sort(idOrdering);
		      files = data.files && data.files.sort(idOrdering);
		    });
		  });
		};

		apiGetFolderContent().then(data => {
		  folders = data.folders && data.folders.sort(idOrdering);
		  files = data.files && data.files.sort(idOrdering);
		});
</script>

<h1>Remote File System</h1>
<h2>
{#if currentFolder.id}
	<span class="folder-name" on:click={() => {
			openFolder(currentFolder.parentId);
		}
	}>
		[to parent]
	</span>
{/if}
	/{currentFolder.name}
</h2>

{#if movingFolder}
	<div class="notif-bar">
		<span>Drop "/{movingFolder.name}" here?</span>
		<button on:click={moveFolder}>Confirm</button>
		<button on:click={stopMoveMode}>Cancel</button>
	</div>
{/if}

{#if movingFile}
	<div class="notif-bar">
		<span>Drop "{movingFile.name}" here?</span>
		<button on:click={moveFile}>Confirm</button>
		<button on:click={stopMoveMode}>Cancel</button>
	</div>
{/if}

<div class="folder">
	{#if isAddingFolder===true}
		<div>
			<input bind:value={addingFolderName}/>
			<button on:click={() => createFolder()}>Create</button>
		</div>
	{:else}
		<button class="add-folder" on:click={() => isAddingFolder=true}>Add Folder</button>
	{/if}
</div>

<div>
	{#each folders || [] as folder}
		<Folder initialName={folder.name}
			on:update-name={(event) => updateFolderName(folder, event.detail)}
			on:delete={() => deleteFolder(folder.id)}
			on:click={() => openFolder(folder.id)}
			on:move={() => startFolderMoveMode(folder)}/>
	{/each}
</div>

<div>
	{#each files || [] as file}
		<File initialName={file.name}
			on:delete={() => deleteFile(file.id)}
			on:click={() => redirectToDownload(file)}
			on:move={() => startFileMoveMode(file)}/>
	{/each}
</div>

<div id="drop-zone"
	on:drop={(event) => {event.preventDefault(); uploadFile(event);}}
	on:dragover={(event) => event.preventDefault()}>
  <p>Drag one or more files to this Drop Zone ...</p>
</div>

<style>
	#drop-zone {
	  color: #19b4c9;
	  border: 2px dashed #19b4c9;
	  width: auto;
	  padding: 1em;
	  margin: 1em;
	  text-align: center;
	}

	.notif-bar {
	  background: rgb(255, 210, 150);
	  padding: 5px 0 0 15px;
	  margin: 5px;
	}

	.add-folder {
	  background: white;
	  color: green;
	  border: 1px solid green;
	}

	.folder {
	  padding: 10px 0px 0px 0px;
	}

	inline {
	  display: inline;
	}

	.folder-name {
	  color: grey;
	}
	.folder-name:hover {
	  color: black;
	  cursor: pointer;
	}
</style>
