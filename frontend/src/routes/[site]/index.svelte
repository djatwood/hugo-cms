<script lang="ts" context="module">
	export async function load({ page, fetch, session, context }): Promise<{}> {
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

<ol>
	{#each Object.keys(site.Sections) as key}
		<li><a href="/{siteKey}/{key}">{site.Sections[key].Label}</a></li>
	{/each}
</ol>
