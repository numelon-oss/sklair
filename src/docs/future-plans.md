{
  "description": "Future plans for Sklair, effectively a to-do list for the developers."
}
---

# Future plans

## Package distribution

Sklair is currently distributed as standalone binaries through GitHub Releases. Native Homebrew, WinGet, and APT channels are planned so installations and upgrades can follow the conventions of each platform.

Those package channels will remain wrappers around the same self-contained binary.

## Update awareness

Package managers should own upgrades when they installed Sklair. The compiler may still report that a newer release exists, while leaving the actual update operation to the installation method that owns the binary.

## Documentation tooling

The documentation site will continue to serve readable HTML alongside direct links to its Markdown sources. More structured discovery for tools and agents or something later if needed.
