# Python → Go Cheat Sheet

## Types

| Python | Go |
|--------|-----|
| `int` | `int` (int64 on 64-bit) |
| `str` | `string` |
| `bool` | `bool` |
| `float` | `float64` |
| `list[T]` | `[]T` (slice) |
| `dict[K,V]` | `map[K]V` |
| `tuple` | No direct equivalent (struct or multiple return) |
| `set` | No built-in (use `map[T]struct{}` or wait for stdlib) |
| `None` | `nil` (only for pointers, slices, maps, interfaces, funcs, channels) |
| `dataclass` | `struct` with exported fields |
| `Optional[T]` | `*T` (pointer, nil means absent) or `(T, bool)` |
| `Exception` | `error` interface (returned, not raised) |

## Variable declaration

```python
# Python
x = 42
name: str = "Alice"
```

```go
// Go — short declaration (inside functions only)
x := 42
name := "Alice"

// Go — explicit declaration (any scope)
var x int = 42
var name string = "Alice"
```

## Functions

```python
# Python
def add(a: int, b: int) -> int:
    return a + b
```

```go
// Go
func add(a, b int) int {
    return a + b
}
```

## Multiple return (Go's error pattern)

```python
# Python
def get_user(id: int) -> User | None:
    ...
```

```go
// Go — return value + error
func getUser(id int) (User, error) {
    ...
}
```

## Methods

```python
# Python
class Counter:
    value: int = 0
    def increment(self) -> None:
        self.value += 1
```

```go
// Go — method receiver before func name
type Counter struct {
    value int
}

func (c *Counter) Increment() {
    c.value++
}
```

## Slices (dynamic arrays)

```python
# Python
items = ["a", "b", "c"]
items.append("d")
first = items[0]
subset = items[1:3]
```

```go
// Go
items := []string{"a", "b", "c"}
items = append(items, "d")
first := items[0]
subset := items[1:3]
```

## Maps

```python
# Python
scores = {"alice": 10, "bob": 20}
scores["charlie"] = 30
val = scores.get("dave", 0)      # default
val, ok = scores["dave"]         # check existence
```

```go
// Go
scores := map[string]int{"alice": 10, "bob": 20}
scores["charlie"] = 30
val, ok := scores["dave"]        // ok is false if key missing
```

## Error handling

```python
# Python
try:
    result = do_something()
except ValueError as e:
    print(f"error: {e}")
```

```go
// Go
result, err := doSomething()
if err != nil {
    fmt.Printf("error: %v\n", err)
}
```

## Concurrency

```python
# Python (asyncio)
import asyncio
async def fetch(url: str) -> str: ...
await asyncio.gather(fetch(a), fetch(b))
```

```go
// Go (goroutines + channels)
func fetch(url string) string { ... }

// Run concurrently, collect results via channel
ch := make(chan string, 2)
go func() { ch <- fetch(a) }()
go func() { ch <- fetch(b) }()
x, y := <-ch, <-ch
```
