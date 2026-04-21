# progress

## text filled out so far
- `index.md`: main identity blurb, homepage structure, and top-level direction
- `other.md`: organized idea backlog with sections (`firethief`, `trails`, `gragent`, `diffeq club`, etc.)
- `progress.md`: now tracking completed work + next steps
- created skeleton pages for homepage links: `cv`, `portfolio`, `skills`, `socials`, `literate programming`, `microservices`, `architecture`

## site/content state
- markdown rendering now uses goldmark (better standard markdown support)
- routes are set up for core pages and new skeleton pages
- homepage links now point to route paths (ex: `/portfolio`) instead of loose `.md` paths
- output can be published to `docs/` for GitHub Pages
- typography was adjusted toward a more default, readable web style
- Open Sans is loaded from Google Fonts for a cleaner minimalist look

## what to do next
- fill out each skeleton page with first-pass content (short and useful before polish)
- decide page structure conventions (titles, blurbs, sections, links, footer style)
- tighten organization of idea pages: separate active projects vs future ideas vs essays

## tooling roadmap from homepage + other.md
- `listit`: structure complex page elements
- `nav`: navigation and route model
- `head`: page header and metadata conventions
- `quest`: question/checklist sections that map to answer pages
- `blurby`: consistent subheading and blurb format
- `pub`: publishing pipeline from markdown to live site
- `toc`: optional table-of-contents block for long pages
- `styl`: minimalist readability defaults
- `linc`: internal linking conventions between related pages
- `footy`: lightweight footer pattern
- `oigit`: simple version-history surface tied to git
- `authentik`: admin/auth layer for future editing workflow
- `treemaker`: structure/generate future pages