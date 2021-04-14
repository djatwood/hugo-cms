<script lang="ts" context="module">
	export async function load({ fetch }): Promise<{}> {
		const result = await fetch('http://127.0.0.1:4120');
		return {
			props: {
				sites: await result.json()
			}
		};
	}
</script>

<script lang="ts">
	export let sites: { Path; Name }[];
</script>

<svelte:head>
	<title>Hugo CMS</title>
</svelte:head>

<main class="p-12">
	<div class="max-w-screen-lg mx-auto">
		<h1 class="text-3xl mb-4">Sites</h1>
		<ol>
			{#each sites as { Path, Name }}
				<li class="border-t py-2">
					<a href={Path}>{Name}</a>
					<span class="float-right">{Path}</span>
				</li>
			{/each}
		</ol>
	</div>
</main>
