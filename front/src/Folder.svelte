<script>
	import { createEventDispatcher } from "svelte";

	export let initialName;

	let name;
	$: name = initialName;

	export let isEditing = false;
	let disabled = false;

	const dispatch = createEventDispatcher();
	const onUpdateName = () => dispatch("update-name", name);
	const onDelete = () => dispatch("delete");
	const onClick = () => dispatch("click");
	const onClickMove = () => dispatch("move");
</script>

<div class="folder" on:click={onClick}>
	{#if isEditing===false}
		<span class="folder-name center-v">/{name}</span>
		<button class="center-v m-5" on:click={(event) => {isEditing = true; event.stopPropagation();} }>
			Rename
		</button>
		<button class="center-v m-5" on:click={(event) => {onClickMove(); event.stopPropagation();}}>
			Move
		</button>
		<button class="delete-button center-v m-5" on:click={(event) => {onDelete(); event.stopPropagation();} }>
			Delete
		</button>
	{:else}
		<input on:click={(event) => event.stopPropagation()} bind:value={name}>
		<button on:click={(event) => {isEditing = false; onUpdateName(); event.stopPropagation();} }>
			Validate
		</button>
	{/if}

</div>

<style>
	.folder {
	  padding: 5px 5px 5px 10px;
	  margin: 5px 0 5px 0;
	  border: 1px solid rgb(200, 200, 200);
	  background: rgb(250, 250, 250);
	  display: flex;
	  border-radius: 2px;
	}
	.folder:hover {
	  background: rgb(245, 245, 245);
	  cursor: pointer;
	}

	.folder-name {
	  color: grey;
	}

	.center-v {
	  align-self: center;
	}

	.m-5 {
	  margin: 5px;
	}

	.delete-button {
	  margin-left: auto;
	  color: red;
	  border: 1px solid red;
	  background: white;
	}
</style>
