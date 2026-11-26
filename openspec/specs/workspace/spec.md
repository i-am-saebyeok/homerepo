# workspace Specification

## Purpose

스레드가 코드 저장소에서 일할 때 쓰는 격리된 Git worktree를 만들고, 목록을 보고, 해제한다.

## Requirements

### Requirement: worktree 생성
`homerepo workspace create <저장소>`는 그 저장소의 clone 안 `.worktrees/<스레드 이름>`에 worktree를 만들어야 한다(SHALL). 브랜치의 이름은 스레드의 이름과 같고, 원격에서 받은 기본 브랜치에서 갈라진다. 사람의 원래 작업 폴더에는 fetch 말고 아무것도 하지 않는다.

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

### Requirement: worktree 목록
`homerepo workspace list`는 스레드의 worktree마다 저장소, 경로, 브랜치, 커밋하지 않은 변경이 있는지를 출력해야 한다(SHALL). 이 시스템이 만들지 않은 worktree는 구분해서 표시하고 다루지 않는다.

#### Scenario: 다른 도구가 만든 worktree
- **WHEN** 같은 clone에 `.worktrees/` 밖의 worktree가 있다
- **THEN** 목록에 "밖"으로 표시되고, 해제와 닫기의 대상이 되지 않는다

### Requirement: worktree 해제
`homerepo workspace release <저장소>`는 그 스레드의 worktree를 지워야 한다(SHALL). 커밋하지 않은 변경이 있으면 지우지 않고 알려야 한다(MUST). 브랜치는 지우지 않는다. Git이 무시하는 파일은 커밋하지 않은 변경으로 치지 않는다.

#### Scenario: 깨끗한 worktree
- **WHEN** 커밋하지 않은 변경이 없는 worktree를 해제한다
- **THEN** worktree 폴더가 지워지고 브랜치와 커밋은 남는다

#### Scenario: 변경이 남아 있다
- **WHEN** 추적되는 파일의 변경이나 추적되지 않는 파일이 있는 worktree를 해제한다
- **THEN** 아무것도 지우지 않고, 0이 아닌 코드로 끝나며 남은 파일을 알린다
