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
	<title>Hello world!</title>
</svelte:head>

<main>
	<ol>
		{#each sites as { Path, Name }}
			<li><a href={Path}>{Name}</a></li>
		{/each}
	</ol>
</main>
