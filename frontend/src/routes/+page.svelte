<script>
	import { page } from '$app/stores';
	if ($page.data.apiAnswer) {
		let result = JSON.parse($page.data.apiAnswer);
	}
</script>

{#if $page.data.session}
	<div class="container">
		<div class="item">
			<h1>Welcome to Kubelab</h1>
			<p>Your Roles are: {$page.data.session?.user?.roles}</p>
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
