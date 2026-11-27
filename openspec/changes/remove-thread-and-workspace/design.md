# Design

## Context

명령은 `init`, `repo`, `workspace`, `thread`의 네 묶음이다. `thread`가 `workspace`와 `herdr`를 부르고, `workspace`는 스레드의 이름을 받아야 동작한다. `dogfood`는 Herdr의 탭에 런타임을 띄워 스레드의 흐름을 확인한다. `repo add`는 원격의 기본 브랜치를 읽는 함수를 `internal/workspace`에서 가져다 쓴다.

## Goals / Non-Goals

**Goals:**

- `thread`, `workspace`, `dogfood`와 그것들만 쓰던 코드를 지운다.
- 남는 명령(`init`, `repo list`, `repo path`, `repo add`)의 동작과 출력은 바꾸지 않는다.

**Non-Goals:**

- clone을 받는 명령을 더하지 않는다.
- 기기에 남은 스레드의 기록, 훅, worktree를 옮기거나 지우지 않는다.

## Decisions

### 기본 브랜치를 읽는 함수는 `internal/cli`로 옮긴다

`workspace.DefaultBranch`를 부르는 곳은 `repo add`뿐이다. 함수 하나를 위한 패키지를 두지 않고 `internal/cli/repo.go`의 내보내지 않는 함수로 둔다. `internal/home`은 명령을 실행하지 않는 패키지라 그쪽에는 두지 않는다.

### 읽지 않는 설정 키는 지금처럼 거절한다

`.homerepo/config.yaml`은 모르는 키가 있으면 읽기를 거절한다. `runtimes`를 설정의 모양에서 빼면 그 키가 남은 기기에서는 모든 명령이 그 키의 이름과 줄을 알리며 멈춘다. 키를 조용히 무시하면 스레드의 설정이 여전히 쓰이는 것처럼 보이므로 거절을 유지한다. 이 동작을 `repository` 사양의 시나리오로 적는다.

### 스킬은 새 이름으로만 설치하고 옛 스킬은 두지 않는다

`init`은 이미 있는 파일을 덮어쓰지 않는다는 규칙을 지킨다. `homerepo-thread`가 설치된 Home에서 `init`을 다시 실행하면 `homerepo` 스킬이 새로 생기고 옛 스킬은 남는다. 옛 스킬을 지우는 일은 사람이 한다. `init`이 파일을 지우기 시작하면 사람이 고친 스킬을 잃을 수 있다.

### 요구사항이 모두 없어진 사양은 보관한 뒤 폴더를 지운다

`thread`, `workspace`, `dogfood`는 요구사항이 하나도 남지 않는다. 변경을 보관한 뒤 세 사양 폴더를 지우고 `openspec validate --all --strict`로 확인한다.

### 쓰이지 않게 되는 것은 함께 지운다

`home.WorktreesDir`와 `WorktreesPath`, 설정의 `RuntimeConfig`와 `Switch`, `App`의 시각과 대기 함수처럼 스레드와 worktree만 쓰던 것은 남기지 않는다. 지운 뒤 `go mod tidy`로 의존을 맞춘다.

## Risks / Trade-offs

- `runtimes`가 남은 기기에서 모든 명령이 멈춘다 → 알림에 키의 이름이 나오고, 그 키를 지우면 된다.
- 실제 런타임으로 흐름을 확인하는 수단이 없어진다 → 남는 명령은 파일을 읽고 쓰는 것뿐이라 `go test`가 확인한다.
