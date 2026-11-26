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
`thread create`는 탭을 만들고, 그 탭의 창에서 포크 명령을 실행하고, 스레드 기록을 쓴 뒤 끝나야 한다(SHALL). 세션이 입력을 받을 준비가 되기를 기다리지 않으며, 그 세션에 메시지나 키 입력을 보내지 않는다. 출력은 스레드의 이름과 탭을 알리는 한 줄이다. 포크 명령의 실행 파일 이름과 인자는 각각 따로 넘겨, 빈칸이 든 경로나 인자도 하나의 인자로 남는다.

#### Scenario: 띄우고 돌아온다
- **WHEN** `homerepo thread create login-fix`를 실행한다
- **THEN** 새 탭의 창에서 포크 명령이 실행되고, 명령은 세션의 준비를 기다리지 않고 0으로 끝나며 스레드 이름과 탭을 한 줄로 출력한다

#### Scenario: 세션이 확인 창에서 멈춘다
- **WHEN** 포크된 세션이 신뢰나 승인을 묻는 창을 띄운다
- **THEN** 시스템은 그 탭에 입력을 보내지 않고, 사람이 그 탭에서 답한 뒤 말을 걸어 스레드를 시작한다

#### Scenario: 인자에 빈칸이 있다
- **WHEN** clone의 경로나 실행 인자의 묶음에 빈칸이 있다
- **THEN** 그 경로와 인자는 쪼개지지 않고 하나의 인자로 포크 명령에 넘겨진다

### Requirement: 실행 인자의 묶음
시스템은 `.homerepo/config.yaml`의 `runtimes.<런타임>.thread.<묶음>`에 적힌 인자를 포크 명령에 덧붙여야 한다(SHALL). `--mode`가 없으면 `default` 묶음을 쓰고, `default`가 없으면 인자를 붙이지 않는다.

#### Scenario: 묶음을 고른다
- **WHEN** `--mode yolo`를 주고, 그 런타임의 `thread.yolo`에 인자가 적혀 있다
- **THEN** 포크된 세션이 그 인자와 함께 뜬다

#### Scenario: 없는 묶음
- **WHEN** 설정에 없는 묶음 이름을 준다
- **THEN** 아무것도 만들지 않고 있는 묶음 이름들을 알린다

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
