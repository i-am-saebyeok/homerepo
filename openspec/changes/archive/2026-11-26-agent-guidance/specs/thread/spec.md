# Spec Delta

## MODIFIED Requirements

### Requirement: 세션이 준비되기를 기다리지 않는다
`thread create`는 탭을 만들고, 그 탭의 창에서 포크 명령을 실행하고, 스레드 기록을 쓰고, 뒤의 훅이 있으면 그 훅이 끝난 뒤 끝나야 한다(SHALL). 세션이 입력을 받을 준비가 되기를 기다리지 않으며, 그 세션에 메시지나 키 입력을 보내지 않는다. 출력의 첫 줄은 스레드의 이름과 탭을 알린다. 포크 명령의 실행 파일 이름과 인자는 각각 따로 넘겨, 빈칸이 든 경로나 인자도 하나의 인자로 남는다. 탭의 셸이 아직 명령을 받지 못해 Herdr가 포크 명령을 거절하면 정해진 시간 안에서 다시 보낸다.

#### Scenario: 띄우고 돌아온다
- **WHEN** 훅이 없는 Home에서 `homerepo thread create login-fix`를 실행한다
- **THEN** 새 탭의 창에서 포크 명령이 실행되고, 명령은 세션의 준비를 기다리지 않고 0으로 끝나며 스레드 이름과 탭을 한 줄로 출력한다

#### Scenario: 세션이 확인 창에서 멈춘다
- **WHEN** 포크된 세션이 신뢰나 승인을 묻는 창을 띄운다
- **THEN** 시스템은 그 탭에 입력을 보내지 않는다. 그 창을 어떻게 다룰지는 사용자가 정하며, 훅을 두지 않았으면 사람이 그 탭에서 답한 뒤 말을 걸어 스레드를 시작한다

#### Scenario: 인자에 빈칸이 있다
- **WHEN** clone의 경로나 실행 인자의 묶음에 빈칸이 있다
- **THEN** 그 경로와 인자는 쪼개지지 않고 하나의 인자로 포크 명령에 넘겨진다

#### Scenario: 인자에 셸이 해석하는 글자가 있다
- **WHEN** 실행 인자의 묶음에 따옴표, `$`, 중괄호, 빈칸이 든 인자가 있다(예: `permissions.x.unix_sockets={"/tmp/a b.sock"="allow"}`)
- **THEN** 탭의 셸은 그 인자를 해석하지 않고 적힌 그대로 하나의 인자로 포크 명령에 넘긴다

#### Scenario: 탭의 셸이 아직 준비되지 않았다
- **WHEN** 탭을 만든 직후라 Herdr가 포크 명령을 `agent_pane_busy`나 `pane_send_failed`로 거절한다
- **THEN** 시스템은 같은 명령을 다시 보내고, 받아들여지면 스레드가 뜬다

#### Scenario: 끝내 받지 않는다
- **WHEN** 정해진 시간이 지나도록 Herdr가 포크 명령을 거절한다
- **THEN** 시스템은 다시 보내기를 멈추고, 탭은 만들었지만 세션을 띄우지 못했다는 것과 Herdr의 오류를 알린다

### Requirement: 포크와 함께 worktree 만들기
`thread create`에 `--repo <이름>`을 주면, 시스템은 세션을 띄우기 전에 그 저장소의 worktree를 만들어야 한다(SHALL). 여러 번 줄 수 있다. 출력의 첫 줄 다음에, 만든 worktree마다 저장소와 경로를 한 줄씩 출력한다.

#### Scenario: 저장소를 지정한다
- **WHEN** `homerepo thread create login-fix --repo frontend`를 실행한다
- **THEN** `frontend`의 clone에 `.worktrees/login-fix`가 생긴 뒤 스레드가 뜨고, 출력의 둘째 줄에 `frontend`와 그 worktree의 경로가 나온다

#### Scenario: 저장소를 주지 않는다
- **WHEN** `--repo` 없이 스레드를 만든다
- **THEN** 출력은 스레드의 이름과 탭을 알리는 한 줄뿐이다
