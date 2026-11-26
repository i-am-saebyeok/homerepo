# Spec Delta

## ADDED Requirements

### Requirement: 스레드 생성의 훅
시스템은 `.homerepo/hooks/pre-thread-create`가 있으면 요청의 검사를 마친 뒤 worktree와 탭을 만들기 전에, `.homerepo/hooks/post-thread-create`가 있으면 포크 명령을 실행한 뒤에 그 파일을 실행하고 끝나기를 기다려야 한다(SHALL). 파일이 없으면 아무 일도 하지 않는다. 실행할 수 없는 훅 파일은 0이 아닌 코드로 끝난 훅과 같이 다룬다.

#### Scenario: 훅이 없다
- **WHEN** `.homerepo/hooks/`에 두 파일이 모두 없다
- **THEN** `thread create`는 이 변경 전과 같이 동작한다

#### Scenario: 앞의 훅이 0으로 끝난다
- **WHEN** `pre-thread-create`가 있고 0으로 끝난다
- **THEN** 훅이 끝난 뒤에 worktree와 탭이 만들어지고 스레드가 뜬다

#### Scenario: 앞의 훅이 실패한다
- **WHEN** `pre-thread-create`가 0이 아닌 코드로 끝난다
- **THEN** 탭도 worktree도 기록도 생기지 않고, 명령은 어느 훅이 어떤 코드로 끝났는지 알리며 0이 아닌 코드로 끝난다

#### Scenario: 거절되는 요청
- **WHEN** 스레드 안에서 불렀거나, 이름이 겹치거나, 없는 저장소나 묶음을 준 요청이다
- **THEN** 어느 훅도 실행되지 않는다

#### Scenario: 뒤의 훅이 실패한다
- **WHEN** `post-thread-create`가 0이 아닌 코드로 끝난다
- **THEN** 탭, worktree, 기록은 그대로 남고, 명령은 스레드가 만들어졌다는 것과 훅이 어떤 코드로 끝났는지 알리며 0이 아닌 코드로 끝난다

#### Scenario: 실행할 수 없는 훅
- **WHEN** `pre-thread-create` 파일에 실행 권한이 없다
- **THEN** 아무것도 만들지 않고, 그 훅을 실행할 수 없다고 알리며 0이 아닌 코드로 끝난다

### Requirement: 훅이 받는 값
시스템은 훅을 Home을 현재 폴더로 해서 인자 없이 실행하고, 호출한 환경에 아래 환경 변수를 더해 넘겨야 한다(SHALL). 훅의 표준 출력과 표준 오류는 명령의 표준 오류로 보낸다.

| 환경 변수 | 값 | 받는 훅 |
| --- | --- | --- |
| `HOMEREPO_HOOK_THREAD` | 스레드의 이름 | 둘 다 |
| `HOMEREPO_HOOK_RUNTIME` | `claude` 또는 `codex` | 둘 다 |
| `HOMEREPO_HOOK_DIRS` | 세션에 붙는 추가 폴더의 경로. 한 줄에 하나 | 둘 다 |
| `HOMEREPO_HOOK_TAB` | 스레드의 탭 ID | `post-thread-create` |
| `HOMEREPO_HOOK_PANE` | 포크 명령을 실행한 창의 ID | `post-thread-create` |

#### Scenario: 뒤의 훅이 값을 받는다
- **WHEN** `post-thread-create`가 있고 `homerepo thread create login-fix`를 실행한다
- **THEN** 훅은 포크 명령이 실행된 뒤에 돌고, `HOMEREPO_HOOK_THREAD`는 `login-fix`, `HOMEREPO_HOOK_TAB`과 `HOMEREPO_HOOK_PANE`은 새 탭과 그 창의 ID다

#### Scenario: 앞의 훅에는 탭이 없다
- **WHEN** `pre-thread-create`가 실행된다
- **THEN** `HOMEREPO_HOOK_TAB`과 `HOMEREPO_HOOK_PANE`은 넘겨지지 않는다

#### Scenario: 추가 폴더가 여럿이다
- **WHEN** 선언된 저장소 둘의 clone이 있다
- **THEN** `HOMEREPO_HOOK_DIRS`는 두 저장소의 `<clone>/.worktrees` 경로를 한 줄에 하나씩 담는다

#### Scenario: 훅의 출력
- **WHEN** 훅이 표준 출력에 글을 쓴다
- **THEN** 그 글은 명령의 표준 오류로 나가고, 명령의 표준 출력은 스레드의 이름과 탭을 알리는 한 줄뿐이다

## MODIFIED Requirements

### Requirement: 세션이 준비되기를 기다리지 않는다
`thread create`는 탭을 만들고, 그 탭의 창에서 포크 명령을 실행하고, 스레드 기록을 쓰고, 뒤의 훅이 있으면 그 훅이 끝난 뒤 끝나야 한다(SHALL). 세션이 입력을 받을 준비가 되기를 기다리지 않으며, 그 세션에 메시지나 키 입력을 보내지 않는다. 출력은 스레드의 이름과 탭을 알리는 한 줄이다. 포크 명령의 실행 파일 이름과 인자는 각각 따로 넘겨, 빈칸이 든 경로나 인자도 하나의 인자로 남는다.

#### Scenario: 띄우고 돌아온다
- **WHEN** 훅이 없는 Home에서 `homerepo thread create login-fix`를 실행한다
- **THEN** 새 탭의 창에서 포크 명령이 실행되고, 명령은 세션의 준비를 기다리지 않고 0으로 끝나며 스레드 이름과 탭을 한 줄로 출력한다

#### Scenario: 세션이 확인 창에서 멈춘다
- **WHEN** 포크된 세션이 신뢰나 승인을 묻는 창을 띄운다
- **THEN** 시스템은 그 탭에 입력을 보내지 않는다. 그 창을 어떻게 다룰지는 사용자가 정하며, 훅을 두지 않았으면 사람이 그 탭에서 답한 뒤 말을 걸어 스레드를 시작한다

#### Scenario: 인자에 빈칸이 있다
- **WHEN** clone의 경로나 실행 인자의 묶음에 빈칸이 있다
- **THEN** 그 경로와 인자는 쪼개지지 않고 하나의 인자로 포크 명령에 넘겨진다
