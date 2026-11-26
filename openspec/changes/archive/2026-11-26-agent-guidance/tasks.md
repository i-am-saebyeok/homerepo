# Tasks

## 1. 구현

- [x] 1.1 `homerepo repo add`를 구현하고(목록 끝에 덧붙이기, 빈 목록 바꾸기, 원격의 기본 브랜치 읽기, 같은 이름 거절, 검사 실패 시 원래 파일 유지), 임시 Home과 로컬 bare 원격으로 시나리오마다 시험이 통과하는지 확인한다
- [x] 1.2 모든 명령의 `--help`와 `-h`가 그 명령의 사용법을 표준 출력에 내고 0으로 끝나며, 틀린 인자는 표준 오류와 2로 끝나는지 시험으로 확인한다
- [x] 1.3 `thread create --repo`가 만든 worktree마다 저장소와 경로를 한 줄씩 출력하고, `--repo`가 없으면 한 줄뿐인지 시험으로 확인한다
- [x] 1.4 선언한 브랜치가 원격에 없을 때 `workspace create`가 아무것도 만들지 않고 첫 커밋이 필요하다고 알리는지, 로컬 빈 bare 원격으로 시험한다
- [x] 1.5 내장 스킬(`internal/home/skill/SKILL.md`)에 `repo add`, `thread create`가 출력하는 worktree 경로, 스레드 안에서 `workspace create`를 다시 불러도 같은 경로가 나온다는 것을 적고, 설명에 "쓰레드"와 "thread"로 부르는 말을 더한다
- [x] 1.6 `go build ./...`, `go vet ./...`, `gofmt -l .`(빈 출력), `go test -count=1 ./...`, `openspec validate agent-guidance`가 통과하는지 확인한다

## 2. 실제로 돌려 보기

- [x] 2.1 시나리오 시험 `new-repository`를 Claude Code로 돌려 네 확인이 통과하고, 에이전트의 결과 문서에 등록 형식과 worktree 경로로 헤맨 기록이 없는지 본다
