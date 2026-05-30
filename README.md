## Профилирование и оптимизация памяти

### Что было оптимизировано:

1. **GetAllByUserID** в InMemoryRepository
   - Pre-allocation срезов для снижения перераспределений памяти

2. **overall memory usage**
   - Сокращение аллокаций в hot path

### Результаты оптимизации:

Снижение потребления памяти:
- ~1MB в encoding/hex.EncodeToString
- ~600KB в main
- ~500KB в CreateMany
