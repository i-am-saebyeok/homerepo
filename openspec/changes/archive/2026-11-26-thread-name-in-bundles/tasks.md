# Tasks

## 1. 자리표시

- [x] 1.1 묶음의 각 인자에서 `{thread}`를 스레드 이름으로 바꿔 포크 명령에 붙이고, 가짜 실행기로 `pane run`의 인자를 비교하는 시험이 시나리오마다 통과하는지 `go test ./...`로 확인한다
- [x] 1.2 `go build ./...`, `go vet ./...`, `gofmt -l .`(빈 출력), `openspec validate thread-name-in-bundles`가 통과하는지 확인한다

## 2. Home에서 써 보기

- [x] 2.1 실제 Home에서 원격 제어가 켜진 스레드를 만든다(묶음에 `--remote-control {thread}`를 두고 만든 스레드가 그 이름으로 원격 제어에 나타나는지 본다)
