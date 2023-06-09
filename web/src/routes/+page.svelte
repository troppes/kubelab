<script>
	import { page } from '$app/stores';
	import DeploymentTable from '$lib/components/DeploymentTable.svelte';
	import SshUpload from '$lib/components/SSHUpload.svelte';

	let token = null;
	if ($page.data.session) {
		token = $page.data.session.user.id_token;
	}
</script>

{#if $page.data.session}
	<div class="container">
		<h1>Welcome to Kubelab</h1>
		<p>Your Roles are: {$page.data.session?.user?.roles}</p>
		<SshUpload {token} />
		<DeploymentTable {token} />
	</div>
{:else}
	<div class="container">
		<div class="item">
			<h3>Please use the Login-Button on the top of the page to continue.</h3>
		</div>
	</div>
{/if}
