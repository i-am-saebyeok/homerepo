# Tasks

## 1. 훅

- [x] 1.1 `thread create`가 `.homerepo/hooks/pre-thread-create`를 검사 뒤와 worktree·탭 생성 앞에서, `.homerepo/hooks/post-thread-create`를 포크 명령 실행 뒤에서 실행하게 하고, 임시 Home에 둔 훅 파일과 가짜 실행기로 시나리오마다(훅 없음, 앞의 훅 성공·실패, 거절되는 요청, 뒤의 훅이 받는 값·실패, 앞의 훅에 없는 값, 추가 폴더 여럿, 훅의 출력, 실행할 수 없는 훅) 시험이 통과하는지 `go test ./...`로 확인한다
- [x] 1.2 `go build ./...`, `go vet ./...`, `gofmt -l .`(빈 출력), `openspec validate thread-create-hooks`가 통과하는지 확인한다

## 2. Home에서 써 보기

- [x] 2.1 실제 Home에 두 훅을 두고 스레드를 만든다(앞의 훅이 스레드보다 먼저 돌고, 뒤의 훅이 받은 창의 ID로 그 창을 읽을 수 있으며, 앞의 훅이 실패하면 탭이 생기지 않는지 본다)
