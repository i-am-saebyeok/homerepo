# Proposal

## Why

런타임이 띄운 세션에는 시작할 때 사용자마다 다르게 처리하려는 일이 있다. 예를 들어 Claude Code는 추가 폴더를 받으면 그 폴더를 신뢰하는지 묻는 창을 띄우고, 답하기 전에는 원격 제어 같은 기능이 켜지지 않는다. 처리하는 방법은 여럿이고(저장소를 미리 신뢰해 둔다, 런타임의 설정에 폴더를 등록한다, 창에 답한다) 사용자마다 고르는 것이 다르다. 도구가 그 가운데 하나를 고르지 않고, 사용자가 정한 명령을 돌릴 자리를 준다.

## What Changes

- `thread create`가 `.homerepo/hooks/pre-thread-create`가 있으면 탭과 worktree를 만들기 전에 실행한다. 0이 아닌 코드로 끝나면 스레드를 만들지 않는다.
- `thread create`가 `.homerepo/hooks/post-thread-create`가 있으면 포크 명령을 실행한 뒤에 실행한다. 0이 아닌 코드로 끝나면 스레드는 그대로 두고 실패를 알린다.
- 훅은 스레드의 이름, 런타임, 세션에 붙는 추가 폴더를 환경 변수로 받는다. 뒤의 훅은 탭과 창의 ID도 받는다.
- 훅 파일이 없으면 지금과 똑같이 동작한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `thread`: 요구사항 "스레드 생성의 훅"과 "훅이 받는 값"을 더하고, "세션이 준비되기를 기다리지 않는다"에 훅을 기다린다는 것을 더한다.

## Impact

- `internal/thread`의 스레드 생성 순서
- `.homerepo/hooks/`라는 새 자리. `.homerepo/`는 이미 Git에서 빠져 있으므로 훅은 사용자 개인의 것이다.
- 설정 파일의 모양은 바뀌지 않는다.
