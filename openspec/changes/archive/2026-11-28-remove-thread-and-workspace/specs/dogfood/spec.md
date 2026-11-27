## REMOVED Requirements

### Requirement: 시나리오 실행
**Reason**: 시나리오가 확인하던 스레드와 worktree가 없어지고, 남는 명령은 `go test`가 확인한다.
**Migration**: 없음. 남는 명령의 확인은 `go test -count=1 ./...`로 한다.

### Requirement: 스레드를 시작시킨다
**Reason**: 시나리오가 확인하던 스레드와 worktree가 없어지고, 남는 명령은 `go test`가 확인한다.
**Migration**: 없음. 남는 명령의 확인은 `go test -count=1 ./...`로 한다.

### Requirement: 결과 확인
**Reason**: 시나리오가 확인하던 스레드와 worktree가 없어지고, 남는 명령은 `go test`가 확인한다.
**Migration**: 없음. 남는 명령의 확인은 `go test -count=1 ./...`로 한다.

### Requirement: 시험용 Home의 구성
**Reason**: 시나리오가 확인하던 스레드와 worktree가 없어지고, 남는 명령은 `go test`가 확인한다.
**Migration**: 없음. 남는 명령의 확인은 `go test -count=1 ./...`로 한다.
