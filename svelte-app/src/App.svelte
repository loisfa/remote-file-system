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
		    parentId: currentFolder.parentId || 0
		  };
		  console.log(newFolder);
		  apiCreateFolder(newFolder).then(data => {
		    const newFolderWithId = { ...newFolder, id: data };
		    folders = [...folders, newFolderWithId].sort(idOrdering);
		    isAddingFolder = false;
		    addingFolderName = NEW_FOLDER_DEFAULT_NAME;
		  });
		};

		const deleteFolder = folderId => {
		  console.log("delete folder: " + folderId);
		  apiDeleteFolder(folderId).then(() => {
		    apiGetFolderContent().then(content => {
		      folders = content.folders.sort(idOrdering);
		      files = content.files;
		    });
		  });
		};

		const updateFolderName = (prevFolder, name) => {
		  console.log("updated name: '" + name + "' on folder id: " + prevFolder.id);
		  apiUpdateFolder({ ...prevFolder, name }).catch(err => {
		    console.error("issue when updating name of folder: " + prevFolder.id);
		    folders = [...folders].sort(idOrdering); // reinitialize the folders with its name before update trial
		  });
		};

		const openFolder = currentFolderId => {
		  apiGetFolderContent(currentFolderId).then(data => {
		    folders = data.folders && data.folders.sort(idOrdering);
		    files = data.files;
		    currentFolder = data.currentFolder;
		  });
		};

		const redirectToDownload = file => {
		  window.open(getDownloadUrl(file), "_blank");
		};

		const uploadFile = event => {
		  const toUploadFiles = event.dataTransfer.files;
		  toUploadFiles.forEach(toUploadFile => {
		    apiUploadFile(toUploadFile, currentFolder.id);
		    // handle the update of the local list
		  });
		};

		const startFolderMoveMode = folder => {
		  movingFolder = folder;
		};

		const stopMoveMode = () => {
		  movingFolder = null;
		  moveFile = null;
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

		apiGetFolderContent().then(content => {
		  folders = content.folders;
		  files = content.files;
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

<div id="drop_zone"
	on:drop={(event) => {event.preventDefault(); uploadFile(event);}}
	on:dragover={(event) => event.preventDefault()}>
  <p>Drag one or more files to this Drop Zone ...</p>
</div>

<style>
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
