# repository Specification

## Purpose

Home에서 다루는 코드 저장소를 이름으로 선언하고, 각 저장소의 clone이 이 기기의 어디에 있는지를 찾는다.

## Requirements

### Requirement: 저장소 선언
시스템은 Home 루트의 `homerepo.yaml`에서 저장소의 이름, 원격 주소, 기본 브랜치를 읽어야 한다(SHALL). 이 파일에는 실행되는 값을 두지 않는다.

#### Scenario: 선언된 저장소를 읽는다
- **WHEN** `homerepo.yaml`에 저장소 `frontend`가 원격과 기본 브랜치와 함께 적혀 있다
- **THEN** `homerepo repo list`가 `frontend`와 그 원격, clone의 경로, clone이 있는지 여부를 출력한다

#### Scenario: 선언 파일이 없다
- **WHEN** Home 루트에 `homerepo.yaml`이 없다
- **THEN** 명령이 0이 아닌 코드로 끝나고 파일이 없다는 것과 만드는 방법을 알린다

### Requirement: clone의 자리
시스템은 저장소의 clone을 Home의 옆 폴더(`<Home의 상위>/<이름>`)에서 찾아야 한다(SHALL). `.homerepo/config.yaml`의 `repositories.overrides`에 그 이름의 경로가 있으면 그 경로를 써야 한다(SHALL).

#### Scenario: 기본 자리
- **WHEN** 예외가 없는 저장소 `backend`의 경로를 묻는다
- **THEN** `homerepo repo path backend`가 Home 옆의 `backend` 폴더의 절대 경로를 출력한다

#### Scenario: 예외가 있다
- **WHEN** 로컬 설정에 `frontend`의 예외 경로가 적혀 있다
- **THEN** `homerepo repo path frontend`가 그 경로를 출력한다

#### Scenario: 선언되지 않은 이름
- **WHEN** `homerepo.yaml`에 없는 이름의 경로를 묻는다
- **THEN** 명령이 0이 아닌 코드로 끝나고 선언된 이름들을 알린다

### Requirement: 로컬 폴더는 Git에서 빠진다
시스템은 Home의 `.homerepo/`와 각 clone의 `.worktrees/`를 만들 때 그 안에 `*` 한 줄짜리 `.gitignore`를 두어야 한다(SHALL). 저장소의 `.gitignore`는 고치지 않는다.

#### Scenario: 만든 폴더가 상태에 나오지 않는다
- **WHEN** 시스템이 `.homerepo/`나 `.worktrees/`를 처음 만든다
- **THEN** 그 저장소의 `git status`에 그 폴더가 나오지 않는다

### Requirement: 초기화
`homerepo init`은 Home에 `homerepo.yaml`의 뼈대, `.homerepo/`, 스킬 파일을 만들어야 한다(SHALL). 이미 있는 파일은 덮어쓰지 않는다.

#### Scenario: 처음 초기화한다
- **WHEN** 아무것도 없는 Home에서 `homerepo init`을 실행한다
- **THEN** `homerepo.yaml`, `.homerepo/`, 스킬 폴더가 생기고 만든 파일의 목록이 출력된다

#### Scenario: 다시 초기화한다
- **WHEN** 이미 `homerepo.yaml`이 있는 Home에서 `homerepo init`을 실행한다
- **THEN** 그 파일의 내용이 바뀌지 않고, 건너뛴 파일이 출력에 표시된다
