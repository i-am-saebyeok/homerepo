# Spec Delta

## ADDED Requirements

### Requirement: 스레드 닫기의 훅
시스템은 `.homerepo/hooks/pre-thread-close`가 있으면 닫을 수 있는지의 검사를 마친 뒤 worktree를 해제하고 탭을 닫기 전에, `.homerepo/hooks/post-thread-close`가 있으면 탭을 닫은 뒤에 그 파일을 실행하고 끝나기를 기다려야 한다(SHALL). 파일이 없으면 아무 일도 하지 않는다. 실행할 수 없는 훅 파일은 0이 아닌 코드로 끝난 훅과 같이 다룬다.

#### Scenario: 훅이 없다
- **WHEN** `.homerepo/hooks/`에 두 파일이 모두 없다
- **THEN** `thread close`는 이 변경 전과 같이 동작한다

#### Scenario: 닫기 전의 훅이 0으로 끝난다
- **WHEN** `pre-thread-close`가 있고 0으로 끝난다
- **THEN** 훅이 끝난 뒤에 worktree가 해제되고 탭이 닫힌다

#### Scenario: 닫기 전의 훅이 실패한다
- **WHEN** `pre-thread-close`가 0이 아닌 코드로 끝난다
- **THEN** worktree, 탭, 기록이 그대로 남고, 명령은 어느 훅이 어떤 코드로 끝났는지 알리며 0이 아닌 코드로 끝난다

#### Scenario: 닫을 수 없는 스레드
- **WHEN** 커밋하지 않은 변경이 있는 worktree가 있거나 없는 스레드 이름을 준 요청이다
- **THEN** 어느 훅도 실행되지 않는다

#### Scenario: 닫은 뒤의 훅이 실패한다
- **WHEN** `post-thread-close`가 0이 아닌 코드로 끝난다
- **THEN** 스레드는 닫힌 채로 남고, 명령은 스레드가 닫혔다는 것과 훅이 어떤 코드로 끝났는지 알리며 0이 아닌 코드로 끝난다

#### Scenario: 훅이 세션을 먼저 끝낸다
- **WHEN** `pre-thread-close`가 받은 창에서 런타임의 세션을 끝내고 0으로 끝난다
- **THEN** 탭 닫기는 그대로 진행되고 스레드가 닫힌다

## MODIFIED Requirements

### Requirement: 훅이 받는 값
시스템은 훅을 Home을 현재 폴더로 해서 인자 없이 실행하고, 호출한 환경에 아래 환경 변수를 더해 넘겨야 한다(SHALL). 훅의 표준 출력과 표준 오류는 명령의 표준 오류로 보낸다.

| 환경 변수 | 값 | 받는 훅 |
| --- | --- | --- |
| `HOMEREPO_HOOK_THREAD` | 스레드의 이름 | 모두 |
| `HOMEREPO_HOOK_RUNTIME` | `claude` 또는 `codex` | 모두 |
| `HOMEREPO_HOOK_DIRS` | clone이 있거나 `--repo`로 준 저장소의 `<clone>/.worktrees` 경로. 한 줄에 하나 | 모두 |
| `HOMEREPO_HOOK_TAB` | 스레드의 탭 ID | `pre-thread-create`를 뺀 모두 |
| `HOMEREPO_HOOK_PANE` | 스레드의 세션이 도는 창의 ID | `post-thread-create`, `pre-thread-close` |

#### Scenario: 뒤의 훅이 값을 받는다
- **WHEN** `post-thread-create`가 있고 `homerepo thread create login-fix`를 실행한다
- **THEN** 훅은 포크 명령이 실행된 뒤에 돌고, `HOMEREPO_HOOK_THREAD`는 `login-fix`, `HOMEREPO_HOOK_TAB`과 `HOMEREPO_HOOK_PANE`은 새 탭과 그 창의 ID다

#### Scenario: 앞의 훅에는 탭이 없다
- **WHEN** `pre-thread-create`가 실행된다
- **THEN** `HOMEREPO_HOOK_TAB`과 `HOMEREPO_HOOK_PANE`은 넘겨지지 않는다

#### Scenario: 추가 폴더가 여럿이다
- **WHEN** 선언된 저장소 둘의 clone이 있다
- **THEN** `HOMEREPO_HOOK_DIRS`는 두 저장소의 `<clone>/.worktrees` 경로를 한 줄에 하나씩 담는다

#### Scenario: 추가 폴더 인자를 붙이지 않는 설정
- **WHEN** 그 런타임의 `attach`가 `false`다
- **THEN** `HOMEREPO_HOOK_DIRS`는 `attach`가 없을 때와 같은 경로를 담는다

#### Scenario: 훅의 출력
- **WHEN** 훅이 표준 출력에 글을 쓴다
- **THEN** 그 글은 명령의 표준 오류로 나가고, 명령의 표준 출력은 스레드의 이름과 탭을 알리는 한 줄뿐이다

#### Scenario: 닫기 전의 훅이 창을 받는다
- **WHEN** 탭이 열려 있는 스레드 `login-fix`를 닫는다
- **THEN** `pre-thread-close`는 `HOMEREPO_HOOK_THREAD`로 `login-fix`를, `HOMEREPO_HOOK_TAB`과 `HOMEREPO_HOOK_PANE`으로 그 탭과 창의 ID를 받는다

#### Scenario: 탭이 이미 닫혀 있다
- **WHEN** 사람이 탭을 먼저 닫은 스레드를 닫는다
- **THEN** `pre-thread-close`는 `HOMEREPO_HOOK_PANE`을 받지 않는다

#### Scenario: 닫은 뒤의 훅에는 창이 없다
- **WHEN** `post-thread-close`가 실행된다
- **THEN** `HOMEREPO_HOOK_TAB`은 닫힌 탭의 ID이고 `HOMEREPO_HOOK_PANE`은 넘겨지지 않는다
