# Spec Delta

## ADDED Requirements

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
