# thread Specification

## Purpose

진행 중인 에이전트 세션의 대화를 포크해 별도의 탭에서 이어 가는 스레드를 만들고, 목록을 보고, 닫는다.

## Requirements

### Requirement: 스레드 생성
`homerepo thread create <이름>`은 호출한 세션의 대화를 포크해 Herdr의 새 탭에 띄워야 한다(SHALL). 포크된 세션은 Home에서 시작하고, 선언된 각 저장소의 `<clone>/.worktrees`를 추가 폴더로 받는다. 호출한 세션의 탭에는 입력을 보내지 않으며 포커스를 옮기지 않는다.

#### Scenario: Claude Code 세션에서 만든다
- **WHEN** Herdr 안의 Claude Code 세션에서 `homerepo thread create login-fix`를 실행한다
- **THEN** 새 탭에 그 대화를 포크한 Claude Code 세션이 뜨고, 호출한 탭의 포커스와 크기는 그대로다

#### Scenario: Codex 세션에서 만든다
- **WHEN** Herdr 안의 Codex 세션에서 같은 명령을 실행한다
- **THEN** 새 탭에 그 대화를 포크한 Codex 세션이 뜬다

#### Scenario: 런타임을 환경에서 알아낸다
- **WHEN** 환경에 `CODEX_SESSION_ID`가 있고 `CLAUDECODE`가 없다
- **THEN** 런타임은 Codex이고 `CODEX_SESSION_ID`의 값을 세션 ID로 쓴다

#### Scenario: 두 런타임의 환경이 함께 있다
- **WHEN** 환경에 `CLAUDECODE=1`과 `CODEX_SESSION_ID`가 함께 있고 `--runtime`이 없다
- **THEN** 아무것도 만들지 않고 `--runtime`을 요구한다

#### Scenario: 런타임이나 멀티플렉서를 알 수 없다
- **WHEN** 에이전트 세션 밖이거나 Herdr 밖에서 실행한다
- **THEN** 아무것도 만들지 않고, 무엇을 찾지 못했는지 알리며 0이 아닌 코드로 끝난다

#### Scenario: 같은 이름의 스레드가 있다
- **WHEN** 이미 있는 스레드의 이름으로 만든다
- **THEN** 아무것도 만들지 않고 이름이 겹친다고 알린다

#### Scenario: 이름이 겹치는 스레드가 있다
- **WHEN** 스레드 `fix`가 있는데 `fix/login`을 만들거나, `fix/login`이 있는데 `fix`를 만든다
- **THEN** 아무것도 만들지 않고 두 이름이 Git의 브랜치로 함께 있을 수 없다고 알린다

### Requirement: 스레드 안에서 만들기를 거절한다
환경에 `HOMEREPO_THREAD`가 있으면 `thread create`는 아무것도 만들지 않고 0이 아닌 코드로 끝나야 한다(SHALL). 이미 그 스레드 안에 있다는 것과 `homerepo workspace create <저장소>`로 worktree를 만들어 거기서 일하라는 것을 알린다.

#### Scenario: 스레드 안에서 부른다
- **WHEN** 스레드 `login-fix`의 세션에서 `homerepo thread create`를 실행한다
- **THEN** 탭도 worktree도 기록도 생기지 않고, 이미 `login-fix` 안에 있다는 것과 worktree를 만드는 명령을 알린다

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

### Requirement: 실행 인자의 묶음
시스템은 `.homerepo/config.yaml`의 `runtimes.<런타임>.thread.<묶음>`에 적힌 인자를 포크 명령에 덧붙여야 한다(SHALL). `--mode`가 없으면 `default` 묶음을 쓰고, `default`가 없으면 인자를 붙이지 않는다. 각 인자 안의 `{thread}`는 모두 스레드의 이름으로 바꿔야 한다(SHALL). 다른 중괄호 글은 바꾸지 않으며, 바꾼 인자도 하나의 인자로 넘긴다.

#### Scenario: 묶음을 고른다
- **WHEN** `--mode yolo`를 주고, 그 런타임의 `thread.yolo`에 인자가 적혀 있다
- **THEN** 포크된 세션이 그 인자와 함께 뜬다

#### Scenario: 없는 묶음
- **WHEN** 설정에 없는 묶음 이름을 준다
- **THEN** 아무것도 만들지 않고 있는 묶음 이름들을 알린다

#### Scenario: 인자에 스레드 이름을 넣는다
- **WHEN** `default` 묶음이 `["--remote-control", "{thread}"]`이고 스레드 `login-fix`를 만든다
- **THEN** 포크 명령에 `--remote-control login-fix`가 붙는다

#### Scenario: 다른 글과 섞인 자리표시
- **WHEN** 묶음의 한 인자가 `name={thread}`이거나 `{thread}`가 두 번 나오는 `{thread}-{thread}`이다
- **THEN** 각각 `name=login-fix`와 `login-fix-login-fix` 하나의 인자로 넘겨진다

#### Scenario: 다른 중괄호 글
- **WHEN** 묶음의 인자에 `{repo}`처럼 `{thread}`가 아닌 중괄호 글이 있다
- **THEN** 그 글은 바뀌지 않고 그대로 넘겨진다

#### Scenario: 자리표시가 없다
- **WHEN** 묶음에 `{thread}`가 없거나 묶음이 없다
- **THEN** 포크 명령의 인자는 이 변경 전과 같다

### Requirement: 포크와 함께 worktree 만들기
`thread create`에 `--repo <이름>`을 주면, 시스템은 세션을 띄우기 전에 그 저장소의 worktree를 만들어야 한다(SHALL). 여러 번 줄 수 있다.

#### Scenario: 저장소를 지정한다
- **WHEN** `homerepo thread create login-fix --repo frontend`를 실행한다
- **THEN** `frontend`의 clone에 `.worktrees/login-fix`가 생긴 뒤 스레드가 뜬다

### Requirement: 스레드 목록과 닫기
`homerepo thread list`는 스레드의 이름, 런타임, 탭, worktree를 출력해야 한다(SHALL). `homerepo thread close <이름>`은 그 스레드의 worktree를 해제하고 탭을 닫아야 한다(SHALL). 해제할 수 없는 worktree가 있으면 탭을 닫지 않는다.

#### Scenario: 깨끗한 스레드를 닫는다
- **WHEN** 모든 worktree에 커밋하지 않은 변경이 없는 스레드를 닫는다
- **THEN** worktree가 지워지고 탭이 닫히며 목록에서 사라진다. 브랜치는 남는다

#### Scenario: 커밋하지 않은 변경이 있다
- **WHEN** 한 worktree에 커밋하지 않은 변경이 있는 스레드를 닫는다
- **THEN** 그 worktree와 탭이 그대로 남고, 어느 worktree에 변경이 있는지 알린다

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
