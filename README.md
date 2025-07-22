# Cosmos SDK Validator 인프라 자동화 & 모니터링 구축


## 🚀 프로젝트 개요

AWS 기반 2리전(서울/도쿄) Validator 네트워크 구축 및 인프라 자동화/모니터링/알림 체계 완성

- **노드 구성**  
  - AWS EC2(Ubuntu) 2대에 Cosmos SDK(`simd`) Validator 배포  
  - Voting Power 50:50, 블록타임 500 ms 설정

- **자동화**  
  - Ansible 플레이북으로 `simd init` → 키 생성 → gentx → genesis 배포 → config 수정 → 노드 실행 전 과정 자동화

- **모니터링 & 알림**  
  - Prometheus + Grafana + Cosmos Validator Exporter + Tendermint Metrics 로 주요 지표(블록 주기, Missed Signatures, 동기화 지연) 대시보드 구성  
  - Go 툴로 `CatchingUp` 지연/서명 누락 탐지 → Discord Webhook 알림 전송

---

## 🛠 기술 스택

- **언어 & 애플리케이션**: Go 1.24.2, Cosmos SDK (simd)  
- **자동화 & 인프라**: Ansible, AWS EC2(Ubuntu), Docker  
- **모니터링**: Prometheus, Grafana, Cosmos Validator Exporter, Tendermint Metrics  
- **알림**: Discord Webhook  

---

## 📂 디렉토리 구조
<img width="683" height="617" alt="스크린샷 2025-07-22 오후 3 03 09" src="https://github.com/user-attachments/assets/4101cb76-8a46-4482-ae12-fb6d963ba6dd" />



## 🖥️ 실제 그라파나 대시보드
<img width="2530" height="1236" alt="스크린샷 2025-07-20 오후 4 22 15" src="https://github.com/user-attachments/assets/dc1358d5-2309-4bb5-8f41-9e577091c0c2" />
