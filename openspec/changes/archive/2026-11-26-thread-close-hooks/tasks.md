# Tasks

## 1. 훅

- [x] 1.1 `thread close`가 `.homerepo/hooks/pre-thread-close`를 검사 뒤와 worktree 해제 앞에서, `.homerepo/hooks/post-thread-close`를 탭을 닫은 뒤에서 실행하게 하고, 임시 Home에 둔 훅 파일과 가짜 실행기로 시나리오마다(훅 없음, 닫기 전의 훅 성공·실패, 닫을 수 없는 스레드, 닫은 뒤의 훅 실패, 훅이 받는 값, 탭이 이미 닫힘, 훅이 세션을 먼저 끝냄) 시험이 통과하는지 `go test ./...`로 확인한다
- [x] 1.2 `go build ./...`, `go vet ./...`, `gofmt -l .`(빈 출력), `openspec validate thread-close-hooks`가 통과하는지 확인한다

## 2. Home에서 써 보기

- [x] 2.1 실제 Home에 두 훅을 두고 스레드를 닫는다(닫기 전의 훅이 받은 창에서 세션을 끝낼 수 있고, 닫기 전의 훅이 실패하면 탭과 worktree가 남는지 본다)
