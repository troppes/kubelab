<script>
	import { page } from '$app/stores';
</script>

<div>
	<header>
		<div class="container-nav">
			{#if $page.data.session}
				<div class="item">
					<small>Signed in as</small><br />
					<strong>{$page.data.session.user?.email ?? $page.data.session.user?.name}</strong>
				</div>
				<div class="item"><div class="headline">Kubelab</div></div>
				<div class="item">
					<a href="/auth/signout" class="button" data-sveltekit-preload-data="off">Sign out</a>
				</div>
			{:else}
				<div class="item" />
				<div class="item"><div class="headline">Kubelab</div></div>
				<div class="item">
					<a href="/auth/signin" class="button" data-sveltekit-preload-data="off">Sign in</a>
				</div>
			{/if}
		</div>
	</header>
	<slot />
</div>
