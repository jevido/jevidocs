---
title: Private projects
description: Who can read a private project, and how to give access.
position: 5
---

A project with **Public** switched off is invisible to everyone who has no
access: its pages, search, `llms.txt`, sitemap and MCP read tools answer
"not found", exactly like a project that does not exist.

There are three ways in.

| Who | How |
| --- | --- |
| Admins | Always. Sign in on the docs site (or send an admin token). |
| Members | Editors and viewers you add under **Access → Members**. They sign in on the docs site. |
| Email domains | Every user whose email is at a domain listed under **Access → Email domains**. They sign in on the docs site. |
| Share links | Anyone with a link from **Access → Share links**. No account needed. |

## Signing in on the docs site

The docs site has a **Sign in** button in the navbar. It uses the same
accounts as the admin. Signed-in readers see the private projects they may
read, marked **Private**, including in the [project directory](/p).

## Members

Add users (created on the admin's Users page) as members of a private
project. Admins do not need to be members. Removing a member takes access
away on their next request.

## Email domains

List domains such as `example.com` and every account whose email address
ends in `@example.com` can read the project, without adding people one by
one. New colleagues get access as soon as their account exists; changing
someone's email to another domain takes it away.

- The match is exact and case-insensitive: `example.com` does not include
  `mail.example.com`; list that separately if you need it.
- Accounts are still created by an admin on the Users page (there is no
  self sign-up), so a domain only admits people you have given an account.

## Share links

A share link is a URL like
`https://jevidocs.jevido.app/p/<project>?share=jds_…`. Opening it stores the
token in that browser and removes it from the address bar; after that the
project stays readable there until the link expires or is revoked.

<Callout type="warn" title="Treat links like passwords">
Anyone who has the URL can read the project. Give links a name per
recipient, pick an expiry (7, 30 or 90 days, or never), and revoke a link
when it is no longer needed. The admin shows when each link was last used.
</Callout>

## For API and MCP clients

Send `Authorization: Bearer <token>` (a user's API token) or
`X-Jevidocs-Share: <share token>` with any public read request or MCP call.

Draft [preview links](/docs/admin#draft-previews) work for private projects
too: a preview link is its own permission for that one page.
