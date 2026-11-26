# Spec Delta

## MODIFIED Requirements

### Requirement: worktree 생성
`homerepo workspace create <저장소>`는 그 저장소의 clone 안 `.worktrees/<스레드 이름>`에 worktree를 만들어야 한다(SHALL). 브랜치의 이름은 스레드의 이름과 같고, 원격에서 받은 기본 브랜치에서 갈라진다. 사람의 원래 작업 폴더에는 fetch 말고 아무것도 하지 않는다. 선언한 브랜치가 원격에 없으면 아무것도 만들지 않고, 그 브랜치가 원격에 없다는 것과 그 브랜치에 첫 커밋을 올려야 한다는 것을 알린다.

#### Scenario: 스레드 안에서 만든다
- **WHEN** 스레드 `login-fix`의 세션에서 `homerepo workspace create frontend`를 실행한다
- **THEN** `frontend`의 clone에 `.worktrees/login-fix`가 브랜치 `login-fix`로 생기고 그 경로가 출력된다

#### Scenario: 스레드 밖에서 만든다
- **WHEN** 스레드가 아닌 세션에서 `--thread` 없이 실행한다
- **THEN** 아무것도 만들지 않고 스레드의 이름이 필요하다고 알린다

#### Scenario: clone이 없다
- **WHEN** 그 저장소의 clone이 찾는 자리에 없다
- **THEN** 그 자리에 원격에서 clone한 뒤 worktree를 만든다

#### Scenario: 이미 있다
- **WHEN** 그 스레드의 그 저장소 worktree가 이미 있다
- **THEN** 새로 만들지 않고 있는 worktree의 경로를 출력하며 0으로 끝난다

#### Scenario: 원래 작업 폴더에 변경이 있다
- **WHEN** 사람의 원래 작업 폴더에 커밋하지 않은 변경이 있다
- **THEN** worktree는 만들어지고 원래 작업 폴더의 파일과 브랜치는 바뀌지 않는다

#### Scenario: 원격에 선언한 브랜치가 없다
- **WHEN** 선언한 브랜치 `main`이 아직 원격에 없다(빈 원격)
- **THEN** worktree와 브랜치가 생기지 않고, 원격에 `main`이 없으니 첫 커밋을 올려야 한다고 알리며 0이 아닌 코드로 끝난다
