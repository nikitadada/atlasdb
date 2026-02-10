# AtlasDB — High-Level Architecture

## Цель проекта

AtlasDB — production-like DBaaS-платформа, реализованная как Kubernetes-native система. Проект демонстрирует, как строится инфраструктурная платформа уровня bigtech:

- управление базами данных через Kubernetes API
- строгий lifecycle (create / update / delete)
- автоматизация, observability, UI
- масштабирование и эксплуатация на кластерах

---

## Поддерживаемые базы данных (план)

- PostgreSQL
- Redis
- MongoDB
- ClickHouse

Каждая БД реализуется как отдельный Kubernetes Operator, но подчиняется **единому архитектурному контракту**.

---

## Архитектурные принципы

### 1. Kubernetes-native

- CRD — единственный публичный API
- reconciliation loop как источник истины
- status.conditions как контракт с пользователем

### 2. Declarative over imperative

Пользователь описывает желаемое состояние:

- количество реплик
- storage
- version

Система приводит кластер в это состояние.

### 3. Separation of concerns

- Control plane ≠ data plane
- Operators ≠ UI
- Reconciliation ≠ provisioning logic

---

## Компоненты системы

### 1. Kubernetes Cluster

Базовая среда исполнения.

- Managed (GKE / EKS / Yandex Cloud)
- Self-hosted (kubeadm)

Минимум:
- API Server
- etcd
- CoreDNS
- CSI driver

---

### 2. Operators (Control Plane)

Каждая БД имеет свой оператор:

- postgres-operator
- redis-operator
- mongo-operator
- clickhouse-operator

Общее:

- controller-runtime
- CRD + Reconciler
- finalizers
- conditions

Операторы:

- создают StatefulSet
- создают Service
- управляют PVC
- обновляют status

Операторы **не занимаются**:

- backup storage
- UI
- auth

---

### 3. Custom Resources (API)

Пример:

PostgresCluster

- spec
  - replicas
  - version
  - storage
  - resources
- status
  - conditions
  - observedGeneration

CR — основной интерфейс для:

- UI
- CI/CD
- пользователей

---

### 4. Data Plane (Stateful Workloads)

Каждая БД разворачивается как:

- StatefulSet
- Headless Service
- PersistentVolumeClaims

Гарантии:

- стабильные DNS имена
- сохранность данных
- контролируемые апдейты

---

### 5. Storage Layer

Используется Kubernetes CSI.

Возможные реализации:

- local-pv (dev)
- network storage (prod)
- cloud disks

Каждая БД:

- использует volumeClaimTemplates
- управляет lifecycle storage через operator

---

### 6. Networking

- Headless Service для pod-addressing
- (позже) ClusterIP / LoadBalancer для client access

DNS:

- pod-name.service.namespace.svc.cluster.local

---

### 7. Observability

#### Metrics

- controller-runtime metrics
- custom metrics per DB

Экспорт:

- Prometheus
- Grafana

#### Events

- Kubernetes Events
- Conditions как high-level signals

---

### 8. UI (Control Plane UI)

UI — **отдельный компонент**, не часть операторов.

Функции:

- список кластеров БД
- статус (Ready / Failed)
- создание / удаление
- просмотр spec

Реализация:

- Backend: Go
  - читает CR через Kubernetes API
- Frontend: React

UI **не управляет напрямую** StatefulSet.

---

### 9. Auth & RBAC

- Kubernetes RBAC — источник прав
- UI работает через ServiceAccount

Пользовательские роли:

- viewer
- editor
- admin

---

### 10. CI/CD

- Git — source of truth
- manifests / helm
- kubectl / ArgoCD

Изменения:

- CRD → platform evolution
- CR → runtime operations

---

## Масштабирование

### Horizontal

- увеличение replicas в CR
- оператор приводит StatefulSet в нужное состояние

### Control Plane

- операторы stateless
- можно масштабировать replicas операторов

### UI

- stateless backend
- horizontal scaling

---

## Развёртывание по средам

### Dev

- kind / k3d
- local-pv
- один оператор

### Staging

- managed Kubernetes
- network storage
- Prometheus

### Production

- multi-node cluster
- backups
- alerting

---

## Поток данных (упрощённо)

User → UI → Kubernetes API → CR → Operator → StatefulSet → Pod → Storage

Обратный поток:

Pod / StatefulSet → Operator → Status → Kubernetes API → UI

---

## Эволюция проекта

Этапы:

1. PostgreSQL operator (core)
2. Storage lifecycle
3. Conditions & health
4. Redis / Mongo
5. ClickHouse
6. UI
7. Observability
8. Backups

