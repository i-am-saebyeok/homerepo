## REMOVED Requirements

### Requirement: worktree 생성
**Reason**: worktree는 스레드의 이름으로 만들어지므로 스레드와 함께 없앤다.
**Migration**: worktree는 `git worktree`로 직접 만들고 지운다. 이미 있는 `.worktrees/`의 worktree는 `git worktree remove`로 지운다.

### Requirement: worktree 목록
**Reason**: worktree는 스레드의 이름으로 만들어지므로 스레드와 함께 없앤다.
**Migration**: worktree는 `git worktree`로 직접 만들고 지운다. 이미 있는 `.worktrees/`의 worktree는 `git worktree remove`로 지운다.

### Requirement: worktree 해제
**Reason**: worktree는 스레드의 이름으로 만들어지므로 스레드와 함께 없앤다.
**Migration**: worktree는 `git worktree`로 직접 만들고 지운다. 이미 있는 `.worktrees/`의 worktree는 `git worktree remove`로 지운다.
