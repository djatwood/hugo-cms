<script lang="ts" context="module">
	export async function load({ page, fetch }): Promise<{}> {
		const result = await fetch(
			`/file.json?path=${page.params.site}/${page.params.section}/${page.params.path}`
		);

		return {
			props: {
				siteKey: page.params.site,
				sectionKey: page.params.section,
				filepath: page.params.path,
				file: await result.json()
			}
		};
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
				<a href="/{siteKey}/{sectionKey}/{filepath}/{Path}">{Name}</a>
				<span class="float-right">{sectionKey}/{filepath}/{Path}</span>
			</li>
		{/each}
	</ol>
{:else}
	<div class="grid grid-cols-2">
		{#each Object.keys(file.frontmatter) as key}
			<p>{key}</p>
			<p>{file.frontmatter[key]}</p>
		{/each}
	</div>
	<pre>
    	{file.content}
	</pre>
{/if}
