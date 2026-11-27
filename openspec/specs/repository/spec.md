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

### Requirement: 저장소 등록
`homerepo repo add <이름> --remote <주소> [--branch <브랜치>]`는 `homerepo.yaml`의 `repositories` 목록 끝에 저장소를 더해야 한다(SHALL). clone은 받지 않는다. `--branch`가 없으면 원격의 기본 브랜치를 읽어 쓰고, 읽을 수 없으면 아무것도 바꾸지 않고 `--branch`를 요구한다. 파일의 다른 줄은 바꾸지 않는다. 더한 뒤의 파일이 선언 파일의 검사를 통과하지 못하면 원래 파일을 그대로 둔다.

#### Scenario: 저장소를 등록한다
- **WHEN** `homerepo repo add notes --remote https://example.com/notes.git --branch main`을 실행한다
- **THEN** `homerepo.yaml`에 `notes`가 그 원격과 브랜치로 더해지고, `homerepo repo list`에 나온다

#### Scenario: 빈 목록에 처음 등록한다
- **WHEN** `homerepo.yaml`이 `homerepo init`이 만든 `repositories: []`뿐이다
- **THEN** 그 줄이 저장소 하나를 가진 목록으로 바뀐다

#### Scenario: 브랜치를 원격에서 읽는다
- **WHEN** `--branch` 없이 기본 브랜치가 `trunk`인 원격을 등록한다
- **THEN** 저장소가 브랜치 `trunk`로 더해진다

#### Scenario: 원격의 기본 브랜치를 알 수 없다
- **WHEN** `--branch` 없이 아직 커밋이 없는 원격을 등록한다
- **THEN** 파일이 바뀌지 않고, 원격의 기본 브랜치를 알 수 없으니 `--branch`를 달라고 알리며 0이 아닌 코드로 끝난다

#### Scenario: 같은 이름이 있다
- **WHEN** 이미 선언된 이름으로 등록한다
- **THEN** 파일이 바뀌지 않고 이름이 겹친다고 알린다

#### Scenario: 사람이 쓴 줄은 그대로다
- **WHEN** `homerepo.yaml`에 사람이 쓴 YAML 주석과 다른 저장소가 있다
- **THEN** 새 저장소만 목록 끝에 더해지고 주석과 다른 줄은 그대로다
