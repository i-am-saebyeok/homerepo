---
name: homerepo
description: Declare a code repository in this Home, list the declared repositories, or find where a repository's clone is, with the homerepo command. Use when the user says things like "저장소 등록해줘", "레포 추가해", "declare a repository", "add a repo", or "where is the clone".
metadata:
  author: homerepo
---

Use `homerepo` to declare the code repositories this Home works with and to find their clones. A clone lives next to Home, in `<parent of Home>/<name>`, unless `.homerepo/config.yaml` overrides its path. Every command takes `--help`.

- `homerepo repo add <name> --remote <address> [--branch <branch>]` declares a repository in `homerepo.yaml`; without `--branch` it reads the remote's default branch. It does not clone
- `homerepo repo list` prints each repository's remote, clone path, and whether the clone is present
- `homerepo repo path <repository>` prints the path of the clone

To clone a declared repository, run `git clone <remote> "$(homerepo repo path <name>)"`.
