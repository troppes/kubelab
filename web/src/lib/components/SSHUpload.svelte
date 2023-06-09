<script ssr="false">
	import { postSSHToken } from '$lib/kubelab-requests.js';
	import { toast } from '@zerodevx/svelte-toast';

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

	export let token;

	let file = null;

	const uploadHandler = async (e) => {
		try {
			file = e.target.files[0];
			const formData = new FormData();
			formData.append('file', file);
			await postSSHToken(token, formData);
			successToast('File uploaded correctly.');
		} catch (error) {
			errorToast('Something went wrong!');
			console.log(error);
		}
	};
</script>

<div>
	<div class="upload-wrapper">
		<input type="file" id="upload" class="upload-input" on:change={uploadHandler} />
		<label for="upload" class="upload-label button"> Upload SSH Key </label>
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
</style>
