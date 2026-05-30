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
