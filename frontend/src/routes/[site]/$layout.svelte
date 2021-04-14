<script lang="ts" context="module">
	export async function load({ page, fetch }): Promise<{}> {
		const result = await fetch(`http://127.0.0.1:4120/${page.params.site}`);
		return {
			props: {
				siteKey: page.params.site,
				site: await result.json()
			}
		};
	}
</script>

<script lang="ts">
	export let siteKey, site;
</script>

<svelte:head>
	<title>{siteKey} - Hugo CMS</title>
</svelte:head>

<div class="flex">
	<aside class="flex-shrink-0 h-screen p-12 border-r">
		<h2 class="text-2xl mb-8">{siteKey}</h2>
		<h3 class="text-xl mb-3">Sections</h3>
		<ol>
			{#each Object.keys(site.Sections) as key}
				<li class="mb-1"><a href="/{siteKey}/{key}">{site.Sections[key].Label}</a></li>
			{/each}
		</ol>
	</aside>
	<main class="w-full h-screen max-w-screen-lg p-12 overflow-y-auto">
		<slot />
	</main>
</div>
