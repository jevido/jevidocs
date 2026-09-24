import type { PageRef, TreeNode } from './api'

export function containsSlug(node: TreeNode, slug: string): boolean {
  if (node.type === 'page') return node.slug === slug
  if (node.type === 'separator') return false
  if (node.index?.slug === slug) return true
  return node.children.some((c) => containsSlug(c, slug))
}

export function flatten(nodes: TreeNode[], out: PageRef[] = []): PageRef[] {
  for (const n of nodes) {
    if (n.type === 'page') out.push(n)
    else if (n.type === 'folder') {
      if (n.index) out.push(n.index)
      flatten(n.children, out)
    }
  }
  return out
}
