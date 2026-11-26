# Proposal

## Why

Home 저장소에서 에이전트 세션을 열고 여러 코드 저장소를 다룰 때, 대화를 갈라 병렬로 일하게 하고 각 일에 격리된 Git 작업공간을 주는 절차를 지금은 사람이 손으로 한다. 이 절차를 명령 하나로 만들어 한 Home에서 실제로 써 보며 구조가 맞는지 확인한다.

## What Changes

- `homerepo` 명령을 새로 만든다. Go로 쓴 실행 파일 하나다.
- 스레드를 만든다: 지금 세션의 대화를 포크해 터미널 멀티플렉서의 새 탭에 띄운다.
- 작업공간을 만들고 지운다: 스레드가 필요한 코드 저장소의 worktree를 그 저장소의 clone 안에 만든다.
- 저장소를 선언하고 찾는다: 공유 설정에 이름과 원격을 적고, 로컬 설정으로 clone의 자리를 바꾼다.
- 에이전트가 이 명령을 부르도록 안내하는 스킬 파일을 Home에 설치한다.

이 변경은 개념 검증의 범위다. 다음은 넣지 않는다.

- 다른 벤더의 런타임으로 넘기는 스레드
- Herdr가 아닌 멀티플렉서
- 패키지 배포

## Capabilities

### New Capabilities

- `repository`: 코드 저장소의 선언, clone의 자리 찾기, 로컬 설정
- `thread`: 세션을 포크해 새 탭에 띄우는 스레드의 생성, 목록, 닫기
- `workspace`: 스레드가 쓰는 worktree의 생성, 목록, 해제

### Modified Capabilities

없음.

## Impact

- 새 Go 모듈 `github.com/i-am-saebyeok/homerepo`
- 실행할 때 필요한 프로그램: `git`, Herdr, 그리고 Claude Code나 Codex
- 쓰는 Home에 생기는 파일: `homerepo.yaml`, Git에서 제외되는 `.homerepo/`, 스킬 폴더 하나
- 각 코드 저장소의 clone에 생기는 폴더: Git에서 제외되는 `.worktrees/`
