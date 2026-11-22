// Preserve authorized child paths and query strings on layout mount/reload.
// A redirect is necessary only when the path has no accessible menu parent.
export function initialRouteRedirect(
  path: string,
  links: string[],
): string | undefined {
  const pathname = path.split('?')[0];
  if (
    links.some((link) => pathname === link || pathname.startsWith(link + '/'))
  )
    return undefined;
  return links[0] || '/overview';
}
