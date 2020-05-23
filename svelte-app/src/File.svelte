<script>
	import { createEventDispatcher } from "svelte";

	export let initialName;

	let disabled = false;

	const dispatch = createEventDispatcher();
	const onDelete = () => dispatch("delete");
	const onClick = () => dispatch("click");
	const onClickMove = () => dispatch("move");
</script>

<div class="file" on:click={onClick}>
	<span class="file-name center-v">{initialName}</span>
	<button class="center-v m-5" on:click={(event) => {onClickMove(); event.stopPropagation();}}>
		Move
	</button>
	<button class="delete-button center-v m-5" on:click={() => {onDelete(); event.stopPropagation();} }>
		Delete
	</button>
</div>

<style>
	.file {
	  padding: 0px 0px 0px 15px;
	  display: flex;
	}
	.file:hover {
	  background: rgb(250, 250, 250);
	  cursor: pointer;
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
