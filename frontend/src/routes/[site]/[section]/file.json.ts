
import { load } from "js-yaml"

/**
 * @type {import('@sveltejs/kit').RequestHandler}
 */
export async function get({ params, query }) {
    const result = await fetch(
        `http://127.0.0.1:4120/${params.site}/${params.section}/${query.get('path')}`
    );

    const file = await result.json();
    if (file.kind == "dir") return { body: file };

    const data = file.data.split("---", 3)
    return {
        body: Object.assign(file, { frontmatter: load(data[1]), content: data[2].slice(1) })
    }
}