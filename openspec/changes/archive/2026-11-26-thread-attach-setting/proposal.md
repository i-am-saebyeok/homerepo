# Proposal

## Why

도구는 스레드를 띄울 때 저장소의 `.worktrees`를 추가 폴더 인자로 항상 붙인다. 세션에 어떤 폴더를 어떻게 붙일지는 사용자마다 다르다.

- 런타임의 설정 파일로 붙이는 사용자. Claude Code는 인자로 받은 폴더마다 신뢰를 묻는 창을 띄우지만, 신뢰된 프로젝트의 설정에 선언된 폴더에는 묻지 않는다.
- 저장소들이 모인 상위 폴더를 통째로 붙이는 사용자. 스레드 안에서 새 저장소를 만들고 이어서 작업하려면 아직 없는 폴더까지 열려 있어야 한다.
- 지금처럼 worktree만 붙이는 사용자.

인자를 끌 수 없으면 앞의 두 방식은 쓸 수 없다. 설정에 선언한 폴더도 인자로 다시 받으면 창이 뜨기 때문이다.

## What Changes

- `.homerepo/config.yaml`의 `runtimes.<런타임>.attach`가 `false`면 그 런타임의 포크 명령에 추가 폴더 인자를 붙이지 않는다.
- `attach`가 없거나 `true`면 지금과 똑같이 동작한다.
- 훅이 받는 `HOMEREPO_HOOK_DIRS`는 `attach`의 값과 무관하게 같은 경로를 담는다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `thread`: 요구사항 "추가 폴더 붙이기"를 더하고, "스레드 생성"에서 추가 폴더의 문장을 그리로 옮기고, "훅이 받는 값"의 `HOMEREPO_HOOK_DIRS`를 다시 정의한다.

## Impact

- `internal/home`의 개인 설정 읽기, `internal/thread`의 포크 인자
- `.homerepo/config.yaml`에 `runtimes.<런타임>.attach`라는 새 키. 기존 설정 파일은 그대로 동작한다.
