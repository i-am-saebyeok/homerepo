# Proposal

## Why

스레드를 닫을 때 사용자마다 함께 하려는 일이 있다. 예를 들어 런타임의 세션을 탭과 함께 죽이지 않고 런타임의 방식으로 끝내거나, 원격 제어 목록에 남는 세션을 정리하거나, 닫은 스레드를 기록해 두는 일이다. 스레드를 만들 때는 사용자가 정한 명령을 돌릴 자리가 있지만 닫을 때는 없다.

## What Changes

- `thread close`가 `.homerepo/hooks/pre-thread-close`가 있으면 닫을 수 있는지 검사한 뒤, worktree를 해제하고 탭을 닫기 전에 실행한다. 0이 아닌 코드로 끝나면 아무것도 닫지 않는다.
- `thread close`가 `.homerepo/hooks/post-thread-close`가 있으면 탭을 닫은 뒤에 실행한다. 0이 아닌 코드로 끝나면 스레드는 닫힌 채로 두고 실패를 알린다.
- 닫기의 훅도 생성의 훅과 같은 환경 변수를 받는다. 닫기 전의 훅은 세션이 도는 창의 ID를 받는다.
- 훅 파일이 없으면 지금과 똑같이 동작한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `thread`: 요구사항 "스레드 닫기의 훅"을 더하고, "훅이 받는 값"에 닫기의 훅을 더한다.

## Impact

- `internal/thread`의 스레드 닫기 순서와 훅 실행
- 스레드 기록에는 탭의 ID만 있으므로, 닫기 전의 훅에 넘길 창의 ID는 Herdr에 물어 얻는다.
