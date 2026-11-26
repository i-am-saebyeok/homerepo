# Spec Delta

## Purpose

진행 중인 에이전트 세션의 대화를 포크해 별도의 탭에서 이어 가는 스레드를 만들고, 목록을 보고, 닫는다.

## ADDED Requirements

### Requirement: 스레드 생성
`homerepo thread create <이름>`은 호출한 세션의 대화를 포크해 Herdr의 새 탭에 띄워야 한다(SHALL). 포크된 세션은 Home에서 시작하고, 선언된 각 저장소의 `<clone>/.worktrees`를 추가 폴더로 받는다. 호출한 세션의 탭에는 입력을 보내지 않으며 포커스를 옮기지 않는다.

#### Scenario: Claude Code 세션에서 만든다
- **WHEN** Herdr 안의 Claude Code 세션에서 `homerepo thread create login-fix`를 실행한다
- **THEN** 새 탭에 그 대화를 포크한 Claude Code 세션이 뜨고, 호출한 탭의 포커스와 크기는 그대로다

#### Scenario: Codex 세션에서 만든다
- **WHEN** Herdr 안의 Codex 세션에서 같은 명령을 실행한다
- **THEN** 새 탭에 그 대화를 포크한 Codex 세션이 뜬다

#### Scenario: 런타임이나 멀티플렉서를 알 수 없다
- **WHEN** 에이전트 세션 밖이거나 Herdr 밖에서 실행한다
- **THEN** 아무것도 만들지 않고, 무엇을 찾지 못했는지 알리며 0이 아닌 코드로 끝난다

#### Scenario: 같은 이름의 스레드가 있다
- **WHEN** 이미 있는 스레드의 이름으로 만든다
- **THEN** 아무것도 만들지 않고 이름이 겹친다고 알린다

### Requirement: 첫 메시지
시스템은 포크된 세션이 입력을 받을 준비가 되면 첫 메시지를 보내야 한다(SHALL). 첫 메시지는 이 세션이 스레드라는 것, 스레드의 이름, 코드 작업은 worktree에서 한다는 것, worktree를 만드는 명령을 알린다.

#### Scenario: 준비된 뒤에 보낸다
- **WHEN** 포크된 세션이 입력 대기 상태가 된다
- **THEN** 첫 메시지가 그 세션에 제출된다

#### Scenario: 세션이 확인 창에서 멈춘다
- **WHEN** 포크된 세션이 신뢰나 승인을 묻는 창을 띄워 준비되지 않는다
- **THEN** 시스템은 그 창에 입력을 보내지 않고, 사람이 그 탭에서 답해야 한다고 알린다

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
