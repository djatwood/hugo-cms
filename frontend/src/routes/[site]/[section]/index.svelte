<script lang="ts" context="module">
	export async function load({ page, fetch, session, context }): Promise<{}> {
		const result = await fetch(`http://127.0.0.1:4120/${page.params.site}/${page.params.section}`);
		return {
			props: {
				siteKey: page.params.site,
				sectionKey: page.params.section,
				section: await result.json()
			}
		};
	}
</script>

<script lang="ts">
	export let siteKey, sectionKey, section;
</script>

{#each section as name}
	<li><a href="/{siteKey}/{sectionKey}/{name}">{name}</a></li>
{/each}
