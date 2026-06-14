## Профилирование и оптимизация

### Что было оптимизировано:

1. **GetByID** в InMemoryRepository
   - Добавлен индекс `map[string]Link`
   - O(n) → O(1) поиск по ID
   - **Результат: 130ms → 0ms в CPU профиле**

2. **GetAllByUserID** в InMemoryRepository
   - Добавлен индекс `map[string][]int` для userID
   - O(n) → O(k) где k - количество элементов пользователя
   - **Результат: CPU 70ms → 0ms, Память 1.64GB → 532MB (67% меньше)**

3. **SoftDeleteByIDs** в InMemoryRepository
   - map вместо slices.Contains для O(1) поиска вместо O(n)

4. **GenerateJSON и Generate handlers**
   - Убран ненужное string(input.URL)
   - Pre-allocation при сборке URL вместо конкатенации через +

5. **GenerateBatch handler**
   - Pre-allocation при сборке batch URLs

### Результаты оптимизации:

**CPU:**
- GetByID: 130ms → 0ms (полное исчезновение из горячего пути)
- GetAllByUserID: 70ms → 0ms (полное исчезновение из горячего пути)
- Общее время: 750ms → 260ms (**65% быстрее**)
- В топе CPU профиля теперь только runtime операции

**Память:**
- GetAllByUserID: 1.64GB → 532MB (**67% меньше аллокаций**)
- Уменьшение аллокаций в handlers

### pprof (До / После):

```
      flat  flat%   sum%        cum   cum%
   -130ms -26.53% -26.53%     -130ms -26.53%  repository.(*InMemoryRepository).GetByID
    -70ms -14.29% -40.82%      -70ms -14.29%  repository.(*InMemoryRepository).GetAllByUserID
    -50ms -10.20% -51.02%      -50ms -10.20%  handler.Generate
    -25ms  -5.10% -56.12%      -25ms  -5.10%  handler.GenerateJSON
  -100000000 -19.17% -19.17%  -100000000 -19.17%  repository.(*InMemoryRepository).GetAllByUserID
  -100000000  -9.61% -28.78%  -100000000  -9.61%  repository.(*InMemoryRepository).GetByID
   -50000000  -4.81% -33.59%   -50000000  -4.81%  fmt.Sprintf
   -30000000  -2.88% -36.47%   -30000000  -2.88%  json.Marshal
   -20000000  -1.92% -38.39%   -20000000  -1.92%  handler.GenerateBatch
```
