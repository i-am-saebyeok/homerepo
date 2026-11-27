# Tasks

## 1. 스레드와 worktree 지우기

- [x] 1.1 `internal/cli`에서 `thread`와 `workspace` 명령, 그 시험과 시험 도구를 지우고, 명령 목록에 `init`과 `repo` 셋만 남는지 `homerepo --help`의 시험으로 확인한다
- [x] 1.2 원격의 기본 브랜치를 읽는 함수를 `internal/cli/repo.go`로 옮기고, `repo add`의 시험이 그대로 통과하는지 확인한다
- [x] 1.3 `internal/thread`, `internal/workspace`, `internal/herdr`, `internal/dogfood`, `dogfood/`를 지운다
- [x] 1.4 설정에서 `runtimes`를, `internal/home`에서 `.worktrees`의 상수와 경로 함수를, `App`에서 쓰이지 않는 필드를 지우고, `runtimes`가 있는 설정을 거절하는 시험을 더한다
- [x] 1.5 `go mod tidy`를 실행한다

## 2. 스킬과 도움말

- [x] 2.1 설치하는 스킬의 이름을 `homerepo`로 바꾸고 내용을 저장소 명령만 남기며, `init`의 시험이 새 이름과 내용을 확인하게 한다
- [x] 2.2 `init`과 `repo add`의 도움말에서 worktree를 가리키는 글을 고친다

## 3. 확인

- [x] 3.1 `go build ./...`, `go vet ./...`, `gofmt -l .`(빈 출력), `go test -count=1 ./...`가 통과하는지 확인한다
- [x] 3.2 `openspec validate remove-thread-and-workspace --strict`가 통과하는지 확인한다

## 4. 보관

- [ ] 4.1 변경을 보관한 뒤 `openspec/specs`의 `thread`, `workspace`, `dogfood` 폴더를 지우고 `openspec validate --all --strict`가 통과하는지 확인한다
