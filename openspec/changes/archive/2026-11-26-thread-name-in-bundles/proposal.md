# Proposal

## Why

스레드를 띄울 때 런타임의 원격 제어 같은 기능을 켜려면 인자에 스레드의 이름이 들어가야 한다(예: Claude Code의 `--remote-control <이름>`). 지금 인자 묶음은 고정된 글이라 스레드마다 다른 값을 넣을 수 없다.

## What Changes

- `.homerepo/config.yaml`의 `runtimes.<런타임>.thread.<묶음>`에 적은 인자 안의 `{thread}`를 스레드의 이름으로 바꿔 포크 명령에 붙인다.
- 한 인자 안에서 다른 글과 섞여 있거나 여러 번 나와도 모두 바꾼다.
- 자리표시는 `{thread}` 하나뿐이다. 다른 중괄호 글은 그대로 넘긴다.
- `{thread}`가 없는 묶음은 지금과 똑같이 동작한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `thread`: 요구사항 "실행 인자의 묶음"에 `{thread}` 자리표시를 더한다.

## Impact

- `internal/thread`의 인자 묶음 처리
- 설정 파일의 모양은 바뀌지 않는다. 기존 묶음은 그대로 동작한다.
