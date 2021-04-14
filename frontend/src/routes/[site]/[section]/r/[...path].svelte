<script lang="ts" context="module">
	export async function load({ page, fetch }): Promise<{}> {
		console.log(
			`http://127.0.0.1:4120/${page.params.site}/${page.params.section}/${page.params.path}`
		);
		const result = await fetch(
			`http://127.0.0.1:4120/${page.params.site}/${page.params.section}/${page.params.path}`
		);
		return {
			props: {
				siteKey: page.params.site,
				sectionKey: page.params.section,
				filepath: page.params.path,
				file: await result.json()
			}
		};
		return {};
	}
</script>

<script lang="ts">
	export let siteKey, sectionKey, filepath, file;
</script>

{#if file.kind == 'dir'}
	<p class="text-lg mb-4">{sectionKey}/{filepath}</p>

	<ol>
		{#each file.data as { Path, Name }}
			<li class="py-2 border-t">
				<a href="/{siteKey}/{sectionKey}/r/{filepath}/{Path}">{Name}</a>
				<span class="float-right">{sectionKey}/{filepath}/{Path}</span>
			</li>
		{/each}
	</ol>
{:else}
	<pre>
    {file.data}
</pre>
{/if}
