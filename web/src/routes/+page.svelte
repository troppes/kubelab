<script>
	import { page } from '$app/stores';
	import DeploymentTable from '$lib/components/DeploymentTable.svelte';
	import ClassTable from '$lib/components/ClassTable.svelte';
	import SshUpload from '$lib/components/SSHUpload.svelte';
	import { onMount, onDestroy } from 'svelte';
	import { getDeployments, scaleDeployment, getConnectionString } from '$lib/kubelab-requests.js';

	let deployments = { items: [] };
	let interval = null;

	let token = null;
	if ($page.data.session) {
		token = $page.data.session.user.id_token;
	}

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
</script>

{#if $page.data.session}
	{#if $page.data.session.user.roles.includes('teacher')}
		<div class="container">
			<h1>Welcome to Kubelab for Teachers</h1>
			<p>Your Roles are: {$page.data.session?.user?.roles}</p>

			<h2>You classes</h2>
			<ClassTable {token} />
		</div>
	{:else}
		<div class="container">
			<h1>Welcome to Kubelab</h1>
			<p>Your Roles are: {$page.data.session?.user?.roles}</p>
			<SshUpload {token} />
			<DeploymentTable {token} {deployments} {scaleDeployment} {getConnectionString} />
		</div>
	{/if}
{:else}
	<div class="container">
		<div class="item">
			<h3>Please use the Login-Button on the top of the page to continue.</h3>
		</div>
	</div>
{/if}
