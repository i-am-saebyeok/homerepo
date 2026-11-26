---
name: homerepo-thread
description: Fork this session into a parallel thread or give a thread its own Git worktree with the homerepo command. Use when the user says things like "스레드로 빼줘", "split this into a thread", or "make a worktree".
metadata:
  author: homerepo
---

Use `homerepo` to split work into threads (forked sessions in new Herdr tabs) and to give each thread an isolated worktree of a declared repository.

- `homerepo thread create <name> [--repo <repository>]... [--mode <bundle>]`
- `homerepo thread list`
- `homerepo thread close <name>`
- `homerepo workspace create <repository>`
- `homerepo workspace list`
- `homerepo workspace release <repository>`
- `homerepo repo list`
- `homerepo repo path <repository>`

Before `thread create`, write in one or two lines what the thread should do; the fork takes the conversation with it. Afterwards, tell the user the tab's name: the thread starts when the user talks to it in that tab.

If `HOMEREPO_THREAD` is set, you are inside that thread. Do not create threads; work in the path `homerepo workspace create` prints.

Do code work in a worktree.
