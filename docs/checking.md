## Проверка состояния PostgresCluster

Получить список кластеров:

```bash
kubectl get postgrescluster
```

Посмотреть подробную информацию по конкретному кластеру:

```bash
kubectl describe postgrescluster pg-test
```

### Ожидаемое состояние

- ✔ `phase` **не находится в Reconciling бесконечно**
- ✔ в `events` **нет ошибок**
- ✔ `finalizer` установлен корректно
- ✔ `status` **обновляется**

---

## Проверка DNS внутри Pod

Запускаем временный pod с сетевыми утилитами:

```bash
kubectl run dns-debug \
  --image=busybox:1.35 \
  -it --rm --restart=Never -- sh
```

После запуска можно выполнять DNS-проверки внутри pod (например `nslookup`, `ping` и т.д.).

Проверка готовности postgres
```bash
kubectl exec -it pg-test-0 -- pg_isready
```

Подключение к postgres
```bash
kubectl run psql \
  --image=postgres:15 \
  -it --rm --restart=Never \
  --env="PGPASSWORD=postgres" -- \
  psql -h pg-test-rw -U postgres
```