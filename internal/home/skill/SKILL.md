---
name: homerepo-thread
description: Fork this session into a parallel thread, give a thread its own Git worktree, or declare a repository, with the homerepo command. Use when the user says things like "스레드로 빼줘", "쓰레드 만들어", "make a thread", "split this into a thread", or "make a worktree".
metadata:
  author: homerepo
---

Use `homerepo` to split work into threads (forked sessions in new Herdr tabs) and to give each thread an isolated worktree of a declared repository. Every command takes `--help`.

- `homerepo repo add <name> --remote <address> [--branch <branch>]` declares a repository in `homerepo.yaml`; without `--branch` it reads the remote's default branch
- `homerepo repo list`, `homerepo repo path <repository>`
- `homerepo thread create <name> [--repo <repository>]... [--mode <bundle>]` prints the tab, then `worktree <repository> <path>` for each `--repo`
- `homerepo thread list`, `homerepo thread close <name>`
- `homerepo workspace create <repository>` prints the worktree path; inside a thread, calling it again prints the same path
- `homerepo workspace list`, `homerepo workspace release <repository>`

Before `thread create`, write in one or two lines what the thread should do; the fork takes the conversation with it. Afterwards, tell the user the tab's name: the thread starts when the user talks to it in that tab.

If `HOMEREPO_THREAD` is set, you are inside that thread. Do not create threads; work in the path `homerepo workspace create` prints.

Do code work in a worktree.
