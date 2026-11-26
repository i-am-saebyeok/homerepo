# Spec Delta

## MODIFIED Requirements

### Requirement: 실행 인자의 묶음
시스템은 `.homerepo/config.yaml`의 `runtimes.<런타임>.thread.<묶음>`에 적힌 인자를 포크 명령에 덧붙여야 한다(SHALL). `--mode`가 없으면 `default` 묶음을 쓰고, `default`가 없으면 인자를 붙이지 않는다. 각 인자 안의 `{thread}`는 모두 스레드의 이름으로 바꿔야 한다(SHALL). 다른 중괄호 글은 바꾸지 않으며, 바꾼 인자도 하나의 인자로 넘긴다.

#### Scenario: 묶음을 고른다
- **WHEN** `--mode yolo`를 주고, 그 런타임의 `thread.yolo`에 인자가 적혀 있다
- **THEN** 포크된 세션이 그 인자와 함께 뜬다

#### Scenario: 없는 묶음
- **WHEN** 설정에 없는 묶음 이름을 준다
- **THEN** 아무것도 만들지 않고 있는 묶음 이름들을 알린다

#### Scenario: 인자에 스레드 이름을 넣는다
- **WHEN** `default` 묶음이 `["--remote-control", "{thread}"]`이고 스레드 `login-fix`를 만든다
- **THEN** 포크 명령에 `--remote-control login-fix`가 붙는다

#### Scenario: 다른 글과 섞인 자리표시
- **WHEN** 묶음의 한 인자가 `name={thread}`이거나 `{thread}`가 두 번 나오는 `{thread}-{thread}`이다
- **THEN** 각각 `name=login-fix`와 `login-fix-login-fix` 하나의 인자로 넘겨진다

#### Scenario: 다른 중괄호 글
- **WHEN** 묶음의 인자에 `{repo}`처럼 `{thread}`가 아닌 중괄호 글이 있다
- **THEN** 그 글은 바뀌지 않고 그대로 넘겨진다

#### Scenario: 자리표시가 없다
- **WHEN** 묶음에 `{thread}`가 없거나 묶음이 없다
- **THEN** 포크 명령의 인자는 이 변경 전과 같다
