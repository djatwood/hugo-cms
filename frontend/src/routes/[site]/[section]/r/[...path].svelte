<script lang="ts" context="module">
	export async function load({ page, fetch }): Promise<{}> {
		console.log(`http://127.0.0.1:4120/${page.params.site}/${page.params.section}/${page.params.path}`)
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


{#if file.kind == "dir"}
<ol>
{#each file.data as f}
<li><a href="/{siteKey}/{sectionKey}/r/{filepath}/{f}">{f}</a></li>
{/each}
</ol>
{:else}
<pre>
    {file.data}
</pre>
{/if}

