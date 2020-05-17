<script>
	import Folder from "./Folder.svelte";
	let name = "world";

	let initialFolderName = "folder 1";

	const deleteFolder = () => console.log("delete folder");
	const updateName = event => console.log("updated name: " + event.detail);
	const openFolder = () => console.log("opening folder");
</script>

<h1>Hello {name}!</h1>
<Folder initialName={initialFolderName}
	on:update-name={updateName}
	on:delete={deleteFolder}
	on:click={openFolder}/>
