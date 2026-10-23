{
  "description": "Pin a Sklair release in GitHub Actions, compile your site, and deploy the output to GitHub Pages."
}
---

# Deploying Sklair

This document tells you how to use CI to produce your website and then deploy it. Although this is mostly specific to GitHub, there is nothing stopping you from running the same commands using a different service.

`sklair build` is used to produce your website into a build directory, and then you may pass that directory to your hosting provider's deployment tool.

## Pin the compiler

Choose a specific Sklair release in your workflow. Installing the latest release on every build can change compiler behaviour even when your project has not changed.

The workflow below uses [Setup Sklair](https://github.com/numelon-oss/setup-sklair) to download a pinned release and add it to your runner's PATH. Set its `version` input to a release which supports your project.

!!! info "Setup Sklair support"

    The custom Setup Sklair action supports version 0.4.0 and above.

A versioned `$schema` in `sklair.json` helps your editor describe that version's configuration. It does not however install or select the compiler; the action's `version` input determines which version runs.

## Deploy to GitHub Pages

In your repository's **Settings -> Pages**, choose **GitHub Actions** as the publishing source. Add this workflow as `.github/workflows/website.yml`:

```yaml title=".github/workflows/website.yml"
name: Publish Sklair website

on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: read
  pages: read

concurrency:
  group: github-pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/configure-pages@v5

      - uses: numelon-oss/setup-sklair@v1
        with:
          version: '0.4.0'

      - name: Build website
        run: sklair build

      - uses: actions/upload-pages-artifact@v4
        with:
          path: build

  deploy:
    if: github.ref == 'refs/heads/main'
    needs: build
    runs-on: ubuntu-latest
    permissions:
      pages: write
      id-token: write
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - uses: actions/deploy-pages@v4
        id: deployment
```

The upload path must match your configured output directory. This example assumes `sklair.json` is at the repository root and its output is `build`. If your configuration is `src/sklair.json` with output `./build`, add `working-directory: src` to the build step and upload `src/build` instead.

Any push to the `main` branch of your repository will build and deploy the website. Manual runs on other branches build it without deploying.

## Other CI services

Download a pinned archive and its `SHA256SUMS` from [GitHub Releases](https://github.com/numelon-oss/sklair/releases), verify the archive's SHA-256 checksum, then extract it and add the executable to `PATH`.

Then, run `sklair build` and pass the configured output directory to your deployment tool.

The compiler does not depend on GitHub Actions. Note that GitHub Actions does support third-party runers.

## URLs and custom domains

A project Pages site normally lives under a path such as `https://owner.github.io/repository/`. A root-relative URL like `/styles/site.css` points to the domain root, not `/repository/styles/site.css`.

Use a custom domain when your site is authored for the domain root, or author its URLs for the repository path. Sklair does not automatically prefix all URLs with a deployment base path.

Configure a custom domain in Pages settings and set its DNS as GitHub requires. The Pages deployment uses that setting. See [GitHub's Pages workflow documentation](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages) for the deployment actions and permissions.
