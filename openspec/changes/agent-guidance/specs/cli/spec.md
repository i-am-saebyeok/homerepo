# Spec Delta

## Purpose

명령의 사용법을 사람과 에이전트가 소스를 읽지 않고 알 수 있게, 명령마다의 도움말과 잘못 부른 명령의 알림을 정한다.

## ADDED Requirements

### Requirement: 명령의 도움말
모든 명령은 `--help`나 `-h`를 받으면 그 명령의 사용법과 각 옵션의 설명을 표준 출력에 내고 0으로 끝나야 한다(SHALL). 하위 명령 없이 부른 `homerepo --help`, `homerepo repo --help` 같은 명령은 그 아래 명령들의 사용법을 낸다. 인자가 틀렸을 때는 사용법을 표준 오류에 내고 2로 끝난다.

#### Scenario: 명령의 도움말을 본다
- **WHEN** `homerepo thread create --help`를 실행한다
- **THEN** `thread create`의 사용법과 `--repo`, `--mode`, `--runtime`, `--session`의 설명이 표준 출력에 나오고 0으로 끝난다

#### Scenario: 묶음 명령의 도움말을 본다
- **WHEN** `homerepo repo -h`를 실행한다
- **THEN** `repo list`, `repo path`, `repo add`의 사용법이 표준 출력에 나오고 0으로 끝난다

#### Scenario: 인자가 틀렸다
- **WHEN** `homerepo thread create`를 이름 없이 실행한다
- **THEN** 무엇이 틀렸는지와 그 명령의 사용법이 표준 오류에 나오고 2로 끝난다
