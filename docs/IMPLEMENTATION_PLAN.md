# Multi-Git 구현 계획서

## 1. 프로젝트 초기 설정

### 1.1 Go 모듈 초기화
1. `go mod init github.com/lotto/multi-git` 실행
2. 필수 의존성 추가
   - `github.com/spf13/cobra` (CLI 프레임워크)
   - `github.com/go-git/go-git/v5` (Git 라이브러리)
   - `gopkg.in/yaml.v3` (YAML 파싱)
3. 디렉토리 구조 생성
   - `cmd/multi-git/`
   - `internal/commands/`
   - `internal/config/`
   - `internal/repository/`
   - `internal/git/`

### 1.2 CLI 진입점 구성
1. `cmd/multi-git/main.go` 생성
2. Cobra root command 설정
3. 서브 명령어 등록 (clone, checkout, tag, push)
4. 글로벌 플래그 설정 (`--config`, `--verbose`)

---

## 2. Config Manager 구현

### 2.1 설정 구조체 정의
1. `Config` 구조체 정의 (base_dir, default_remote, parallel_workers)
2. `Repository` 구조체 정의 (name, url, path)
3. YAML 태그 매핑

### 2.2 설정 파일 로더 구현
1. 기본 경로 설정 (`~/.multi-git/config.yaml`)
2. 파일 존재 여부 확인
3. YAML 파일 읽기 및 파싱
4. 환경 변수 확장 처리 (`~` → 홈 디렉토리)

### 2.3 설정 검증기 구현
1. 필수 필드 검증 (repositories 목록)
2. URL 형식 검증 (HTTPS/SSH)
3. 중복 저장소 이름 확인
4. 경로 충돌 확인

---

## 3. Repository Manager 구현

### 3.1 저장소 관리자 구조체
1. `Manager` 구조체 정의
2. 설정 데이터 보관
3. 저장소 목록 관리

### 3.2 저장소 작업 실행기
1. 각 저장소에 대해 작업 실행하는 메서드
2. 병렬/순차 실행 옵션
3. 작업 결과 수집

### 3.3 결과 집계 및 리포트
1. 성공/실패 개수 집계
2. 실패한 저장소 목록 생성
3. 총 소요 시간 계산
4. 콘솔 출력 포맷팅

---

## 4. Git Operations 구현

### 4.1 Git 클라이언트 래퍼
1. go-git 라이브러리 래핑
2. 공통 옵션 처리 (인증, 타임아웃)
3. 에러 변환 및 래핑

### 4.2 Clone 작업
1. 저장소 URL 파싱
2. 대상 디렉토리 생성
3. `git clone` 실행
4. 클론 옵션 처리 (depth, branch)

### 4.3 Checkout 작업
1. 저장소 열기
2. 현재 브랜치 확인
3. 원격 브랜치 fetch (옵션)
4. 브랜치 체크아웃 실행
5. 브랜치 생성 옵션 처리

### 4.4 Tag 작업
1. 저장소 열기
2. 태그 존재 여부 확인
3. 태그 생성 (annotated/lightweight)
4. 태그 푸시
5. 태그 삭제

### 4.5 Push 작업
1. 저장소 열기
2. 현재 브랜치 확인
3. 원격 연결 확인
4. Push 실행 (일반/강제)

---

## 5. Clone 명령어 구현

### 5.1 명령어 정의
1. Cobra command 생성 (`clone`)
2. 플래그 정의 (`--skip-existing`, `--parallel`, `--depth`)
3. 사용법 및 도움말 작성

### 5.2 실행 로직
1. 설정 파일 로드
2. 저장소 목록 순회
3. 각 저장소에 대해:
   - 대상 디렉토리 경로 계산
   - 디렉토리 존재 여부 확인
   - 존재하면 스킵 또는 에러
   - 존재하지 않으면 클론 실행
4. 결과 집계 및 출력

### 5.3 에러 처리
1. 네트워크 오류 → 재시도
2. 인증 오류 → 명확한 메시지
3. 디렉토리 존재 → 스킵 또는 에러
4. 부분 실패 허용

---

## 6. Checkout 명령어 구현

### 6.1 명령어 정의
1. Cobra command 생성 (`checkout`)
2. 위치 인자 정의 (`branch-name`)
3. 플래그 정의 (`--create`, `--force`, `--fetch`)

### 6.2 실행 로직
1. 설정 파일 로드
2. 브랜치 이름 인자 검증
3. 저장소 목록 순회
4. 각 저장소에 대해:
   - 저장소 디렉토리 존재 확인
   - 현재 브랜치 확인
   - fetch 실행 (옵션)
   - 로컬 변경사항 확인
   - 체크아웃 실행
5. 결과 집계 및 출력

### 6.3 에러 처리
1. 저장소 미존재 → 에러 메시지
2. 브랜치 미존재 → `--create` 안내
3. 로컬 변경사항 → `--force` 안내
4. 이미 해당 브랜치 → 스킵 (성공 처리)

---

## 7. Tag 명령어 구현

### 7.1 명령어 정의
1. Cobra command 생성 (`tag`)
2. 플래그 정의 (`--branch`, `--name`, `--message`, `--push`, `--force`, `--delete`)
3. 필수 플래그 검증

### 7.2 태그 생성 로직
1. 설정 파일 로드
2. 브랜치, 태그 이름 검증
3. 저장소 목록 순회
4. 각 저장소에 대해:
   - 저장소 디렉토리 존재 확인
   - 지정된 브랜치 체크아웃
   - 태그 존재 여부 확인
   - 존재하면 `--force` 확인
   - 태그 생성
   - `--push` 옵션 시 푸시
5. 결과 집계 및 출력

### 7.3 태그 삭제 로직
1. `--delete` 플래그 확인
2. 각 저장소에서 로컬 태그 삭제
3. `--push` 옵션 시 원격 태그 삭제
4. 결과 집계 및 출력

### 7.4 에러 처리
1. 태그 이미 존재 → `--force` 안내
2. 브랜치 미존재 → 에러 메시지
3. 푸시 실패 → 권한/네트워크 구분

---

## 8. Push 명령어 구현

### 8.1 명령어 정의
1. Cobra command 생성 (`push`)
2. 플래그 정의 (`--branch`, `--force`, `--remote`, `--dry-run`, `--yes`)
3. `--force` 필수 검증

### 8.2 안전장치 구현
1. `--force` 플래그 없으면 에러
2. 확인 프롬프트 표시
3. `--yes` 플래그로 스킵 가능
4. `--dry-run` 모드 지원

### 8.3 실행 로직
1. 설정 파일 로드
2. 안전장치 확인
3. 저장소 목록 순회
4. 각 저장소에 대해:
   - 저장소 디렉토리 존재 확인
   - 현재 브랜치 확인
   - 지정된 브랜치가 아니면 체크아웃
   - `--dry-run`이면 시뮬레이션만
   - 강제 푸시 실행
5. 결과 집계 및 출력

### 8.4 에러 처리
1. 브랜치 미존재 → 에러 메시지
2. 권한 없음 → 인증 에러 메시지
3. 네트워크 오류 → 재시도 또는 에러

---

## 9. 구현 순서 (권장)

### Phase 1: 기반 인프라
1. 프로젝트 초기 설정
2. Config Manager 구현
3. Repository Manager 기본 구조
4. Git Operations 기본 구조

### Phase 2: Clone 기능
1. Git Clone 작업 구현
2. Clone 명령어 구현
3. 테스트 및 검증

### Phase 3: Checkout 기능
1. Git Checkout 작업 구현
2. Checkout 명령어 구현
3. 테스트 및 검증

### Phase 4: Tag 기능
1. Git Tag 작업 구현
2. Tag 명령어 구현
3. 테스트 및 검증

### Phase 5: Push 기능
1. Git Push 작업 구현
2. Push 명령어 구현 (안전장치 포함)
3. 테스트 및 검증

### Phase 6: 마무리
1. 에러 메시지 개선
2. 도움말 작성
3. 빌드 스크립트 작성
4. README 작성

