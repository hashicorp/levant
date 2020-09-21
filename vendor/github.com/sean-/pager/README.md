# `pager`

Easy `$PAGER` support for Go (i.e. `less(1)`, `more(1)`) with sane defaults.

By default `pager` uses `less(1)` with the args: `-X -F -R --buffers=65535` and
sets `LESSSECURE=1` before starting `less(1)`.

If `less(1)` is unavailable, `pager` falls back to `more(1)`.

The `PAGER` environment variable is honored.

## Example Usage

```go
import (
  "fmt"

  "github.com/sean-/pager"
)

func main() {
  p, err := pager.New()
  if err != nil {
    panic(fmt.Sprintf("unable to get pager: %v", err))
  }
  defer p.Wait()

  foo(p)
}

func foo(w io.Writer) {
  fmt.Fprintf(w, "header\n")
  for i := 0; i < 1000; i++ {
    fmt.Fprintf(w, "line %03d\n", i)
  }
  fmt.Fprintf(w, "trailer\n")
}
```

# Credit

Much of this was pulled from
https://gist.github.com/dchapes/1d0c538ce07902b76c75 and reworked slightly.
