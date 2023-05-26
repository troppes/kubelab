<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { getDeployments } from '$lib/kubelab-requests.js';
	import { scaleDeployment } from '../lib/kubelab-requests';

	let token = null;
	let deployments = { items: [] };
	if ($page.data.session) {
		token = $page.data.session.user.id_token;
	}

	// write onmount to fetch deployments
	onMount(async () => {
		try {
			deployments = getDeployments(token);
		} catch (error) {
			console.log(error);
		}
	});

	const scaleHandler = async (e) => {
		try {
			await scaleDeployment(token, e.srcElement.dataset.id);
			deployments = await getDeployments(token); // pull updated list
		} catch (error) {
			deployments = error;
		}
	};
</script>

{#if $page.data.session}
	<div class="container">
		<div class="item">
			<h1>Welcome to Kubelab</h1>
			<p>Your Roles are: {$page.data.session?.user?.roles}</p>
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
								</tr>
							</thead>
							<tbody>
								{#each deployments.items as deploy}
									<tr>
										<td>
											{deploy.metadata.name}
										</td>
										<td>
											{deploy.status.replicas == 1 ? 'On' : 'Off'}
										</td>
										<td>
											<button class="button" data-id={deploy.metadata.name} on:click={scaleHandler}
												>{deploy.status.replicas == 1 ? 'Stop' : 'Start'}</button
											>
										</td>
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
			<p>This is a protected content. You can access this content because you are signed in.</p>
			<p>Session expiry: {$page.data.session?.expires}</p>
		</div>
	</div>
{:else}
	<div class="container">
		<div class="item">
			<h3>Please use the Login-Button on the top of the page to continue.</h3>
		</div>
	</div>
{/if}
