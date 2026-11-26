# Proposal

## Why

시나리오 시험(`go run ./dogfood`)에서 Claude Code와 Codex의 에이전트가 사람의 말만 받고 도구를 썼을 때, 둘 다 같은 곳에서 헤맸다.

- 저장소를 Home에 등록하는 명령이 없고 `homerepo.yaml`의 형식을 알려 주는 곳도 없었다. 에이전트는 실행 파일의 문자열이나 소스를 읽어 필드를 추정했고, `branch`를 빠뜨렸다가 오류를 보고 고쳤다.
- `homerepo thread create --help`가 사용법 대신 `flag: help requested`와 종료 코드 2를 냈다.
- `thread create --repo`가 만든 worktree의 경로를 알려 주지 않아, 스레드가 `workspace create`를 다시 불러 경로를 얻었다. 다시 불러도 되는지는 스킬에 없었다.
- 빈 원격에는 선언한 브랜치가 아직 없어 worktree를 만들 수 없다는 것을 에이전트가 추측으로 알아냈다.

## What Changes

- `homerepo repo add <이름> --remote <주소> [--branch <브랜치>]`가 `homerepo.yaml`에 저장소를 더한다. `--branch`가 없으면 원격의 기본 브랜치를 읽어 쓰고, 읽을 수 없으면 `--branch`를 요구한다.
- 모든 명령이 `--help`와 `-h`에 그 명령의 사용법을 표준 출력에 내고 0으로 끝난다.
- `thread create`가 스레드의 이름과 탭을 알리는 줄 다음에, 만든 worktree마다 저장소와 경로를 한 줄씩 출력한다.
- 선언한 브랜치가 원격에 없으면 `workspace create`가 그 사실과, 원격에 그 브랜치의 첫 커밋을 올려야 한다는 것을 알린다.
- 설치되는 스킬에 `repo add`, `thread create`가 출력하는 worktree 경로, 스레드 안에서 `workspace create`를 다시 불러도 같은 경로가 나온다는 것을 적는다.

## Capabilities

### New Capabilities

- `cli`: 명령의 사용법과 도움말.

### Modified Capabilities

- `repository`: 요구사항 "저장소 등록"을 더한다.
- `workspace`: 요구사항 "worktree 생성"에 원격에 브랜치가 없을 때를 더한다.
- `thread`: "세션이 준비되기를 기다리지 않는다"의 출력과 "포크와 함께 worktree 만들기"에 worktree 경로의 출력을 더한다.

## Impact

- `internal/cli`의 명령과 도움말, `internal/home`의 선언 파일 쓰기와 내장 스킬, `internal/workspace`의 오류, `internal/thread`의 출력
- 이미 설치된 스킬은 `homerepo init`이 덮어쓰지 않으므로 사용자가 직접 바꾼다.
