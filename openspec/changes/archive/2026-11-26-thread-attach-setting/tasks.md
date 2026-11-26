# Tasks

## 1. 설정

- [x] 1.1 `runtimes.<런타임>.attach`를 읽어 `false`면 포크 명령에 추가 폴더 인자를 붙이지 않게 하고, 가짜 실행기로 `pane run`의 인자를 비교하는 시험이 시나리오마다(설정 없음, 붙이지 않음, 런타임마다 따로, 묶음으로 직접, 참도 거짓도 아닌 값, 훅이 받는 경로) 통과하는지 `go test ./...`로 확인한다
- [x] 1.2 `go build ./...`, `go vet ./...`, `gofmt -l .`(빈 출력), `openspec validate thread-attach-setting`이 통과하는지 확인한다

## 2. Home에서 써 보기

- [x] 2.1 실제 Home에서 `attach`를 끄고 런타임의 설정으로 폴더를 붙여 스레드를 만든다(포크된 세션이 확인 창 없이 뜨고 worktree의 파일을 고칠 수 있는지 본다)
