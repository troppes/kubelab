<script ssr="false">
	import { onMount, onDestroy } from 'svelte';
	import { getDeployments, scaleDeployment, getConnectionString } from '$lib/kubelab-requests.js';
	import { toast } from '@zerodevx/svelte-toast';

	export let token;
	let interval;
	let deployments = { items: [] };

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

	const renewDeployments = async () => {
		deployments = await getDeployments(token);
	};

	// write onmount to fetch deployments
	onMount(async () => {
		try {
			renewDeployments();
			interval = setInterval(renewDeployments, 5000);
		} catch (error) {
			console.log(error);
		}
	});

	onDestroy(() => {
		// Clean up the interval when the component is destroyed
		clearInterval(interval);
	});

	const connectionHandler = async (e) => {
		try {
			let string = await getConnectionString(token, e.srcElement.dataset.id);
			navigator.clipboard
				.writeText(string)
				.then(() => successToast('Copied!'))
				.catch((e) => errorToast(e));
		} catch (error) {
			console.log(error);
		}
	};

	const scaleHandler = async (e) => {
		try {
			await scaleDeployment(token, e.srcElement.dataset.id);
			deployments = await getDeployments(token); // pull updated list
		} catch (error) {
			deployments = error;
		}
	};
</script>

<div class="item">
	{#await deployments}
		<div>
			<p>Fetching Classrooms ...</p>
		</div>
	{:then deployments}
		<div>
			<div>
				<table>
					<thead>
						<tr>
							<th>Name</th>
							<th>Status</th>
							<th>Action</th>
							<th>Connection</th>
						</tr>
					</thead>
					<tbody>
						{#each deployments.items as deploy}
							<tr>
								<td>
									{deploy.metadata.name}
								</td>
								<td>
									{deploy.spec.replicas == 1 ? 'On' : 'Off'}
								</td>
								<td>
									<button class="button" data-id={deploy.metadata.name} on:click={scaleHandler}
										>{deploy.spec.replicas == 1 ? 'Stop' : 'Start'}</button
									>
								</td>
								<td>
									<button
										class="button"
										data-id={deploy.metadata.name}
										disabled={!(deploy.status.availableReplicas == 1)}
										on:click={connectionHandler}>Connect</button
									>
								</td>
							</tr>
							<tr class="details" data-id={deploy.metadata.name}>
								<td colspan="4" />
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{:catch error}
		<div>
			<p style="color: red">Error loading deployments.</p>
			<p style="color: red">Error message: {error.body.message}</p>
		</div>
	{/await}
</div>

<style>
	.details {
		display: none;
	}
</style>
