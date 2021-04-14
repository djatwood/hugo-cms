<script lang="ts" context="module">
	export async function load({ page, fetch }): Promise<{}> {
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

<p class="text-lg mb-4">{sectionKey}</p>
<ol>
	{#each section as name}
		<li class="py-2 border-t">
			<a href="/{siteKey}/{sectionKey}/r/{name}">{name}</a>
		</li>
	{/each}
</ol>
