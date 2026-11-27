# Proposal

## Why

스레드는 진행 중인 세션의 대화를 포크해 새 탭에서 이어 간다. 포크는 대화의 문맥을 통째로 들고 가므로 스레드마다 그 문맥만큼 토큰을 더 쓴다. 이 비용에 견줄 만큼 스레드가 쓰이지 않았다. worktree 명령은 스레드의 이름으로 worktree를 만들기 때문에 스레드 없이는 성립하지 않는다.

남기는 것은 Home Repo 구조를 코드로 정한 부분이다. Home 루트의 `homerepo.yaml`에 저장소를 선언하고, clone은 Home 옆 폴더에 두고, 기기마다 다른 값은 `.homerepo/`에 둔다.

## What Changes

- **BREAKING** `thread create`, `thread list`, `thread close`를 없앤다. 스레드의 훅, 실행 인자의 묶음, 추가 폴더 붙이기, Herdr 호출도 함께 없앤다.
- **BREAKING** `workspace create`, `workspace list`, `workspace release`를 없앤다. clone을 받는 명령은 두지 않는다.
- **BREAKING** `.homerepo/config.yaml`의 `runtimes`를 읽지 않는다. 이 키가 남아 있으면 명령은 그 키를 알리며 멈춘다.
- 시나리오 실행기(`dogfood`)와 그 사양을 없앤다.
- `homerepo init`이 설치하는 스킬의 이름을 `homerepo-thread`에서 `homerepo`로 바꾸고, 저장소 명령만 안내한다. 이미 설치된 `homerepo-thread` 스킬은 지우지 않는다.
- 이미 만들어진 `.worktrees/`, `.homerepo/threads/`, `.homerepo/hooks/`는 건드리지 않는다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `thread`: 요구사항을 모두 없앤다. 보관한 뒤 사양 폴더를 지운다.
- `workspace`: 요구사항을 모두 없앤다. 보관한 뒤 사양 폴더를 지운다.
- `dogfood`: 요구사항을 모두 없앤다. 보관한 뒤 사양 폴더를 지운다.
- `repository`: "clone의 자리"에 읽지 않는 설정 키의 시나리오를 더하고, "로컬 폴더는 Git에서 빠진다"에서 `.worktrees/`를 빼고, "초기화"에 스킬의 이름과 내용을 적는다.
- `cli`: "명령의 도움말"의 예를 남는 명령으로 바꾼다.

## Impact

- 지우는 코드: `internal/thread`, `internal/workspace`, `internal/herdr`, `internal/dogfood`, `dogfood/`, `internal/cli`의 스레드와 worktree 명령
- 옮기는 코드: 원격의 기본 브랜치를 읽는 함수(`repo add`가 쓴다)
- 설정: `.homerepo/config.yaml`에 `runtimes`를 둔 기기는 그 키를 지워야 한다
- 의존: Herdr가 필요 없어진다
