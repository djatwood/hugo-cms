<script lang="ts" context="module">
	export async function load({ page, fetch }): Promise<{}> {
		const path = page.params.path;
		const ext = path.substring(path.lastIndexOf('.') + 1, path.length) || path;

		if (ext != 'md') {
			return {
				props: {
					ext,
					file: `//127.0.0.1:4120/${page.params.site}/${page.params.section}/${page.params.path}`
				}
			};
		}

		const result = await fetch(
			`/file.json?path=${page.params.site}/${page.params.section}/${page.params.path}`
		);

		const body = await result.text();
		return {
			props: {
				ext,
				siteKey: page.params.site,
				sectionKey: page.params.section,
				filepath: page.params.path,
				file: JSON.parse(body)
			}
		};
	}
</script>

<script lang="ts">
	export let siteKey, sectionKey, filepath, ext, file;
</script>

{#if ext == ''}
	<p class="text-lg mb-4">{sectionKey}/{filepath}</p>

	<ol>
		{#each file.data as { Path, Name }}
			<li class="py-2 border-t">
				<a href="/{siteKey}/{sectionKey}/{filepath}/{Path}">{Name}</a>
				<span class="float-right">{sectionKey}/{filepath}/{Path}</span>
			</li>
		{/each}
	</ol>
{:else if ext == 'md'}
	<div class="grid grid-cols-2">
		{#each Object.keys(file.frontmatter) as key}
			<p>{key}</p>
			<p>{file.frontmatter[key]}</p>
		{/each}
	</div>
	<pre>
    	{file.content}
	</pre>
{:else}
	<iframe class="w-full h-full" src={file} frameborder="0" />
{/if}
