<script ssr="false">
	import { toast } from '@zerodevx/svelte-toast';
	export let token;
	export let currentClass;
	export let files = [];
	export let deleteFiles;
	export let uploadFiles;

	const errorToast = (message) => {
		toast.push(message, {
			theme: {
				'--toastColor': 'mintcream',
				'--toastBackground': '#f27474',
				'--toastBarBackground': '#fa5555'
			}
		});
	};

	const successToast = (message) => {
		toast.push(message, {
			theme: {
				'--toastColor': 'mintcream',
				'--toastBackground': 'rgba(72,187,120,0.9)',
				'--toastBarBackground': '#2F855A'
			}
		});
	};

	const uploadHandler = async (e) => {
		try {
			const files = e.target.files;
			const formData = new FormData();
			for (let file of files) {
				const blob = new Blob([file]);
				formData.append(file.name, blob);
			}
			await uploadFiles(token, formData, currentClass);
			successToast('File uploaded correctly.');
		} catch (error) {
			errorToast('Something went wrong!');
			console.log(error);
		}
	};

	const deleteHandler = async (e) => {
		try {
			const file = e.srcElement.dataset.id;
			await deleteFiles(token, [{ name: file }], currentClass);
			successToast('File deleted correctly.');
		} catch (error) {
			errorToast('Something went wrong!');
			console.log(error);
		}
	};
</script>

<div>
	<h1>Files</h1>
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Size</th>
				<th>Action</th>
			</tr>
		</thead>
		<tbody>
			{#each files as file}
				<tr>
					<td>
						{file.name}
					</td>
					<td>
						{file.size}
					</td>
					<td>
						<button class="button" data-id={file.name} on:click={deleteHandler}>Delete</button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>

	<div class="upload-wrapper">
		<input type="file" id="upload" class="upload-input" on:change={uploadHandler} multiple />
		<label for="upload" class="upload-label button"> Upload new File </label>
	</div>
</div>

<style>
	.upload-wrapper {
		position: relative;
	}

	.upload-input {
		position: absolute;
		left: -9999px;
	}

	.upload-label {
		display: inline-block;
		cursor: pointer;
	}

	table {
		border-spacing: 10px;
		border-collapse: separate;
		text-align: center;
	}
</style>
