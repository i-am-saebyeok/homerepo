# Spec Delta

## ADDED Requirements

### Requirement: 추가 폴더 붙이기
시스템은 clone이 있거나 `--repo`로 준 저장소의 `<clone>/.worktrees`를 포크 명령에 추가 폴더 인자로 붙여야 한다(SHALL). `.homerepo/config.yaml`의 `runtimes.<런타임>.attach`가 `false`면 그 런타임의 포크 명령에 추가 폴더 인자를 붙이지 않는다. 이때 세션이 접근할 폴더는 사용자가 런타임의 설정이나 실행 인자의 묶음으로 정한다.

#### Scenario: 설정이 없다
- **WHEN** 그 런타임에 `attach`가 없거나 `true`다
- **THEN** 포크 명령에 각 저장소의 `<clone>/.worktrees`가 추가 폴더 인자로 붙는다

#### Scenario: 붙이지 않는다
- **WHEN** `runtimes.claude.attach`가 `false`이고 Claude Code 세션에서 스레드를 만든다
- **THEN** 포크 명령에 추가 폴더 인자가 없다

#### Scenario: 런타임마다 따로 정한다
- **WHEN** `runtimes.claude.attach`가 `false`이고 `runtimes.codex`에는 `attach`가 없다
- **THEN** Codex 세션에서 만든 스레드의 포크 명령에는 추가 폴더 인자가 붙는다

#### Scenario: 묶음으로 직접 붙인다
- **WHEN** `attach`가 `false`이고 `default` 묶음이 `["--add-dir", "/work"]`다
- **THEN** 포크 명령의 추가 폴더 인자는 묶음에 적힌 `--add-dir /work`뿐이다

#### Scenario: 참도 거짓도 아닌 값
- **WHEN** `attach`에 `true`와 `false`가 아닌 값이 적혀 있다
- **THEN** 아무것도 만들지 않고 설정 파일의 어느 값이 잘못됐는지 알린다

## MODIFIED Requirements

### Requirement: 스레드 생성
`homerepo thread create <이름>`은 호출한 세션의 대화를 포크해 Herdr의 새 탭에 띄워야 한다(SHALL). 포크된 세션은 Home에서 시작한다. 호출한 세션의 탭에는 입력을 보내지 않으며 포커스를 옮기지 않는다.

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

### Requirement: 훅이 받는 값
시스템은 훅을 Home을 현재 폴더로 해서 인자 없이 실행하고, 호출한 환경에 아래 환경 변수를 더해 넘겨야 한다(SHALL). 훅의 표준 출력과 표준 오류는 명령의 표준 오류로 보낸다.

| 환경 변수 | 값 | 받는 훅 |
| --- | --- | --- |
| `HOMEREPO_HOOK_THREAD` | 스레드의 이름 | 둘 다 |
| `HOMEREPO_HOOK_RUNTIME` | `claude` 또는 `codex` | 둘 다 |
| `HOMEREPO_HOOK_DIRS` | clone이 있거나 `--repo`로 준 저장소의 `<clone>/.worktrees` 경로. 한 줄에 하나 | 둘 다 |
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

#### Scenario: 추가 폴더 인자를 붙이지 않는 설정
- **WHEN** 그 런타임의 `attach`가 `false`다
- **THEN** `HOMEREPO_HOOK_DIRS`는 `attach`가 없을 때와 같은 경로를 담는다

#### Scenario: 훅의 출력
- **WHEN** 훅이 표준 출력에 글을 쓴다
- **THEN** 그 글은 명령의 표준 오류로 나가고, 명령의 표준 출력은 스레드의 이름과 탭을 알리는 한 줄뿐이다
