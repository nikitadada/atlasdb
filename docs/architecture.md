# AtlasDB — Architecture

Этот документ описывает **высокоуровневую архитектуру** проекта AtlasDB и ключевые принципы, заложенные при разработке Kubernetes DBaaS-оператора.

Проект ориентирован на **platform / infra / DBaaS** команды и следует best practices Kubernetes Operators.

---

## 🧩 Высокоуровневая схема

```text
User / Platform API
        |
        v
PostgresCluster (CRD)
        |
        v
AtlasDB Operator (controller-runtime)
        |
        v
Kubernetes Primitives
(StatefulSet, Service, PVC)
