+++
title = "Go's io.Reader"
description = "A brief overview of Go's io.Reader interface."
date = 2025-09-11

[author]
name = "Nicholas Kim"
email = "nickdraggy@gmail.com"
+++


Go's io.Reader is defined as:

```go
type Reader interface {
  Read(p []byte) (n int, err error)
}
```

## Next up

Put more content here under the h2 tag.
