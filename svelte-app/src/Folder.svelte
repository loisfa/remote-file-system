<script>
	import { createEventDispatcher } from "svelte";

	export let initialName;
	let name = initialName;
	let isEditing = false;

	const dispatch = createEventDispatcher();
	const onUpdateName = () => dispatch("update-name", name);
	const onDelete = () => dispatch("delete");
	const onClick = () => dispatch("click");
</script>

<div class="inline">
	{#if isEditing===false}
		<span class="folder-name">{name}</span>
		<button on:click={() => {isEditing = true} }>
			Edit
		</button>
		<button on:click={() => {onDelete();} }>
			Delete
		</button>
	{:else}
		<input bind:value={name}>
		<button on:click={() => {isEditing = false; onUpdateName();} }>
			Validate
		</button>
	{/if}

</div>

<style>
	inline {
	  display: inline;
	}

	folder-name {
	  color: grey;
	}
	folder-name:hover {
	  color: black;
	}
</style>
