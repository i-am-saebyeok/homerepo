## MODIFIED Requirements

### Requirement: clone의 자리
시스템은 저장소의 clone을 Home의 옆 폴더(`<Home의 상위>/<이름>`)에서 찾아야 한다(SHALL). `.homerepo/config.yaml`의 `repositories.overrides`에 그 이름의 경로가 있으면 그 경로를 써야 한다(SHALL). 이 파일에 시스템이 읽지 않는 키가 있으면 명령은 아무것도 하지 않고 그 키를 알려야 한다(MUST).

#### Scenario: 기본 자리
- **WHEN** 예외가 없는 저장소 `backend`의 경로를 묻는다
- **THEN** `homerepo repo path backend`가 Home 옆의 `backend` 폴더의 절대 경로를 출력한다

#### Scenario: 예외가 있다
- **WHEN** 로컬 설정에 `frontend`의 예외 경로가 적혀 있다
- **THEN** `homerepo repo path frontend`가 그 경로를 출력한다

#### Scenario: 선언되지 않은 이름
- **WHEN** `homerepo.yaml`에 없는 이름의 경로를 묻는다
- **THEN** 명령이 0이 아닌 코드로 끝나고 선언된 이름들을 알린다

#### Scenario: 읽지 않는 키가 있다
- **WHEN** `.homerepo/config.yaml`에 `runtimes`가 남아 있는 Home에서 `homerepo repo list`를 실행한다
- **THEN** 명령이 0이 아닌 코드로 끝나고 그 파일과 `runtimes`를 알린다

### Requirement: 로컬 폴더는 Git에서 빠진다
시스템은 Home의 `.homerepo/`를 만들 때 그 안에 `*` 한 줄짜리 `.gitignore`를 두어야 한다(SHALL). Home의 `.gitignore`는 고치지 않는다.

#### Scenario: 만든 폴더가 상태에 나오지 않는다
- **WHEN** 시스템이 `.homerepo/`를 처음 만든다
- **THEN** Home의 `git status`에 그 폴더가 나오지 않는다

### Requirement: 초기화
`homerepo init`은 Home에 `homerepo.yaml`의 뼈대, `.homerepo/`, 스킬 `homerepo`의 파일을 만들어야 한다(SHALL). 스킬은 저장소를 선언하고 찾는 명령만 안내한다. 이미 있는 파일은 덮어쓰지 않고, 어떤 파일도 지우지 않는다.

#### Scenario: 처음 초기화한다
- **WHEN** 아무것도 없는 Home에서 `homerepo init`을 실행한다
- **THEN** `homerepo.yaml`, `.homerepo/`, 스킬 `homerepo`의 폴더가 생기고 만든 파일의 목록이 출력된다

#### Scenario: 다시 초기화한다
- **WHEN** 이미 `homerepo.yaml`이 있는 Home에서 `homerepo init`을 실행한다
- **THEN** 그 파일의 내용이 바뀌지 않고, 건너뛴 파일이 출력에 표시된다

#### Scenario: 옛 스킬이 있다
- **WHEN** 스킬 `homerepo-thread`가 설치된 Home에서 `homerepo init`을 실행한다
- **THEN** 스킬 `homerepo`가 새로 생기고 `homerepo-thread`의 파일은 그대로 남는다
