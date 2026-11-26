# Tasks

## 1. 뼈대

- [x] 1.1 Go 모듈과 명령의 진입점을 만든다(`homerepo`, 하위 명령의 틀, 오류와 종료 코드의 규칙)
- [x] 1.2 Home 루트를 찾는다(`homerepo.yaml`이 있는 상위 폴더)
- [x] 1.3 `git`과 `herdr`를 부르는 실행기와, 시험에서 바꿔 끼울 수 있는 가짜 실행기를 만든다

## 2. repository

- [x] 2.1 `homerepo.yaml`과 `.homerepo/config.yaml`을 읽고 검증한다
- [x] 2.2 clone의 자리를 계산한다(기본은 Home 옆, 예외는 로컬 설정)
- [x] 2.3 스스로 Git에서 빠지는 폴더를 만드는 함수를 만든다
- [x] 2.4 `repo list`, `repo path`를 만든다
- [x] 2.5 `init`을 만든다(뼈대 파일과 스킬 설치, 있는 파일은 건너뜀)

## 3. workspace

- [x] 3.1 주소로 원격을 찾고 fetch한다
- [x] 3.2 `workspace create`를 만든다(clone이 없으면 clone, 이미 있으면 그대로)
- [x] 3.3 `workspace list`를 만든다(`git worktree list`에서 읽고 밖의 것을 구분)
- [x] 3.4 `workspace release`를 만든다(변경이 있으면 멈춤)
- [x] 3.5 임시 저장소로 위 시나리오를 시험한다

## 4. thread

- [x] 4.1 환경에서 런타임, 세션 ID, Herdr의 창을 읽는다(`CLAUDECODE`, `CODEX_SESSION_ID`, 둘 다 있으면 `--runtime`을 요구)
- [x] 4.2 런타임 어댑터를 만든다(Claude Code와 Codex의 포크 명령, 추가 폴더, 인자의 묶음)
- [x] 4.3 Herdr 어댑터를 만든다(탭 만들기, 창에서 포크 명령 실행, 탭 조회와 닫기)
- [x] 4.4 `thread create`를 만든다(`--repo`, `--mode`, 스레드 기록, 기다리지 않기, 스레드 안에서 거절, 이름 겹침 거절)
- [x] 4.5 `thread list`, `thread close`를 만든다
- [x] 4.6 가짜 실행기로 명령의 조립과 실패 경로를 시험한다

## 5. Home에서 써 보기

- [x] 5.1 실제 Home에 `init`하고 저장소 하나를 선언한다
- [ ] 5.2 Claude Code 세션에서 스레드를 만들어 worktree에서 커밋하고 닫는다
- [ ] 5.3 Codex 세션에서 같은 흐름을 돌린다
- [x] 5.4 부모가 도구를 실행하는 도중에 만든 스레드가 일을 이어 가는지 본다
- [x] 5.5 신뢰 창이 스레드마다 뜨는지 기록한다
- [x] 5.6 써 보며 드러난 것을 사양과 설계에 되돌린다
